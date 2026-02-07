package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/devrimcavusoglu/scribe/internal/output"
	"github.com/devrimcavusoglu/scribe/internal/platform"
	"github.com/devrimcavusoglu/scribe/internal/registry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestDetector overrides newDetectorFunc to use temp directories with all platforms detected.
func setupTestDetector(t *testing.T, home, project string) {
	t.Helper()

	// Create platform directories so they are detected
	require.NoError(t, os.MkdirAll(filepath.Join(home, ".claude"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(home, ".agents"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(home, ".config", "opencode"), 0o755))

	original := newDetectorFunc
	newDetectorFunc = func() (*platform.Detector, error) {
		return platform.NewDetectorWithPlatforms([]platform.Platform{
			platform.NewClaudeCode(home, project),
			platform.NewCodexCLI(home, project),
			platform.NewOpenCode(home, project),
		}), nil
	}
	t.Cleanup(func() { newDetectorFunc = original })
}

// setupTestRegistryWithDirs overrides newRegistryFunc and returns the dirs used.
func setupTestRegistryWithDirs(t *testing.T) (userDir, projectDir string) {
	t.Helper()
	userDir = filepath.Join(t.TempDir(), "user-skills")
	projectDir = filepath.Join(t.TempDir(), "project-skills")

	original := newRegistryFunc
	newRegistryFunc = func() (*registry.Registry, error) {
		return registry.New(userDir, projectDir), nil
	}
	t.Cleanup(func() { newRegistryFunc = original })
	return userDir, projectDir
}

// --- skill install ---

func TestSkillInstall(t *testing.T) {
	userDir, _ := setupTestRegistryWithDirs(t)
	home := t.TempDir()
	project := t.TempDir()
	setupTestDetector(t, home, project)

	// Create a skill first
	_, err := runCmd(t, "skill", "create", "install-me", "--description", "A test skill")
	require.NoError(t, err)

	// Install to claude-code
	out, err := runCmd(t, "skill", "install", "install-me", "--platform", "claude-code", "--json")
	require.NoError(t, err)

	var result output.SkillInstallResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	assert.Equal(t, "install-me", result.Skill)
	assert.Len(t, result.Platforms, 1)
	assert.True(t, result.Platforms[0].Success)
	assert.Equal(t, "claude-code", result.Platforms[0].Platform)

	// Verify file exists
	installed := filepath.Join(home, ".claude", "skills", "install-me", "SKILL.md")
	_, err = os.Stat(installed)
	require.NoError(t, err)

	// Verify it's a copy of the registry skill
	registrySKILL := filepath.Join(userDir, "install-me", "SKILL.md")
	regContent, err := os.ReadFile(registrySKILL)
	require.NoError(t, err)
	installedContent, err := os.ReadFile(installed)
	require.NoError(t, err)
	assert.Equal(t, string(regContent), string(installedContent))
}

func TestSkillInstall_Text(t *testing.T) {
	setupTestRegistryWithDirs(t)
	home := t.TempDir()
	project := t.TempDir()
	setupTestDetector(t, home, project)

	_, err := runCmd(t, "skill", "create", "text-install", "--description", "Test")
	require.NoError(t, err)

	out, err := runCmd(t, "skill", "install", "text-install", "--platform", "claude-code")
	require.NoError(t, err)
	assert.Contains(t, out, "Installed")
	assert.Contains(t, out, "text-install")
	assert.Contains(t, out, "claude-code")
}

func TestSkillInstall_AllPlatforms(t *testing.T) {
	setupTestRegistryWithDirs(t)
	home := t.TempDir()
	project := t.TempDir()
	setupTestDetector(t, home, project)

	_, err := runCmd(t, "skill", "create", "all-platforms", "--description", "Test")
	require.NoError(t, err)

	out, err := runCmd(t, "skill", "install", "all-platforms", "--platform", "all", "--json")
	require.NoError(t, err)

	var result output.SkillInstallResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	assert.Equal(t, "all-platforms", result.Skill)
	assert.Len(t, result.Platforms, 3)
	for _, p := range result.Platforms {
		assert.True(t, p.Success, "expected success for %s", p.Platform)
	}
}

func TestSkillInstall_Duplicate(t *testing.T) {
	setupTestRegistryWithDirs(t)
	home := t.TempDir()
	project := t.TempDir()
	setupTestDetector(t, home, project)

	_, err := runCmd(t, "skill", "create", "dup-install", "--description", "Test")
	require.NoError(t, err)

	_, err = runCmd(t, "skill", "install", "dup-install", "--platform", "claude-code")
	require.NoError(t, err)

	// Second install should fail
	_, err = runCmd(t, "skill", "install", "dup-install", "--platform", "claude-code")
	assert.Error(t, err)
}

func TestSkillInstall_NotFound(t *testing.T) {
	setupTestRegistryWithDirs(t)
	home := t.TempDir()
	project := t.TempDir()
	setupTestDetector(t, home, project)

	_, err := runCmd(t, "skill", "install", "nonexistent", "--platform", "claude-code")
	assert.Error(t, err)
}

func TestSkillInstall_InvalidPlatform(t *testing.T) {
	setupTestRegistryWithDirs(t)
	home := t.TempDir()
	project := t.TempDir()
	setupTestDetector(t, home, project)

	_, err := runCmd(t, "skill", "create", "my-skill", "--description", "Test")
	require.NoError(t, err)

	_, err = runCmd(t, "skill", "install", "my-skill", "--platform", "invalid")
	assert.Error(t, err)
}

func TestSkillInstall_MissingPlatformFlag(t *testing.T) {
	setupTestRegistryWithDirs(t)
	home := t.TempDir()
	project := t.TempDir()
	setupTestDetector(t, home, project)

	_, err := runCmd(t, "skill", "install", "my-skill")
	assert.Error(t, err)
}

func TestSkillInstall_InvalidName(t *testing.T) {
	setupTestRegistryWithDirs(t)
	home := t.TempDir()
	project := t.TempDir()
	setupTestDetector(t, home, project)

	_, err := runCmd(t, "skill", "install", "INVALID", "--platform", "claude-code")
	assert.Error(t, err)
}

// --- skill uninstall ---

func TestSkillUninstall(t *testing.T) {
	setupTestRegistryWithDirs(t)
	home := t.TempDir()
	project := t.TempDir()
	setupTestDetector(t, home, project)

	_, err := runCmd(t, "skill", "create", "remove-platform", "--description", "Test")
	require.NoError(t, err)

	_, err = runCmd(t, "skill", "install", "remove-platform", "--platform", "claude-code")
	require.NoError(t, err)

	out, err := runCmd(t, "skill", "uninstall", "remove-platform", "--platform", "claude-code", "--json")
	require.NoError(t, err)

	var result output.SkillUninstallResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	assert.Equal(t, "remove-platform", result.Skill)
	assert.Len(t, result.Platforms, 1)
	assert.True(t, result.Platforms[0].Success)

	// Verify removed
	installed := filepath.Join(home, ".claude", "skills", "remove-platform")
	_, err = os.Stat(installed)
	assert.True(t, os.IsNotExist(err))
}

func TestSkillUninstall_Text(t *testing.T) {
	setupTestRegistryWithDirs(t)
	home := t.TempDir()
	project := t.TempDir()
	setupTestDetector(t, home, project)

	_, err := runCmd(t, "skill", "create", "text-uninstall", "--description", "Test")
	require.NoError(t, err)

	_, err = runCmd(t, "skill", "install", "text-uninstall", "--platform", "claude-code")
	require.NoError(t, err)

	out, err := runCmd(t, "skill", "uninstall", "text-uninstall", "--platform", "claude-code")
	require.NoError(t, err)
	assert.Contains(t, out, "Uninstalled")
	assert.Contains(t, out, "text-uninstall")
}

func TestSkillUninstall_NotInstalled(t *testing.T) {
	setupTestRegistryWithDirs(t)
	home := t.TempDir()
	project := t.TempDir()
	setupTestDetector(t, home, project)

	_, err := runCmd(t, "skill", "uninstall", "nonexistent", "--platform", "claude-code")
	assert.Error(t, err)
}

// --- platform list ---

func TestPlatformList(t *testing.T) {
	home := t.TempDir()
	project := t.TempDir()
	setupTestDetector(t, home, project)

	out, err := runCmd(t, "platform", "list", "--json")
	require.NoError(t, err)

	var result output.PlatformListResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	assert.Equal(t, 3, result.Count)

	// All should be detected since setupTestDetector creates the directories
	for _, p := range result.Platforms {
		assert.True(t, p.Detected, "expected %s to be detected", p.Name)
	}
}

func TestPlatformList_Text(t *testing.T) {
	home := t.TempDir()
	project := t.TempDir()
	setupTestDetector(t, home, project)

	out, err := runCmd(t, "platform", "list")
	require.NoError(t, err)
	assert.Contains(t, out, "claude-code")
	assert.Contains(t, out, "codex-cli")
	assert.Contains(t, out, "opencode")
	assert.Contains(t, out, "yes")
}

func TestPlatformList_PartialDetection(t *testing.T) {
	home := t.TempDir()
	project := t.TempDir()

	// Only create .claude directory
	require.NoError(t, os.MkdirAll(filepath.Join(home, ".claude"), 0o755))

	original := newDetectorFunc
	newDetectorFunc = func() (*platform.Detector, error) {
		return platform.NewDetectorWithPlatforms([]platform.Platform{
			platform.NewClaudeCode(home, project),
			platform.NewCodexCLI(home, project),
			platform.NewOpenCode(home, project),
		}), nil
	}
	t.Cleanup(func() { newDetectorFunc = original })

	out, err := runCmd(t, "platform", "list", "--json")
	require.NoError(t, err)

	var result output.PlatformListResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))

	detectedCount := 0
	for _, p := range result.Platforms {
		if p.Detected {
			detectedCount++
			assert.Equal(t, "claude-code", p.Name)
		}
	}
	assert.Equal(t, 1, detectedCount)
}

