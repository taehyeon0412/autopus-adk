package scheduler

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseCron_WildcardMatchesAll(t *testing.T) {
	t.Parallel()
	expr, err := ParseCron("* * * * *")
	require.NoError(t, err)

	// Any time should match.
	assert.True(t, expr.Match(time.Date(2026, 4, 2, 14, 30, 0, 0, time.UTC)))
	assert.True(t, expr.Match(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)))
}

func TestParseCron_SpecificNumber(t *testing.T) {
	t.Parallel()
	expr, err := ParseCron("5 * * * *")
	require.NoError(t, err)

	assert.True(t, expr.Match(time.Date(2026, 4, 2, 10, 5, 0, 0, time.UTC)))
	assert.False(t, expr.Match(time.Date(2026, 4, 2, 10, 6, 0, 0, time.UTC)))
}

func TestParseCron_Range(t *testing.T) {
	t.Parallel()
	expr, err := ParseCron("* 1-5 * * *")
	require.NoError(t, err)

	for h := 1; h <= 5; h++ {
		assert.True(t, expr.Match(time.Date(2026, 4, 2, h, 0, 0, 0, time.UTC)), "hour %d should match", h)
	}
	assert.False(t, expr.Match(time.Date(2026, 4, 2, 0, 0, 0, 0, time.UTC)))
	assert.False(t, expr.Match(time.Date(2026, 4, 2, 6, 0, 0, 0, time.UTC)))
}

func TestParseCron_List(t *testing.T) {
	t.Parallel()
	expr, err := ParseCron("1,3,5 * * * *")
	require.NoError(t, err)

	assert.True(t, expr.Match(time.Date(2026, 4, 2, 10, 1, 0, 0, time.UTC)))
	assert.True(t, expr.Match(time.Date(2026, 4, 2, 10, 3, 0, 0, time.UTC)))
	assert.True(t, expr.Match(time.Date(2026, 4, 2, 10, 5, 0, 0, time.UTC)))
	assert.False(t, expr.Match(time.Date(2026, 4, 2, 10, 2, 0, 0, time.UTC)))
	assert.False(t, expr.Match(time.Date(2026, 4, 2, 10, 4, 0, 0, time.UTC)))
}

func TestParseCron_Step(t *testing.T) {
	t.Parallel()
	expr, err := ParseCron("*/15 * * * *")
	require.NoError(t, err)

	expected := map[int]bool{0: true, 15: true, 30: true, 45: true}
	for m := 0; m < 60; m++ {
		tm := time.Date(2026, 4, 2, 10, m, 0, 0, time.UTC)
		if expected[m] {
			assert.True(t, expr.Match(tm), "minute %d should match */15", m)
		} else {
			assert.False(t, expr.Match(tm), "minute %d should not match */15", m)
		}
	}
}

func TestParseCron_InvalidExpressions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		expr string
	}{
		{"too few fields", "* * *"},
		{"too many fields", "* * * * * *"},
		{"invalid minute", "60 * * * *"},
		{"invalid hour", "* 25 * * *"},
		{"invalid dom", "* * 32 * *"},
		{"invalid month", "* * * 13 *"},
		{"invalid dow", "* * * * 7"},
		{"bad step", "*/0 * * * *"},
		{"bad range", "* 5-3 * * *"},
		{"non-numeric", "abc * * * *"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := ParseCron(tc.expr)
			assert.Error(t, err)
		})
	}
}

func TestCronExpr_Match(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		expr  string
		time  time.Time
		match bool
	}{
		{
			name:  "every Monday at 9:00",
			expr:  "0 9 * * 1",
			time:  time.Date(2026, 4, 6, 9, 0, 0, 0, time.UTC), // Monday
			match: true,
		},
		{
			name:  "every Monday at 9:00 - wrong day",
			expr:  "0 9 * * 1",
			time:  time.Date(2026, 4, 7, 9, 0, 0, 0, time.UTC), // Tuesday
			match: false,
		},
		{
			name:  "first of month",
			expr:  "0 0 1 * *",
			time:  time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
			match: true,
		},
		{
			name:  "range step combo",
			expr:  "0-30/10 * * * *",
			time:  time.Date(2026, 4, 2, 10, 20, 0, 0, time.UTC),
			match: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			expr, err := ParseCron(tc.expr)
			require.NoError(t, err)
			assert.Equal(t, tc.match, expr.Match(tc.time))
		})
	}
}
