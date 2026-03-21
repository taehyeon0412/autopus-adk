package tui_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/insajin/autopus-adk/internal/cli/tui"
)

func TestBanner_ContainsBrandName(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	tui.Banner(&buf)
	out := buf.String()
	assert.Contains(t, out, "Autopus")
	assert.Contains(t, out, "🐙")
}

func TestBannerWithInfo_ContainsProjectAndMode(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	tui.BannerWithInfo(&buf, "myproject", "full")
	out := buf.String()
	assert.Contains(t, out, "myproject")
	assert.Contains(t, out, "full")
}

func TestSuccess_WritesMessage(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	tui.Success(&buf, "done")
	assert.Contains(t, buf.String(), "done")
	assert.Contains(t, buf.String(), "✓")
}

func TestError_WritesMessage(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	tui.Error(&buf, "fail")
	assert.Contains(t, buf.String(), "fail")
	assert.Contains(t, buf.String(), "✗")
}

func TestWarn_WritesMessage(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	tui.Warn(&buf, "caution")
	assert.Contains(t, buf.String(), "caution")
	assert.Contains(t, buf.String(), "⚠")
}

func TestInfo_WritesMessage(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	tui.Info(&buf, "note")
	assert.Contains(t, buf.String(), "note")
	assert.Contains(t, buf.String(), "ℹ")
}

func TestStep_WritesCounter(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	tui.Step(&buf, 2, 5, "loading")
	out := buf.String()
	assert.Contains(t, out, "[2/5]")
	assert.Contains(t, out, "loading")
}

func TestBullet_WritesItem(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	tui.Bullet(&buf, "item")
	assert.Contains(t, buf.String(), "item")
}

func TestOK_WritesLabel(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	tui.OK(&buf, "check passed")
	assert.Contains(t, buf.String(), "OK")
	assert.Contains(t, buf.String(), "check passed")
}

func TestFAIL_WritesLabel(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	tui.FAIL(&buf, "check failed")
	assert.Contains(t, buf.String(), "ERROR")
	assert.Contains(t, buf.String(), "check failed")
}

func TestBox_ContainsTitleAndContent(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	tui.Box(&buf, "Title", "body text")
	out := buf.String()
	assert.Contains(t, out, "Title")
	assert.Contains(t, out, "body text")
	// rounded border chars
	assert.Contains(t, out, "╭")
	assert.Contains(t, out, "╰")
}

func TestResultBox_Pass(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	tui.ResultBox(&buf, true, "all good")
	out := buf.String()
	assert.Contains(t, out, "PASS")
	assert.Contains(t, out, "all good")
}

func TestResultBox_Fail(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	tui.ResultBox(&buf, false, "issues found")
	out := buf.String()
	assert.Contains(t, out, "FAIL")
	assert.Contains(t, out, "issues found")
}

func TestTag_FormatsLabelValue(t *testing.T) {
	t.Parallel()
	result := tui.Tag("mode", "full")
	assert.Contains(t, result, "mode")
	assert.Contains(t, result, "full")
}

func TestSectionHeader_WritesTitle(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	tui.SectionHeader(&buf, "Dependencies")
	assert.Contains(t, buf.String(), "Dependencies")
}

func TestDivider_WritesLine(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	tui.Divider(&buf)
	assert.Contains(t, buf.String(), "─")
}
