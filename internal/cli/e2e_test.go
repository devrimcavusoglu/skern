package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/devrimcavusoglu/scribe/internal/output"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEndToEnd_FullLifecycle exercises the complete scribe workflow:
// create -> validate -> install (all 3 platforms) -> platform status -> uninstall one ->
// verify remaining -> uninstall all -> remove from registry.
func TestEndToEnd_FullLifecycle(t *testing.T) {
	userDir, _ := setupTestRegistryWithDirs(t)
	home := t.TempDir()
	project := t.TempDir()
	setupTestDetector(t, home, project)

	const skillName = "e2e-test-skill"
	const skillDesc = "End-to-end test skill for lifecycle validation"

	// Step 1: Create a skill with author metadata
	out, err := runCmd(t, "skill", "create", skillName,
		"--description", skillDesc,
		"--author", "e2e-tester",
		"--author-type", "human",
		"--json")
	require.NoError(t, err)

	var createResult output.SkillCreateResult
	require.NoError(t, json.Unmarshal([]byte(out), &createResult))
	assert.Equal(t, skillName, createResult.Name)
	assert.Equal(t, "user", createResult.Scope)
	assert.NotEmpty(t, createResult.Path)

	// Verify SKILL.md exists in registry
	skillMD := filepath.Join(userDir, skillName, "SKILL.md")
	_, err = os.Stat(skillMD)
	require.NoError(t, err, "SKILL.md should exist in registry")

	// Step 2: Validate the skill
	out, err = runCmd(t, "skill", "validate", skillName, "--json")
	require.NoError(t, err)

	var validateResult output.SkillValidateResult
	require.NoError(t, json.Unmarshal([]byte(out), &validateResult))
	assert.Equal(t, skillName, validateResult.Name)
	assert.True(t, validateResult.Valid, "freshly created skill should be valid")
	assert.Equal(t, 0, validateResult.Errors, "no validation errors expected")

	// Step 3: Show skill details and verify author
	out, err = runCmd(t, "skill", "show", skillName, "--json")
	require.NoError(t, err)

	var showResult output.SkillResult
	require.NoError(t, json.Unmarshal([]byte(out), &showResult))
	assert.Equal(t, skillName, showResult.Name)
	assert.Equal(t, skillDesc, showResult.Description)
	assert.Equal(t, "e2e-tester", showResult.Author.Name)
	assert.Equal(t, "human", showResult.Author.Type)

	// Step 4: Install to all 3 platforms
	out, err = runCmd(t, "skill", "install", skillName, "--platform", "all", "--json")
	require.NoError(t, err)

	var installResult output.SkillInstallResult
	require.NoError(t, json.Unmarshal([]byte(out), &installResult))
	assert.Equal(t, skillName, installResult.Skill)
	assert.Len(t, installResult.Platforms, 3, "should install to all 3 platforms")
	for _, p := range installResult.Platforms {
		assert.True(t, p.Success, "install should succeed for %s", p.Platform)
		assert.Empty(t, p.Error, "no error expected for %s", p.Platform)
	}

	// Step 5: Verify files exist on all 3 platforms
	platformPaths := map[string]string{
		"claude-code": filepath.Join(home, ".claude", "skills", skillName, "SKILL.md"),
		"codex-cli":   filepath.Join(home, ".agents", "skills", skillName, "SKILL.md"),
		"opencode":    filepath.Join(home, ".config", "opencode", "skills", skillName, "SKILL.md"),
	}
	for plat, path := range platformPaths {
		_, err = os.Stat(path)
		require.NoError(t, err, "SKILL.md should exist on %s at %s", plat, path)
	}

	// Step 6: Verify platform status shows all installed
	out, err = runCmd(t, "platform", "status", "--json")
	require.NoError(t, err)

	var statusResult output.PlatformStatusResult
	require.NoError(t, json.Unmarshal([]byte(out), &statusResult))
	require.Len(t, statusResult.Status, 1, "one skill should appear in status")
	assert.Equal(t, skillName, statusResult.Status[0].Skill)

	installedCount := 0
	for _, p := range statusResult.Status[0].Platforms {
		if p.Installed {
			installedCount++
		}
	}
	assert.Equal(t, 3, installedCount, "skill should be installed on all 3 platforms")

	// Step 7: Uninstall from claude-code only
	out, err = runCmd(t, "skill", "uninstall", skillName, "--platform", "claude-code", "--json")
	require.NoError(t, err)

	var uninstallResult output.SkillUninstallResult
	require.NoError(t, json.Unmarshal([]byte(out), &uninstallResult))
	assert.Equal(t, skillName, uninstallResult.Skill)
	assert.Len(t, uninstallResult.Platforms, 1)
	assert.True(t, uninstallResult.Platforms[0].Success)

	// Verify claude-code no longer has it
	_, err = os.Stat(platformPaths["claude-code"])
	assert.True(t, os.IsNotExist(err), "skill should be removed from claude-code")

	// Verify codex-cli and opencode still have it
	for _, plat := range []string{"codex-cli", "opencode"} {
		_, err = os.Stat(platformPaths[plat])
		require.NoError(t, err, "skill should still be on %s", plat)
	}

	// Step 8: Verify platform status reflects partial uninstall
	out, err = runCmd(t, "platform", "status", "--json")
	require.NoError(t, err)

	require.NoError(t, json.Unmarshal([]byte(out), &statusResult))
	require.Len(t, statusResult.Status, 1)

	for _, p := range statusResult.Status[0].Platforms {
		if p.Platform == "claude-code" {
			assert.False(t, p.Installed, "claude-code should show not installed")
		} else {
			assert.True(t, p.Installed, "%s should still show installed", p.Platform)
		}
	}

	// Step 9: Uninstall from remaining platforms
	_, err = runCmd(t, "skill", "uninstall", skillName, "--platform", "codex-cli")
	require.NoError(t, err)
	_, err = runCmd(t, "skill", "uninstall", skillName, "--platform", "opencode")
	require.NoError(t, err)

	// All platform files should be gone
	for plat, path := range platformPaths {
		_, err = os.Stat(path)
		assert.True(t, os.IsNotExist(err), "skill should be removed from %s", plat)
	}

	// Step 10: Skill should still exist in registry
	out, err = runCmd(t, "skill", "show", skillName, "--json")
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal([]byte(out), &showResult))
	assert.Equal(t, skillName, showResult.Name)

	// Step 11: Search should find the skill
	out, err = runCmd(t, "skill", "search", "e2e", "--json")
	require.NoError(t, err)

	var searchResult output.SkillSearchResult
	require.NoError(t, json.Unmarshal([]byte(out), &searchResult))
	assert.Equal(t, 1, searchResult.Count)
	assert.Equal(t, skillName, searchResult.Results[0].Name)

	// Step 12: Remove from registry
	out, err = runCmd(t, "skill", "remove", skillName, "--json")
	require.NoError(t, err)

	var removeResult output.SkillRemoveResult
	require.NoError(t, json.Unmarshal([]byte(out), &removeResult))
	assert.Equal(t, skillName, removeResult.Name)

	// Step 13: Verify skill is gone
	_, err = runCmd(t, "skill", "show", skillName)
	assert.Error(t, err, "skill should no longer exist in registry")

	// Step 14: List should be empty
	out, err = runCmd(t, "skill", "list", "--json")
	require.NoError(t, err)

	var listResult output.SkillListResult
	require.NoError(t, json.Unmarshal([]byte(out), &listResult))
	assert.Equal(t, 0, listResult.Count)
}

