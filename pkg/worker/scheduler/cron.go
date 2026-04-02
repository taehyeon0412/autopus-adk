package scheduler

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// CronExpr represents a parsed 5-field cron expression.
type CronExpr struct {
	minutes []int // 0-59
	hours   []int // 0-23
	doms    []int // 1-31
	months  []int // 1-12
	dows    []int // 0-6 (0=Sunday)
}

// ParseCron parses a 5-field cron expression (minute hour dom month dow).
// Supports: *, numbers, ranges (1-5), lists (1,3,5), steps (*/5, 1-10/2).
func ParseCron(expr string) (*CronExpr, error) {
	fields := strings.Fields(expr)
	if len(fields) != 5 {
		return nil, fmt.Errorf("cron: expected 5 fields, got %d", len(fields))
	}

	minutes, err := parseField(fields[0], 0, 59)
	if err != nil {
		return nil, fmt.Errorf("cron minute: %w", err)
	}
	hours, err := parseField(fields[1], 0, 23)
	if err != nil {
		return nil, fmt.Errorf("cron hour: %w", err)
	}
	doms, err := parseField(fields[2], 1, 31)
	if err != nil {
		return nil, fmt.Errorf("cron dom: %w", err)
	}
	months, err := parseField(fields[3], 1, 12)
	if err != nil {
		return nil, fmt.Errorf("cron month: %w", err)
	}
	dows, err := parseField(fields[4], 0, 6)
	if err != nil {
		return nil, fmt.Errorf("cron dow: %w", err)
	}

	return &CronExpr{
		minutes: minutes,
		hours:   hours,
		doms:    doms,
		months:  months,
		dows:    dows,
	}, nil
}

// Match returns true if the given time matches this cron expression.
func (c *CronExpr) Match(t time.Time) bool {
	return contains(c.minutes, t.Minute()) &&
		contains(c.hours, t.Hour()) &&
		contains(c.doms, t.Day()) &&
		contains(c.months, int(t.Month())) &&
		contains(c.dows, int(t.Weekday()))
}

func contains(vals []int, v int) bool {
	for _, x := range vals {
		if x == v {
			return true
		}
	}
	return false
}

// parseField parses a single cron field into a sorted list of integers.
func parseField(field string, min, max int) ([]int, error) {
	var result []int
	for _, part := range strings.Split(field, ",") {
		vals, err := parsePart(part, min, max)
		if err != nil {
			return nil, err
		}
		result = append(result, vals...)
	}
	return result, nil
}

func parsePart(part string, min, max int) ([]int, error) {
	// Handle step: */5 or 1-10/2
	step := 1
	if idx := strings.Index(part, "/"); idx >= 0 {
		var err error
		step, err = strconv.Atoi(part[idx+1:])
		if err != nil || step <= 0 {
			return nil, fmt.Errorf("invalid step: %q", part)
		}
		part = part[:idx]
	}

	var lo, hi int
	switch {
	case part == "*":
		lo, hi = min, max
	case strings.Contains(part, "-"):
		r := strings.SplitN(part, "-", 2)
		var err error
		lo, err = strconv.Atoi(r[0])
		if err != nil {
			return nil, fmt.Errorf("invalid range start: %q", part)
		}
		hi, err = strconv.Atoi(r[1])
		if err != nil {
			return nil, fmt.Errorf("invalid range end: %q", part)
		}
	default:
		v, err := strconv.Atoi(part)
		if err != nil {
			return nil, fmt.Errorf("invalid value: %q", part)
		}
		lo, hi = v, v
	}

	if lo < min || hi > max || lo > hi {
		return nil, fmt.Errorf("out of range [%d-%d]: %d-%d", min, max, lo, hi)
	}

	var vals []int
	for i := lo; i <= hi; i += step {
		vals = append(vals, i)
	}
	return vals, nil
}
