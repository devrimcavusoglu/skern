package instructions

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// CandidateProjectFiles lists project-scope agent instruction files that
// `skern init` will probe when the user has not supplied explicit `--target`
// paths. Order is preserved in the output of DiscoverTargets so callers can
// surface a deterministic list to the user.
//
// User-level files (e.g. ~/.claude/CLAUDE.md) are intentionally NOT in this
// list — they are global and require explicit opt-in via `--target`.
var CandidateProjectFiles = []string{
	"AGENTS.md",
	"CLAUDE.md",
	filepath.Join(".claude", "CLAUDE.md"),
}

// DiscoverTargets returns the subset of CandidateProjectFiles that exist
// (as regular files) under projectRoot. Symlinks are followed via os.Stat.
func DiscoverTargets(projectRoot string) ([]string, error) {
	var found []string
	for _, rel := range CandidateProjectFiles {
		path := filepath.Join(projectRoot, rel)
		info, err := os.Stat(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, fmt.Errorf("checking %s: %w", path, err)
		}
		if info.Mode().IsRegular() {
			found = append(found, path)
		}
	}
	return found, nil
}

// WriteResult records the outcome of a single Write call.
type WriteResult struct {
	Path    string `json:"path"`
	Action  string `json:"action"` // "created", "updated", "unchanged", "appended"
	Created bool   `json:"created"`
}

// blockPattern matches a previously-written skern block from start to end
// marker (inclusive), with an optional trailing blank line. Multiline mode
// is required because the rendered block spans multiple lines.
var blockPattern = regexp.MustCompile(
	`(?s)` + regexp.QuoteMeta(StartMarker) + `.*?` + regexp.QuoteMeta(EndMarker) + `\n?`,
)

// Write inserts (or updates) the rendered instruction block in `path`.
//
// Behavior:
//   - File missing: created, block written as the file's only content.
//   - File present, no existing block: rendered block appended (with a blank
//     line separator if the file's last byte is not already a newline).
//   - File present, existing block: replaced in place, leaving surrounding
//     content untouched.
//   - File present, existing block matches rendered block byte-for-byte:
//     no write occurs (idempotent). Action is "unchanged".
//
// The rendered block is the value returned by Render(toolForming).
func Write(path, rendered string) (WriteResult, error) {
	res := WriteResult{Path: path}

	existing, err := os.ReadFile(path)
	if err != nil {
		if !os.IsNotExist(err) {
			return res, fmt.Errorf("reading %s: %w", path, err)
		}
		if mkErr := os.MkdirAll(filepath.Dir(path), 0o755); mkErr != nil {
			return res, fmt.Errorf("creating parent of %s: %w", path, mkErr)
		}
		if writeErr := os.WriteFile(path, []byte(rendered), 0o644); writeErr != nil {
			return res, fmt.Errorf("writing %s: %w", path, writeErr)
		}
		res.Action = "created"
		res.Created = true
		return res, nil
	}

	current := string(existing)
	loc := blockPattern.FindStringIndex(current)
	if loc != nil {
		// Existing block — replace.
		if current[loc[0]:loc[1]] == rendered {
			res.Action = "unchanged"
			return res, nil
		}
		updated := current[:loc[0]] + rendered + current[loc[1]:]
		if err := os.WriteFile(path, []byte(updated), 0o644); err != nil {
			return res, fmt.Errorf("updating %s: %w", path, err)
		}
		res.Action = "updated"
		return res, nil
	}

	// No block — append, preserving original contents.
	var b strings.Builder
	b.WriteString(current)
	if !strings.HasSuffix(current, "\n") {
		b.WriteString("\n")
	}
	if !strings.HasSuffix(current, "\n\n") {
		b.WriteString("\n")
	}
	b.WriteString(rendered)
	if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
		return res, fmt.Errorf("appending to %s: %w", path, err)
	}
	res.Action = "appended"
	return res, nil
}