// TestEndToEnd_MultiSkillWorkflow tests managing multiple skills across scopes and platforms.
func TestEndToEnd_MultiSkillWorkflow(t *testing.T) {
	setupTestRegistryWithDirs(t)
	home := t.TempDir()
	project := t.TempDir()
	setupTestDetector(t, home, project)

	// Create skills in different scopes
	_, err := runCmd(t, "skill", "create", "user-formatter", "--scope", "user", "--description", "Formats user code")
	require.NoError(t, err)
	_, err = runCmd(t, "skill", "create", "project-linter", "--scope", "project", "--description", "Lints project code")
	require.NoError(t, err)

	// List all skills
	out, err := runCmd(t, "skill", "list", "--scope", "all", "--json")
	require.NoError(t, err)

	var listResult output.SkillListResult
	require.NoError(t, json.Unmarshal([]byte(out), &listResult))
	assert.Equal(t, 2, listResult.Count)

	// Install user skill to claude-code
	_, err = runCmd(t, "skill", "install", "user-formatter", "--platform", "claude-code")
	require.NoError(t, err)

	// Install project skill to codex-cli (project scope)
	_, err = runCmd(t, "skill", "install", "project-linter", "--platform", "codex-cli", "--scope", "project")
	require.NoError(t, err)

	// Verify user-formatter on claude-code
	installed := filepath.Join(home, ".claude", "skills", "user-formatter", "SKILL.md")
	_, err = os.Stat(installed)
	require.NoError(t, err)

	// Verify project-linter on codex-cli (project scope)
	installed = filepath.Join(project, ".agents", "skills", "project-linter", "SKILL.md")
	_, err = os.Stat(installed)
	require.NoError(t, err)

	// Search across scopes — search matches on name
	out, err = runCmd(t, "skill", "search", "formatter", "--json")
	require.NoError(t, err)

	var searchResult output.SkillSearchResult
	require.NoError(t, json.Unmarshal([]byte(out), &searchResult))
	assert.Equal(t, 1, searchResult.Count)
	assert.Equal(t, "user-formatter", searchResult.Results[0].Name)

	// Clean up
	_, err = runCmd(t, "skill", "uninstall", "user-formatter", "--platform", "claude-code")
	require.NoError(t, err)
	_, err = runCmd(t, "skill", "uninstall", "project-linter", "--platform", "codex-cli", "--scope", "project")
	require.NoError(t, err)
	_, err = runCmd(t, "skill", "remove", "user-formatter")
	require.NoError(t, err)
	_, err = runCmd(t, "skill", "remove", "project-linter", "--scope", "project")
	require.NoError(t, err)
}