// --- platform status ---

func TestPlatformStatus_Empty(t *testing.T) {
	setupTestRegistryWithDirs(t)
	home := t.TempDir()
	project := t.TempDir()
	setupTestDetector(t, home, project)

	out, err := runCmd(t, "platform", "status", "--json")
	require.NoError(t, err)

	var result output.PlatformStatusResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	assert.Equal(t, "user", result.Scope)
	assert.Empty(t, result.Status)
}

func TestPlatformStatus_WithSkills(t *testing.T) {
	setupTestRegistryWithDirs(t)
	home := t.TempDir()
	project := t.TempDir()
	setupTestDetector(t, home, project)

	// Create and install a skill
	_, err := runCmd(t, "skill", "create", "status-skill", "--description", "Test")
	require.NoError(t, err)

	_, err = runCmd(t, "skill", "install", "status-skill", "--platform", "claude-code")
	require.NoError(t, err)

	out, err := runCmd(t, "platform", "status", "--json")
	require.NoError(t, err)

	var result output.PlatformStatusResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	assert.Len(t, result.Status, 1)
	assert.Equal(t, "status-skill", result.Status[0].Skill)

	// Find claude-code entry
	var found bool
	for _, p := range result.Status[0].Platforms {
		if p.Platform == "claude-code" {
			assert.True(t, p.Installed)
			found = true
		}
	}
	assert.True(t, found, "expected claude-code entry in platforms")
}

