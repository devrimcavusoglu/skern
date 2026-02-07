package platform

import (
	"os"
	"path/filepath"

	"github.com/devrimcavusoglu/scribe/internal/skill"
)

// ClaudeCode is the platform adapter for Claude Code.
type ClaudeCode struct {
	homeDir     string
	projectRoot string
}

// NewClaudeCode creates a Claude Code adapter.
// Empty strings use default paths (home directory, current directory).
func NewClaudeCode(homeDir, projectRoot string) *ClaudeCode {
	if homeDir == "" {
		homeDir, _ = os.UserHomeDir()
	}
	if projectRoot == "" {
		projectRoot = "."
	}
	return &ClaudeCode{homeDir: homeDir, projectRoot: projectRoot}
}

// Name implements Platform.
func (c *ClaudeCode) Name() Type { return TypeClaudeCode }

// Detect implements Platform.
func (c *ClaudeCode) Detect() bool {
	_, err := os.Stat(filepath.Join(c.homeDir, ".claude"))
	return err == nil
}

// UserSkillsDir implements Platform.
func (c *ClaudeCode) UserSkillsDir() string {
	return filepath.Join(c.homeDir, ".claude", "skills")
}

// ProjectSkillsDir implements Platform.
func (c *ClaudeCode) ProjectSkillsDir() string {
	return filepath.Join(c.projectRoot, ".claude", "skills")
}

// Install implements Platform.
func (c *ClaudeCode) Install(skillDir string, skillName string, scope skill.Scope) error {
	return installSkill(skillDir, skillName, c.skillsDir(scope))
}

// Uninstall implements Platform.
func (c *ClaudeCode) Uninstall(skillName string, scope skill.Scope) error {
	return uninstallSkill(skillName, c.skillsDir(scope))
}

// InstalledSkills implements Platform.
func (c *ClaudeCode) InstalledSkills(scope skill.Scope) ([]string, error) {
	return listInstalledSkills(c.skillsDir(scope))
}

func (c *ClaudeCode) skillsDir(scope skill.Scope) string {
	if scope == skill.ScopeProject {
		return c.ProjectSkillsDir()
	}
	return c.UserSkillsDir()
}