// TestEndToEnd_OverlapAndValidation tests the overlap detection and validation
// flows work correctly across the full lifecycle.
func TestEndToEnd_OverlapAndValidation(t *testing.T) {
	setupTestRegistryWithDirs(t)
	home := t.TempDir()
	project := t.TempDir()
	setupTestDetector(t, home, project)

	// Create first skill
	_, err := runCmd(t, "skill", "create", "code-review",
		"--description", "Reviews code for quality issues",
		"--author", "alice", "--author-type", "human")
	require.NoError(t, err)

	// Create similar skill — should succeed (warn level overlap)
	out, err := runCmd(t, "skill", "create", "code-reviewer",
		"--description", "Reviews code changes")
	require.NoError(t, err)
	assert.Contains(t, out, "similar", "should warn about overlap")

	// Both skills should validate
	for _, name := range []string{"code-review", "code-reviewer"} {
		out, err = runCmd(t, "skill", "validate", name, "--json")
		require.NoError(t, err)

		var result output.SkillValidateResult
		require.NoError(t, json.Unmarshal([]byte(out), &result))
		assert.True(t, result.Valid, "%s should be valid", name)
	}

	// Install both to a platform, verify both exist
	_, err = runCmd(t, "skill", "install", "code-review", "--platform", "claude-code")
	require.NoError(t, err)
	_, err = runCmd(t, "skill", "install", "code-reviewer", "--platform", "claude-code")
	require.NoError(t, err)

	// List dedup hints — text output should mention potential duplicates
	out, err = runCmd(t, "skill", "list")
	require.NoError(t, err)
	assert.Contains(t, out, "code-review")
	assert.Contains(t, out, "code-reviewer")
}