func TestPlatformStatus_Text(t *testing.T) {
	setupTestRegistryWithDirs(t)
	home := t.TempDir()
	project := t.TempDir()
	setupTestDetector(t, home, project)

	_, err := runCmd(t, "skill", "create", "text-status", "--description", "Test")
	require.NoError(t, err)

	_, err = runCmd(t, "skill", "install", "text-status", "--platform", "claude-code")
	require.NoError(t, err)

	out, err := runCmd(t, "platform", "status")
	require.NoError(t, err)
	assert.Contains(t, out, "text-status")
	assert.Contains(t, out, "installed")
}

func TestPlatformStatus_ProjectScope(t *testing.T) {
	setupTestRegistryWithDirs(t)
	home := t.TempDir()
	project := t.TempDir()
	setupTestDetector(t, home, project)

	// Create skill in project scope
	_, err := runCmd(t, "skill", "create", "proj-status", "--scope", "project", "--description", "Test")
	require.NoError(t, err)

	_, err = runCmd(t, "skill", "install", "proj-status", "--platform", "claude-code", "--scope", "project")
	require.NoError(t, err)

	out, err := runCmd(t, "platform", "status", "--scope", "project", "--json")
	require.NoError(t, err)

	var result output.PlatformStatusResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	assert.Equal(t, "project", result.Scope)
	assert.Len(t, result.Status, 1)
	assert.Equal(t, "proj-status", result.Status[0].Skill)
}
