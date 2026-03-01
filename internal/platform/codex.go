package platform

import (
	"os"
	"path/filepath"

	"github.com/devrimcavusoglu/skern/internal/skill"
)

// CodexCLI is the platform adapter for Codex CLI.
type CodexCLI struct {
	homeDir     string
	projectRoot string
}

// NewCodexCLI creates a Codex CLI adapter.
// Empty strings use default paths (home directory, current directory).
func NewCodexCLI(homeDir, projectRoot string) *CodexCLI {
	if homeDir == "" {
		homeDir, _ = os.UserHomeDir()
	}
	if projectRoot == "" {
		projectRoot = "."
	}
	return &CodexCLI{homeDir: homeDir, projectRoot: projectRoot}
}

// Name implements Platform.
func (c *CodexCLI) Name() Type { return TypeCodexCLI }

// Detect implements Platform.
func (c *CodexCLI) Detect() bool {
	// Primary: ~/.agents/
	if _, err := os.Stat(filepath.Join(c.homeDir, ".agents")); err == nil {
		return true
	}
	// Fallback: ~/.codex/
	_, err := os.Stat(filepath.Join(c.homeDir, ".codex"))
	return err == nil
}

// UserSkillsDir implements Platform.
func (c *CodexCLI) UserSkillsDir() string {
	return filepath.Join(c.homeDir, ".agents", "skills")
}

// ProjectSkillsDir implements Platform.
func (c *CodexCLI) ProjectSkillsDir() string {
	return filepath.Join(c.projectRoot, ".agents", "skills")
}

// Install implements Platform.
func (c *CodexCLI) Install(skillDir string, skillName string, scope skill.Scope) error {
	return installSkill(skillDir, skillName, c.skillsDir(scope))
}

// Uninstall implements Platform.
func (c *CodexCLI) Uninstall(skillName string, scope skill.Scope) error {
	return uninstallSkill(skillName, c.skillsDir(scope))
}

// InstalledSkills implements Platform.
func (c *CodexCLI) InstalledSkills(scope skill.Scope) ([]string, error) {
	return listInstalledSkills(c.skillsDir(scope))
}

func (c *CodexCLI) skillsDir(scope skill.Scope) string {
	if scope == skill.ScopeProject {
		return c.ProjectSkillsDir()
	}
	return c.UserSkillsDir()
}
