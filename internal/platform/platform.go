// Package platform provides adapters for installing skills to agentic development platforms.
package platform

import (
	"github.com/devrimcavusoglu/skern/internal/skill"
)

// Type identifies a supported platform.
type Type string

// Platform type constants. Adding a new platform also requires an entry in
// Specs (see spec.go).
const (
	TypeClaudeCode    Type = "claude-code"
	TypeCodexCLI      Type = "codex-cli"
	TypeOpenCode      Type = "opencode"
	TypeCursor        Type = "cursor"
	TypeGeminiCLI     Type = "gemini-cli"
	TypeGitHubCopilot Type = "github-copilot"
	TypeWindsurf      Type = "windsurf"
	TypeContinue      Type = "continue"
)

// InstallOptions shapes how a skill is copied onto a platform. The zero
// value copies the whole skill directory.
type InstallOptions struct {
	// Exclude holds the skill's `install.exclude` glob patterns; matching
	// files and directories stay in the registry but are not copied. See
	// skill.MatchExclude for the rules.
	Exclude []string
}

// Platform defines the interface that each platform adapter must implement.
type Platform interface {
	// Name returns the platform type identifier.
	Name() Type
	// Detect returns true if this platform is installed on the system.
	Detect() bool
	// UserSkillsDir returns the absolute path to the user-level skills directory.
	UserSkillsDir() string
	// ProjectSkillsDir returns the absolute path to the project-level skills directory.
	ProjectSkillsDir() string
	// Install copies a skill from the registry into the platform's skills
	// directory, honoring opts.Exclude.
	Install(skillDir string, skillName string, scope skill.Scope, opts InstallOptions) error
	// Uninstall removes a skill from the platform's skills directory.
	Uninstall(skillName string, scope skill.Scope) error
	// InstalledSkills returns the names of skills installed for the given scope.
	InstalledSkills(scope skill.Scope) ([]string, error)
}
