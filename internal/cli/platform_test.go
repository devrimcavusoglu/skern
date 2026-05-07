package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/devrimcavusoglu/skern/internal/output"
	"github.com/devrimcavusoglu/skern/internal/platform"
	"github.com/devrimcavusoglu/skern/internal/registry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// withTestDetector configures a CommandContext with a test detector using temp directories.
func withTestDetector(t *testing.T, cc *CommandContext, home, project string) {
	t.Helper()

	// Create platform directories so they are detected
	require.NoError(t, os.MkdirAll(filepath.Join(home, ".claude"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(home, ".agents"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(home, ".config", "opencode"), 0o755))

	cc.NewDetector = func() (*platform.Detector, error) {
		return platform.NewDetectorWithPlatforms([]platform.Platform{
			platform.New(platform.TypeClaudeCode, home, project),
			platform.New(platform.TypeCodexCLI, home, project),
			platform.New(platform.TypeOpenCode, home, project),
		}), nil
	}
}

// testRegistryWithDirs returns a CommandContext with temp registry dirs.
func testRegistryWithDirs(t *testing.T) (cc *CommandContext, userDir, projectDir string) {
	t.Helper()
	userDir = filepath.Join(t.TempDir(), "user-skills")
	projectDir = filepath.Join(t.TempDir(), "project-skills")

	cc = &CommandContext{
		NewRegistry: func() (*registry.Registry, error) {
			return registry.New(userDir, projectDir), nil
		},
		NewDetector: defaultNewDetector,
	}
	return cc, userDir, projectDir
}

// --- skill install ---

func TestSkillInstall(t *testing.T) {
	cc, userDir, _ := testRegistryWithDirs(t)
	home := t.TempDir()
	project := t.TempDir()
	withTestDetector(t, cc, home, project)

	// Create a skill first
	_, err := runCmd(t, cc, "skill", "create", "install-me", "--description", "A test skill")
	require.NoError(t, err)

	// Install to claude-code
	out, err := runCmd(t, cc, "skill", "install", "install-me", "--platform", "claude-code", "--json")
	require.NoError(t, err)

	var result output.SkillInstallResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	assert.Equal(t, "claude-code", result.Platform)
	require.Len(t, result.Skills, 1)
	assert.Equal(t, "install-me", result.Skills[0].Skill)
	assert.True(t, result.Skills[0].Success)
	require.NotNil(t, result.Capacity)
	assert.Equal(t, "claude-code", result.Capacity.Platform)
	assert.Equal(t, 1, result.Capacity.Installed)

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
	cc, _, _ := testRegistryWithDirs(t)
	home := t.TempDir()
	project := t.TempDir()
	withTestDetector(t, cc, home, project)

	_, err := runCmd(t, cc, "skill", "create", "text-install", "--description", "Test")
	require.NoError(t, err)

	out, err := runCmd(t, cc, "skill", "install", "text-install", "--platform", "claude-code")
	require.NoError(t, err)
	assert.Contains(t, out, "Installed")
	assert.Contains(t, out, "text-install")
	assert.Contains(t, out, "claude-code")
}

// TestSkillInstall_PlatformAllRejected verifies that "--platform all" is no
// longer accepted (per #52 D6). Agents must specify the platform they're
// running on; bulk-deploy semantics are removed.
func TestSkillInstall_PlatformAllRejected(t *testing.T) {
	cc, _, _ := testRegistryWithDirs(t)
	home := t.TempDir()
	project := t.TempDir()
	withTestDetector(t, cc, home, project)

	_, err := runCmd(t, cc, "skill", "create", "all-rejected", "--description", "Test")
	require.NoError(t, err)

	_, err = runCmd(t, cc, "skill", "install", "all-rejected", "--platform", "all")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no longer supported")
}

// TestSkillInstall_Batch verifies that multiple skill names install in a
// single invocation and each gets its own per-skill action entry.
func TestSkillInstall_Batch(t *testing.T) {
	cc, _, _ := testRegistryWithDirs(t)
	home := t.TempDir()
	project := t.TempDir()
	withTestDetector(t, cc, home, project)

	for _, n := range []string{"batch-a", "batch-b", "batch-c"} {
		_, err := runCmd(t, cc, "skill", "create", n, "--description", "Test")
		require.NoError(t, err)
	}

	out, err := runCmd(t, cc, "skill", "install",
		"batch-a", "batch-b", "batch-c",
		"--platform", "claude-code", "--json")
	require.NoError(t, err)

	var result output.SkillInstallResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	assert.Equal(t, "claude-code", result.Platform)
	require.Len(t, result.Skills, 3)
	for i, want := range []string{"batch-a", "batch-b", "batch-c"} {
		assert.Equal(t, want, result.Skills[i].Skill, "input order should be preserved")
		assert.True(t, result.Skills[i].Success)
	}
	require.NotNil(t, result.Capacity)
	assert.Equal(t, 3, result.Capacity.Installed)

	// Files exist on disk for each skill.
	for _, n := range []string{"batch-a", "batch-b", "batch-c"} {
		_, err := os.Stat(filepath.Join(home, ".claude", "skills", n, "SKILL.md"))
		require.NoError(t, err, "expected %s to be installed", n)
	}
}

// TestSkillInstall_BatchPartialFailure verifies that a missing skill in a
// batch produces a per-skill error without aborting the whole batch.
func TestSkillInstall_BatchPartialFailure(t *testing.T) {
	cc, _, _ := testRegistryWithDirs(t)
	home := t.TempDir()
	project := t.TempDir()
	withTestDetector(t, cc, home, project)

	_, err := runCmd(t, cc, "skill", "create", "real-skill", "--description", "Test")
	require.NoError(t, err)

	out, err := runCmd(t, cc, "skill", "install",
		"real-skill", "ghost-skill",
		"--platform", "claude-code", "--json")
	// Exit code 0 because at least one install succeeded.
	require.NoError(t, err)

	var result output.SkillInstallResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	require.Len(t, result.Skills, 2)
	assert.True(t, result.Skills[0].Success, "real-skill should succeed")
	assert.False(t, result.Skills[1].Success, "ghost-skill should fail")
	assert.Contains(t, result.Skills[1].Error, "not registered")
}

// TestSkillInstall_BatchAllFail verifies a non-zero exit when every skill in
// the batch fails to install.
func TestSkillInstall_BatchAllFail(t *testing.T) {
	cc, _, _ := testRegistryWithDirs(t)
	home := t.TempDir()
	project := t.TempDir()
	withTestDetector(t, cc, home, project)

	_, err := runCmd(t, cc, "skill", "install",
		"ghost-a", "ghost-b",
		"--platform", "claude-code")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to install any skill")
}

// TestSkillInstall_EnforceBudget verifies that --enforce-budget refuses to
// install when the resulting count would exceed the platform threshold.
func TestSkillInstall_EnforceBudget(t *testing.T) {
	cc, _, _ := testRegistryWithDirs(t)
	home := t.TempDir()
	project := t.TempDir()
	withTestDetector(t, cc, home, project)

	// Project scope has the smaller threshold (20). Pre-fill the platform
	// dir so the threshold is hit without creating 20 registry skills.
	skillsDir := filepath.Join(project, ".claude", "skills")
	require.NoError(t, os.MkdirAll(skillsDir, 0o755))
	for i := 0; i < 20; i++ {
		dir := filepath.Join(skillsDir, fmt.Sprintf("filler-%02d", i))
		require.NoError(t, os.MkdirAll(dir, 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(dir, "SKILL.md"),
			[]byte("---\nname: filler\ndescription: x\n---\n"), 0o644))
	}

	_, err := runCmd(t, cc, "skill", "create", "over-budget",
		"--description", "Test", "--scope", "project")
	require.NoError(t, err)

	_, err = runCmd(t, cc, "skill", "install", "over-budget",
		"--platform", "claude-code", "--scope", "project", "--enforce-budget")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "capacity")

	// Without --enforce-budget the same install proceeds (only a warning).
	_, err = runCmd(t, cc, "skill", "install", "over-budget",
		"--platform", "claude-code", "--scope", "project")
	require.NoError(t, err)
}

func TestSkillInstall_Duplicate(t *testing.T) {
	cc, _, _ := testRegistryWithDirs(t)
	home := t.TempDir()
	project := t.TempDir()
	withTestDetector(t, cc, home, project)

	_, err := runCmd(t, cc, "skill", "create", "dup-install", "--description", "Test")
	require.NoError(t, err)

	_, err = runCmd(t, cc, "skill", "install", "dup-install", "--platform", "claude-code")
	require.NoError(t, err)

	// Second install should fail
	_, err = runCmd(t, cc, "skill", "install", "dup-install", "--platform", "claude-code")
	assert.Error(t, err)
}

func TestSkillInstall_Force(t *testing.T) {
	cc, _, _ := testRegistryWithDirs(t)
	home := t.TempDir()
	project := t.TempDir()
	withTestDetector(t, cc, home, project)

	_, err := runCmd(t, cc, "skill", "create", "force-install", "--description", "Test")
	require.NoError(t, err)

	_, err = runCmd(t, cc, "skill", "install", "force-install", "--platform", "claude-code")
	require.NoError(t, err)

	// Second install with --force should succeed
	out, err := runCmd(t, cc, "skill", "install", "force-install", "--platform", "claude-code", "--force", "--json")
	require.NoError(t, err)

	var result output.SkillInstallResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	require.Len(t, result.Skills, 1)
	assert.True(t, result.Skills[0].Success)

	// Verify file still exists
	installed := filepath.Join(home, ".claude", "skills", "force-install", "SKILL.md")
	_, err = os.Stat(installed)
	require.NoError(t, err)
}

func TestSkillInstall_NotFound(t *testing.T) {
	cc, _, _ := testRegistryWithDirs(t)
	home := t.TempDir()
	project := t.TempDir()
	withTestDetector(t, cc, home, project)

	_, err := runCmd(t, cc, "skill", "install", "nonexistent", "--platform", "claude-code")
	assert.Error(t, err)
}

func TestSkillInstall_InvalidPlatform(t *testing.T) {
	cc, _, _ := testRegistryWithDirs(t)
	home := t.TempDir()
	project := t.TempDir()
	withTestDetector(t, cc, home, project)

	_, err := runCmd(t, cc, "skill", "create", "my-skill", "--description", "Test")
	require.NoError(t, err)

	_, err = runCmd(t, cc, "skill", "install", "my-skill", "--platform", "invalid")
	assert.Error(t, err)
}

func TestSkillInstall_MissingPlatformFlag(t *testing.T) {
	cc, _, _ := testRegistryWithDirs(t)
	home := t.TempDir()
	project := t.TempDir()
	withTestDetector(t, cc, home, project)

	_, err := runCmd(t, cc, "skill", "install", "my-skill")
	assert.Error(t, err)
}

func TestSkillInstall_InvalidName(t *testing.T) {
	cc, _, _ := testRegistryWithDirs(t)
	home := t.TempDir()
	project := t.TempDir()
	withTestDetector(t, cc, home, project)

	_, err := runCmd(t, cc, "skill", "install", "INVALID", "--platform", "claude-code")
	assert.Error(t, err)
}

// --- skill uninstall ---

func TestSkillUninstall(t *testing.T) {
	cc, _, _ := testRegistryWithDirs(t)
	home := t.TempDir()
	project := t.TempDir()
	withTestDetector(t, cc, home, project)

	_, err := runCmd(t, cc, "skill", "create", "remove-platform", "--description", "Test")
	require.NoError(t, err)

	_, err = runCmd(t, cc, "skill", "install", "remove-platform", "--platform", "claude-code")
	require.NoError(t, err)

	out, err := runCmd(t, cc, "skill", "uninstall", "remove-platform", "--platform", "claude-code", "--json")
	require.NoError(t, err)

	var result output.SkillUninstallResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	assert.Equal(t, "claude-code", result.Platform)
	require.Len(t, result.Skills, 1)
	assert.Equal(t, "remove-platform", result.Skills[0].Skill)
	assert.True(t, result.Skills[0].Success)

	// Verify removed
	installed := filepath.Join(home, ".claude", "skills", "remove-platform")
	_, err = os.Stat(installed)
	assert.True(t, os.IsNotExist(err))
}

func TestSkillUninstall_Text(t *testing.T) {
	cc, _, _ := testRegistryWithDirs(t)
	home := t.TempDir()
	project := t.TempDir()
	withTestDetector(t, cc, home, project)

	_, err := runCmd(t, cc, "skill", "create", "text-uninstall", "--description", "Test")
	require.NoError(t, err)

	_, err = runCmd(t, cc, "skill", "install", "text-uninstall", "--platform", "claude-code")
	require.NoError(t, err)

	out, err := runCmd(t, cc, "skill", "uninstall", "text-uninstall", "--platform", "claude-code")
	require.NoError(t, err)
	assert.Contains(t, out, "Uninstalled")
	assert.Contains(t, out, "text-uninstall")
}

func TestSkillUninstall_NotInstalled(t *testing.T) {
	cc, _, _ := testRegistryWithDirs(t)
	home := t.TempDir()
	project := t.TempDir()
	withTestDetector(t, cc, home, project)

	_, err := runCmd(t, cc, "skill", "uninstall", "nonexistent", "--platform", "claude-code")
	assert.Error(t, err)
}

// --- platform list ---

func TestPlatformList(t *testing.T) {
	cc := &CommandContext{
		NewRegistry: defaultNewRegistry,
		NewDetector: defaultNewDetector,
	}
	home := t.TempDir()
	project := t.TempDir()
	withTestDetector(t, cc, home, project)

	out, err := runCmd(t, cc, "platform", "list", "--json")
	require.NoError(t, err)

	var result output.PlatformListResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	assert.Equal(t, 3, result.Count)

	// All should be detected since withTestDetector creates the directories
	for _, p := range result.Platforms {
		assert.True(t, p.Detected, "expected %s to be detected", p.Name)
	}
}

func TestPlatformList_Text(t *testing.T) {
	cc := &CommandContext{
		NewRegistry: defaultNewRegistry,
		NewDetector: defaultNewDetector,
	}
	home := t.TempDir()
	project := t.TempDir()
	withTestDetector(t, cc, home, project)

	out, err := runCmd(t, cc, "platform", "list")
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

	cc := &CommandContext{
		NewRegistry: defaultNewRegistry,
		NewDetector: func() (*platform.Detector, error) {
			return platform.NewDetectorWithPlatforms([]platform.Platform{
				platform.New(platform.TypeClaudeCode, home, project),
				platform.New(platform.TypeCodexCLI, home, project),
				platform.New(platform.TypeOpenCode, home, project),
			}), nil
		},
	}

	out, err := runCmd(t, cc, "platform", "list", "--json")
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
	cc, _, _ := testRegistryWithDirs(t)
	home := t.TempDir()
	project := t.TempDir()
	withTestDetector(t, cc, home, project)

	out, err := runCmd(t, cc, "platform", "status", "--json")
	require.NoError(t, err)

	var result output.PlatformStatusResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	assert.Equal(t, "user", result.Scope)
	assert.Empty(t, result.Status)
}

func TestPlatformStatus_WithSkills(t *testing.T) {
	cc, _, _ := testRegistryWithDirs(t)
	home := t.TempDir()
	project := t.TempDir()
	withTestDetector(t, cc, home, project)

	// Create and install a skill
	_, err := runCmd(t, cc, "skill", "create", "status-skill", "--description", "Test")
	require.NoError(t, err)

	_, err = runCmd(t, cc, "skill", "install", "status-skill", "--platform", "claude-code")
	require.NoError(t, err)

	out, err := runCmd(t, cc, "platform", "status", "--json")
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
	cc, _, _ := testRegistryWithDirs(t)
	home := t.TempDir()
	project := t.TempDir()
	withTestDetector(t, cc, home, project)

	_, err := runCmd(t, cc, "skill", "create", "text-status", "--description", "Test")
	require.NoError(t, err)

	_, err = runCmd(t, cc, "skill", "install", "text-status", "--platform", "claude-code")
	require.NoError(t, err)

	out, err := runCmd(t, cc, "platform", "status")
	require.NoError(t, err)
	assert.Contains(t, out, "text-status")
	assert.Contains(t, out, "installed")
}

func TestPlatformStatus_ProjectScope(t *testing.T) {
	cc, _, _ := testRegistryWithDirs(t)
	home := t.TempDir()
	project := t.TempDir()
	withTestDetector(t, cc, home, project)

	// Create skill in project scope
	_, err := runCmd(t, cc, "skill", "create", "proj-status", "--scope", "project", "--description", "Test")
	require.NoError(t, err)

	_, err = runCmd(t, cc, "skill", "install", "proj-status", "--platform", "claude-code", "--scope", "project")
	require.NoError(t, err)

	out, err := runCmd(t, cc, "platform", "status", "--scope", "project", "--json")
	require.NoError(t, err)

	var result output.PlatformStatusResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	assert.Equal(t, "project", result.Scope)
	assert.Len(t, result.Status, 1)
	assert.Equal(t, "proj-status", result.Status[0].Skill)
}
