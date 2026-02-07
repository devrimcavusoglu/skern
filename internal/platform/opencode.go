package platform

import (
	"os"
	"path/filepath"

	"github.com/devrimcavusoglu/scribe/internal/skill"
)

// OpenCode is the platform adapter for OpenCode.
type OpenCode struct {
	homeDir     string
	projectRoot string
}

// NewOpenCode creates an OpenCode adapter.
// Empty strings use default paths (home directory, current directory).
func NewOpenCode(homeDir, projectRoot string) *OpenCode {
	if homeDir == "" {
		homeDir, _ = os.UserHomeDir()
	}
	if projectRoot == "" {
		projectRoot = "."
	}
	return &OpenCode{homeDir: homeDir, projectRoot: projectRoot}
}

// Name implements Platform.
func (o *OpenCode) Name() Type { return TypeOpenCode }

// Detect implements Platform.
func (o *OpenCode) Detect() bool {
	_, err := os.Stat(filepath.Join(o.homeDir, ".config", "opencode"))
	return err == nil
}

// UserSkillsDir implements Platform.
func (o *OpenCode) UserSkillsDir() string {
	return filepath.Join(o.homeDir, ".config", "opencode", "skills")
}

// ProjectSkillsDir implements Platform.
func (o *OpenCode) ProjectSkillsDir() string {
	return filepath.Join(o.projectRoot, ".opencode", "skills")
}

// Install implements Platform.
func (o *OpenCode) Install(skillDir string, skillName string, scope skill.Scope) error {
	return installSkill(skillDir, skillName, o.skillsDir(scope))
}

// Uninstall implements Platform.
func (o *OpenCode) Uninstall(skillName string, scope skill.Scope) error {
	return uninstallSkill(skillName, o.skillsDir(scope))
}

// InstalledSkills implements Platform.
func (o *OpenCode) InstalledSkills(scope skill.Scope) ([]string, error) {
	return listInstalledSkills(o.skillsDir(scope))
}

func (o *OpenCode) skillsDir(scope skill.Scope) string {
	if scope == skill.ScopeProject {
		return o.ProjectSkillsDir()
	}
	return o.UserSkillsDir()
}
