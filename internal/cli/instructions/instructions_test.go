package instructions

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRender_BaseAlwaysIncluded(t *testing.T) {
	got := Render(false)
	assert.Contains(t, got, StartMarker)
	assert.Contains(t, got, EndMarker)
	assert.Contains(t, got, "Skern (skill management)")
	assert.Contains(t, got, "ALL skill-related tasks")
}

func TestRender_BaseExcludesToolForming(t *testing.T) {
	got := Render(false)
	assert.NotContains(t, got, "Tool-forming loop")
	assert.NotContains(t, got, "skern skill search")
}

func TestRender_WithToolForming(t *testing.T) {
	got := Render(true)
	assert.Contains(t, got, "Tool-forming loop")
	assert.Contains(t, got, "skern skill search")
	assert.Contains(t, got, "≥ 0.6")
	assert.Contains(t, got, "≥ 0.9")
}

func TestRender_StableMarkers(t *testing.T) {
	for _, tf := range []bool{false, true} {
		got := Render(tf)
		assert.True(t, strings.HasPrefix(got, StartMarker+"\n"),
			"render must start with start marker; toolForming=%v", tf)
		assert.True(t, strings.HasSuffix(got, EndMarker+"\n"),
			"render must end with end marker; toolForming=%v", tf)
	}
}

func TestRender_Idempotent(t *testing.T) {
	a := Render(false)
	b := Render(false)
	assert.Equal(t, a, b)

	c := Render(true)
	d := Render(true)
	assert.Equal(t, c, d)
}
