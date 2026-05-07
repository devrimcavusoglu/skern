// Package instructions assembles and writes the skern usage snippet that
// `skern init` injects into agent instruction files (CLAUDE.md, AGENTS.md,
// etc.). The snippet teaches the agent to use skern for all skill-related
// tasks instead of reading platform-native skill directories directly.
package instructions

import (
	_ "embed"
	"strings"
)

//go:embed base.md
var baseMD string

//go:embed tool_forming.md
var toolFormingMD string

// Sentinel markers wrap the rendered snippet so re-running `skern init` can
// update the block in-place rather than appending duplicates.
const (
	StartMarker = "<!-- skern:instructions:start -->"
	EndMarker   = "<!-- skern:instructions:end -->"
)

// Render returns the instruction snippet wrapped in start/end markers.
// When toolForming is true the opt-in tool-forming-loop section is appended
// to the always-included base.
func Render(toolForming bool) string {
	body := strings.TrimRight(baseMD, "\n")
	if toolForming {
		body += "\n\n" + strings.TrimRight(toolFormingMD, "\n")
	}
	return StartMarker + "\n" + body + "\n" + EndMarker + "\n"
}
