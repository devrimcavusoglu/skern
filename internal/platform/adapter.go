package platform

import (
	"os"
	"path/filepath"

	"github.com/devrimcavusoglu/skern/internal/skill"
)

// Adapter is a generic platform adapter parameterized by a Spec. Every
// supported platform is represented by an Adapter — adding a new one means
// appending to Specs, not writing more Go.
type Adapter struct {
	spec        Spec
	homeDir     string
	projectRoot string
}

// New returns a Platform adapter for the registered spec name, or nil if the
// name is not in Specs. Empty homeDir/projectRoot resolve to the OS home and
// the current directory respectively.
func New(name Type, homeDir, projectRoot string) Platform {
	spec := SpecFor(name)
	if spec == nil {
		return nil
	}
	if homeDir == "" {
		homeDir, _ = os.UserHomeDir()
	}
	if projectRoot == "" {
		projectRoot = "."
	}
	return &Adapter{spec: *spec, homeDir: homeDir, projectRoot: projectRoot}
}

// Name implements Platform.
func (a *Adapter) Name() Type { return a.spec.Name }

// Detect implements Platform — true if any of the spec's home-relative paths
// exist on disk.
func (a *Adapter) Detect() bool {
	for _, p := range a.spec.DetectHome {
		if _, err := os.Stat(filepath.Join(a.homeDir, p)); err == nil {
			return true
		}
	}
	return false
}

// UserSkillsDir implements Platform.
func (a *Adapter) UserSkillsDir() string {
	return filepath.Join(a.homeDir, a.spec.UserDir)
}

// ProjectSkillsDir implements Platform.
func (a *Adapter) ProjectSkillsDir() string {
	return filepath.Join(a.projectRoot, a.spec.ProjectDir)
}

// Install implements Platform.
func (a *Adapter) Install(skillDir, skillName string, scope skill.Scope, opts InstallOptions) error {
	return installSkill(skillDir, skillName, a.skillsDir(scope), opts)
}

// Uninstall implements Platform.
func (a *Adapter) Uninstall(skillName string, scope skill.Scope) error {
	return uninstallSkill(skillName, a.skillsDir(scope))
}

// InstalledSkills implements Platform.
func (a *Adapter) InstalledSkills(scope skill.Scope) ([]string, error) {
	return listInstalledSkills(a.skillsDir(scope))
}

func (a *Adapter) skillsDir(scope skill.Scope) string {
	if scope == skill.ScopeProject {
		return a.ProjectSkillsDir()
	}
	return a.UserSkillsDir()
}
