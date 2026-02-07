// Package output provides structured output formatting for the scribe CLI.
// All commands route their output through this package to support --json and --quiet flags consistently.
package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// Printer handles formatted output for CLI commands.
type Printer struct {
	out    io.Writer
	errOut io.Writer
	json   bool
	quiet  bool
}

// NewPrinter creates a Printer with the given options.
func NewPrinter(jsonMode, quiet bool) *Printer {
	return &Printer{
		out:    os.Stdout,
		errOut: os.Stderr,
		json:   jsonMode,
		quiet:  quiet,
	}
}

// SetOut sets the output writer for the Printer.
func (p *Printer) SetOut(w io.Writer) {
	p.out = w
}

// SetErrOut sets the error output writer for the Printer.
func (p *Printer) SetErrOut(w io.Writer) {
	p.errOut = w
}

// IsJSON returns whether JSON output mode is enabled.
func (p *Printer) IsJSON() bool {
	return p.json
}

// IsQuiet returns whether quiet mode is enabled.
func (p *Printer) IsQuiet() bool {
	return p.quiet
}

// Print outputs a line of text. In quiet mode, this is suppressed.
// In JSON mode, this is suppressed (use PrintResult for structured data).
func (p *Printer) Print(format string, args ...any) {
	if p.quiet || p.json {
		return
	}
	_, _ = fmt.Fprintf(p.out, format, args...)
}

// Println outputs a line of text with a trailing newline.
func (p *Printer) Println(args ...any) {
	if p.quiet || p.json {
		return
	}
	_, _ = fmt.Fprintln(p.out, args...)
}

// PrintResult outputs structured data. In JSON mode, it serializes to JSON.
// In text mode, it uses the provided text representation.
func (p *Printer) PrintResult(data any, text string) {
	if p.quiet {
		return
	}
	if p.json {
		p.printJSON(data)
		return
	}
	_, _ = fmt.Fprint(p.out, text)
}

// PrintError outputs an error message to stderr. Not suppressed by --quiet.
func (p *Printer) PrintError(format string, args ...any) {
	if p.json {
		return
	}
	_, _ = fmt.Fprintf(p.errOut, "Error: "+format+"\n", args...)
}

// PrintErrorResult outputs an error in JSON format when --json is set,
// or as text to stderr otherwise.
func (p *Printer) PrintErrorResult(err error) {
	if p.json {
		p.printJSON(ErrorResult{Error: err.Error()})
		return
	}
	_, _ = fmt.Fprintf(p.errOut, "Error: %s\n", err)
}

func (p *Printer) printJSON(data any) {
	enc := json.NewEncoder(p.out)
	enc.SetIndent("", "  ")
	_ = enc.Encode(data)
}

// ErrorResult is the JSON envelope for error output.
type ErrorResult struct {
	Error string `json:"error"`
}

// VersionResult is the JSON envelope for version output.
type VersionResult struct {
	Version string `json:"version"`
	Commit  string `json:"commit,omitempty"`
	Date    string `json:"date,omitempty"`
}

// AuthorResult is the JSON representation of a skill author.
type AuthorResult struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Platform string `json:"platform,omitempty"`
}

// SkillResult is the JSON representation of a skill.
type SkillResult struct {
	Name         string       `json:"name"`
	Description  string       `json:"description"`
	Version      string       `json:"version"`
	Author       AuthorResult `json:"author"`
	Scope        string       `json:"scope,omitempty"`
	Path         string       `json:"path,omitempty"`
	AllowedTools []string     `json:"allowed_tools,omitempty"`
}

// SkillCreateResult is the JSON envelope for skill create output.
type SkillCreateResult struct {
	Name  string `json:"name"`
	Scope string `json:"scope"`
	Path  string `json:"path"`
}

// SkillListResult is the JSON envelope for skill list output.
type SkillListResult struct {
	Skills []SkillResult `json:"skills"`
	Count  int           `json:"count"`
}

// SkillSearchResult is the JSON envelope for skill search output.
type SkillSearchResult struct {
	Query   string        `json:"query"`
	Results []SkillResult `json:"results"`
	Count   int           `json:"count"`
}

// SkillRemoveResult is the JSON envelope for skill remove output.
type SkillRemoveResult struct {
	Name  string `json:"name"`
	Scope string `json:"scope"`
}
