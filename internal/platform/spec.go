package platform

// Spec is the declarative configuration for one platform adapter. Adding a
// new supported platform is a one-line PR — append to Specs below.
//
// Path conventions follow vercel-labs/skills (https://github.com/vercel-labs/skills#supported-agents),
// the closest thing to a community standard for Agent Skills directory layout.
//
// UserDir is rooted at the user's home directory; ProjectDir at the project
// root. DetectHome lists candidate paths under home — if any of them exists,
// the platform is considered installed. Per-platform detection paths matter
// because several platforms share .agents/skills/ as their project dir but
// each carries a distinct user-level config dir (e.g. ~/.cursor, ~/.gemini,
// ~/.copilot) that disambiguates them.
type Spec struct {
	Name       Type
	UserDir    string
	ProjectDir string
	DetectHome []string
}

// Specs is the source of truth for supported platforms. Each entry is mapped
// to a generic Adapter at runtime; no per-platform Go file is required.
//
// The first three entries (claude-code, codex-cli, opencode) keep the paths
// skern shipped originally — the codex-cli UserDir intentionally stays at
// ~/.agents/skills/ rather than vercel's ~/.codex/skills/ so existing skern
// users see no disk-layout change.
var Specs = []Spec{
	{
		Name:       TypeClaudeCode,
		UserDir:    ".claude/skills",
		ProjectDir: ".claude/skills",
		DetectHome: []string{".claude"},
	},
	{
		Name:       TypeCodexCLI,
		UserDir:    ".agents/skills",
		ProjectDir: ".agents/skills",
		DetectHome: []string{".codex", ".agents"},
	},
	{
		Name:       TypeOpenCode,
		UserDir:    ".config/opencode/skills",
		ProjectDir: ".opencode/skills",
		DetectHome: []string{".config/opencode"},
	},
	{
		Name:       TypeCursor,
		UserDir:    ".cursor/skills",
		ProjectDir: ".agents/skills",
		DetectHome: []string{".cursor"},
	},
	{
		Name:       TypeGeminiCLI,
		UserDir:    ".gemini/skills",
		ProjectDir: ".agents/skills",
		DetectHome: []string{".gemini"},
	},
	{
		Name:       TypeGitHubCopilot,
		UserDir:    ".copilot/skills",
		ProjectDir: ".agents/skills",
		DetectHome: []string{".copilot"},
	},
	{
		Name:       TypeWindsurf,
		UserDir:    ".codeium/windsurf/skills",
		ProjectDir: ".windsurf/skills",
		DetectHome: []string{".codeium/windsurf", ".windsurf"},
	},
	{
		Name:       TypeContinue,
		UserDir:    ".continue/skills",
		ProjectDir: ".continue/skills",
		DetectHome: []string{".continue"},
	},
}

// SpecFor returns the Spec registered under name, or nil if unknown.
func SpecFor(name Type) *Spec {
	for i := range Specs {
		if Specs[i].Name == name {
			return &Specs[i]
		}
	}
	return nil
}

// SupportedNames returns the registered platform names in declaration order.
func SupportedNames() []Type {
	out := make([]Type, len(Specs))
	for i, s := range Specs {
		out[i] = s.Name
	}
	return out
}
