package cli

import (
	"encoding/json"
	"errors"
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

// --- #102: --tag / --category on install and uninstall ---

// setupTaggedSkills creates four registry skills in user scope:
// two tagged "workflow" (one also lang:go), one lang:python only, one untagged.
func setupTaggedSkills(t *testing.T, cc *CommandContext) {
	t.Helper()
	for _, spec := range []struct{ name, tags string }{
		{"wf-plan", "workflow,lang:go"},
		{"wf-review", "workflow"},
		{"py-only", "lang:python"},
		{"plain", ""},
	} {
		args := []string{"skill", "create", spec.name, "--description", "Use when testing tag installs."}
		if spec.tags != "" {
			args = append(args, "--tags", spec.tags)
		}
		_, err := runCmd(t, cc, args...)
		require.NoError(t, err)
	}
}

func TestSkillInstall_ByTag(t *testing.T) {
	cc, _, _ := testRegistryWithDirs(t)
	home := t.TempDir()
	withTestDetector(t, cc, home, t.TempDir())
	setupTaggedSkills(t, cc)

	out, err := runCmd(t, cc, "skill", "install", "--tag", "workflow", "--platform", "claude-code", "--json")
	require.NoError(t, err)

	var result output.SkillInstallResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	require.Len(t, result.Skills, 2)
	assert.Equal(t, "wf-plan", result.Skills[0].Skill, "resolved names are sorted")
	assert.Equal(t, "wf-review", result.Skills[1].Skill)
	for _, e := range result.Skills {
		assert.True(t, e.Success, "%s should install", e.Skill)
	}
	for _, n := range []string{"wf-plan", "wf-review"} {
		_, err := os.Stat(filepath.Join(home, ".claude", "skills", n, "SKILL.md"))
		require.NoError(t, err)
	}
	for _, n := range []string{"py-only", "plain"} {
		_, err := os.Stat(filepath.Join(home, ".claude", "skills", n))
		assert.True(t, os.IsNotExist(err), "%s must not be installed", n)
	}
}

func TestSkillInstall_ByCategory_AndsWithTag(t *testing.T) {
	cc, _, _ := testRegistryWithDirs(t)
	home := t.TempDir()
	withTestDetector(t, cc, home, t.TempDir())
	setupTaggedSkills(t, cc)

	// --tag and --category compose with AND: only wf-plan is workflow AND lang:go.
	out, err := runCmd(t, cc, "skill", "install", "--tag", "workflow", "--category", "lang:go",
		"--platform", "claude-code", "--json")
	require.NoError(t, err)
	var result output.SkillInstallResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	require.Len(t, result.Skills, 1)
	assert.Equal(t, "wf-plan", result.Skills[0].Skill)

	// --category alone, comma list ORs within the namespace.
	out, err = runCmd(t, cc, "skill", "install", "--category", "lang:go,python",
		"--platform", "codex-cli", "--json")
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	require.Len(t, result.Skills, 2)
	assert.Equal(t, "py-only", result.Skills[0].Skill)
	assert.Equal(t, "wf-plan", result.Skills[1].Skill)
}

func TestSkillInstall_Filter_EmptyMatchIsError(t *testing.T) {
	cc, _, _ := testRegistryWithDirs(t)
	withTestDetector(t, cc, t.TempDir(), t.TempDir())
	setupTaggedSkills(t, cc)

	out, err := runCmd(t, cc, "skill", "install", "--tag", "nope", "--platform", "claude-code")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no registered skills match --tag nope in user scope")
	assert.NotContains(t, out, "Installed", "an empty match must not be a silent no-op")
	// Empty match is an operational error (exit 1), not a usage error.
	var ve *ValidationError
	assert.False(t, errors.As(err, &ve))
}

func TestSkillInstall_Filter_MutuallyExclusiveWithNames(t *testing.T) {
	cc, _, _ := testRegistryWithDirs(t)
	withTestDetector(t, cc, t.TempDir(), t.TempDir())
	setupTaggedSkills(t, cc)

	_, err := runCmd(t, cc, "skill", "install", "plain", "--tag", "workflow", "--platform", "claude-code")
	require.Error(t, err)
	var ve *ValidationError
	require.ErrorAs(t, err, &ve)
	assert.Contains(t, err.Error(), "mutually exclusive")

	// Neither names nor a filter is a usage error too.
	_, err = runCmd(t, cc, "skill", "install", "--platform", "claude-code")
	require.Error(t, err)
	require.ErrorAs(t, err, &ve)
	assert.Contains(t, err.Error(), "requires at least one skill name, or a --tag/--category filter")

	// Malformed --category surfaces the same validation error list uses.
	_, err = runCmd(t, cc, "skill", "install", "--category", "noncolon", "--platform", "claude-code")
	require.Error(t, err)
	require.ErrorAs(t, err, &ve)
}

func TestSkillInstall_Filter_RespectsScope(t *testing.T) {
	cc, _, _ := testRegistryWithDirs(t)
	home := t.TempDir()
	project := t.TempDir()
	withTestDetector(t, cc, home, project)
	setupTaggedSkills(t, cc) // user scope
	_, err := runCmd(t, cc, "skill", "create", "proj-wf", "--description", "Use when testing.", "--tags", "workflow", "--scope", "project")
	require.NoError(t, err)

	// The filter resolves against the registry at --scope only.
	out, err := runCmd(t, cc, "skill", "install", "--tag", "workflow", "--platform", "claude-code", "--scope", "project", "--json")
	require.NoError(t, err)
	var result output.SkillInstallResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	require.Len(t, result.Skills, 1)
	assert.Equal(t, "proj-wf", result.Skills[0].Skill)
	_, err = os.Stat(filepath.Join(project, ".claude", "skills", "proj-wf", "SKILL.md"))
	require.NoError(t, err)
}

func TestSkillInstall_Filter_EnforceBudgetCountsResolved(t *testing.T) {
	cc, _, _ := testRegistryWithDirs(t)
	withTestDetector(t, cc, t.TempDir(), t.TempDir())
	setupTaggedSkills(t, cc)

	// Two workflow skills resolve; a budget check must see 2, not 0 args.
	// Under the threshold it proceeds normally.
	out, err := runCmd(t, cc, "skill", "install", "--tag", "workflow", "--platform", "claude-code", "--enforce-budget", "--json")
	require.NoError(t, err)
	var result output.SkillInstallResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	require.Len(t, result.Skills, 2)
	require.NotNil(t, result.Capacity)
	assert.Equal(t, 2, result.Capacity.Installed)
}

func TestSkillUninstall_ByTag_OnlyInstalledMatches(t *testing.T) {
	cc, _, _ := testRegistryWithDirs(t)
	home := t.TempDir()
	withTestDetector(t, cc, home, t.TempDir())
	setupTaggedSkills(t, cc)

	// Install one workflow skill and the untagged one; leave wf-review uninstalled.
	_, err := runCmd(t, cc, "skill", "install", "wf-plan", "plain", "--platform", "claude-code")
	require.NoError(t, err)

	out, err := runCmd(t, cc, "skill", "uninstall", "--tag", "workflow", "--platform", "claude-code", "--json")
	require.NoError(t, err)

	var result output.SkillUninstallResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	// wf-review is tagged but not installed: skipped, not a failure entry.
	require.Len(t, result.Skills, 1)
	assert.Equal(t, "wf-plan", result.Skills[0].Skill)
	assert.True(t, result.Skills[0].Success)

	_, err = os.Stat(filepath.Join(home, ".claude", "skills", "wf-plan"))
	assert.True(t, os.IsNotExist(err), "wf-plan should be removed")
	_, err = os.Stat(filepath.Join(home, ".claude", "skills", "plain", "SKILL.md"))
	require.NoError(t, err, "untagged skill must survive a tag-scoped uninstall")
	// Registry copies are untouched either way.
	_, err = runCmd(t, cc, "skill", "show", "wf-plan")
	require.NoError(t, err)
}

func TestSkillUninstall_ByTag_NothingInstalledIsError(t *testing.T) {
	cc, _, _ := testRegistryWithDirs(t)
	withTestDetector(t, cc, t.TempDir(), t.TempDir())
	setupTaggedSkills(t, cc)

	_, err := runCmd(t, cc, "skill", "uninstall", "--tag", "workflow", "--platform", "claude-code")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no installed skills match --tag workflow on claude-code (user scope)")

	// Unknown tag: fails at the registry step with the same message install uses.
	_, err = runCmd(t, cc, "skill", "uninstall", "--tag", "nope", "--platform", "claude-code")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no registered skills match --tag nope in user scope")
}

func TestSkillUninstall_Filter_MutuallyExclusiveWithNames(t *testing.T) {
	cc, _, _ := testRegistryWithDirs(t)
	withTestDetector(t, cc, t.TempDir(), t.TempDir())

	_, err := runCmd(t, cc, "skill", "uninstall", "plain", "--tag", "workflow", "--platform", "claude-code")
	require.Error(t, err)
	var ve *ValidationError
	require.ErrorAs(t, err, &ve)
	assert.Contains(t, err.Error(), "mutually exclusive")

	_, err = runCmd(t, cc, "skill", "uninstall", "--platform", "claude-code")
	require.Error(t, err)
	require.ErrorAs(t, err, &ve)
}
