package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/devrimcavusoglu/skern/internal/output"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- skill diff: two registry skills ---

func TestSkillDiff_TwoSkills_Identical(t *testing.T) {
	cc := testRegistry(t)

	_, err := runCmd(t, cc, "skill", "create", "diff-a", "--description", "Same skill")
	require.NoError(t, err)
	_, err = runCmd(t, cc, "skill", "create", "diff-b", "--description", "Same skill")
	require.NoError(t, err)

	out, err := runCmd(t, cc, "skill", "diff", "diff-a", "diff-b", "--json")
	require.NoError(t, err)

	var result output.SkillDiffResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	// Names differ, so not fully identical
	assert.False(t, result.Identical)
	assert.Equal(t, "diff-a", result.LeftName)
	assert.Equal(t, "diff-b", result.RightName)

	// Should have a name field diff
	found := false
	for _, f := range result.Fields {
		if f.Field == "name" {
			found = true
			assert.Equal(t, "diff-a", f.Left)
			assert.Equal(t, "diff-b", f.Right)
		}
	}
	assert.True(t, found, "expected name field diff")
}

func TestSkillDiff_TwoSkills_DifferentMetadata(t *testing.T) {
	cc := testRegistry(t)

	_, err := runCmd(t, cc, "skill", "create", "meta-a",
		"--description", "First description",
		"--author", "alice", "--author-type", "human",
		"--tags", "devops")
	require.NoError(t, err)
	_, err = runCmd(t, cc, "skill", "create", "meta-b",
		"--description", "Second description",
		"--author", "bob", "--author-type", "agent",
		"--tags", "testing")
	require.NoError(t, err)

	out, err := runCmd(t, cc, "skill", "diff", "meta-a", "meta-b", "--json")
	require.NoError(t, err)

	var result output.SkillDiffResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	assert.False(t, result.Identical)

	fieldMap := make(map[string]output.FieldDiff)
	for _, f := range result.Fields {
		fieldMap[f.Field] = f
	}

	assert.Contains(t, fieldMap, "name")
	assert.Contains(t, fieldMap, "description")
	assert.Contains(t, fieldMap, "author.name")
	assert.Contains(t, fieldMap, "author.type")
	assert.Contains(t, fieldMap, "tags")
	assert.Equal(t, "First description", fieldMap["description"].Left)
	assert.Equal(t, "Second description", fieldMap["description"].Right)
	assert.Equal(t, "alice", fieldMap["author.name"].Left)
	assert.Equal(t, "bob", fieldMap["author.name"].Right)
}

func TestSkillDiff_TwoSkills_DifferentBody(t *testing.T) {
	cc, userDir, _ := testRegistryWithDirs(t)

	_, err := runCmd(t, cc, "skill", "create", "body-a", "--description", "Same desc", "--author", "alice")
	require.NoError(t, err)
	_, err = runCmd(t, cc, "skill", "create", "body-b", "--description", "Same desc", "--author", "alice")
	require.NoError(t, err)

	// Overwrite body-b SKILL.md with a different body
	skillMdPath := filepath.Join(userDir, "body-b", "SKILL.md")
	content := `---
name: body-b
description: Same desc
metadata:
  author:
    name: alice
    type: human
  version: "0.1.0"
---

## Custom Instructions

Do something different.
`
	require.NoError(t, os.WriteFile(skillMdPath, []byte(content), 0o644))

	out, err := runCmd(t, cc, "skill", "diff", "body-a", "body-b", "--json")
	require.NoError(t, err)

	var result output.SkillDiffResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	assert.False(t, result.Identical)
	assert.True(t, result.BodyDiff)
	assert.NotEmpty(t, result.LeftBody)
	assert.NotEmpty(t, result.RightBody)
	assert.NotEqual(t, result.LeftBody, result.RightBody)
}

func TestSkillDiff_TwoSkills_TextOutput(t *testing.T) {
	cc := testRegistry(t)

	_, err := runCmd(t, cc, "skill", "create", "text-a", "--description", "Desc A", "--author", "alice")
	require.NoError(t, err)
	_, err = runCmd(t, cc, "skill", "create", "text-b", "--description", "Desc B", "--author", "bob")
	require.NoError(t, err)

	out, err := runCmd(t, cc, "skill", "diff", "text-a", "text-b")
	require.NoError(t, err)
	assert.Contains(t, out, "Comparing")
	assert.Contains(t, out, "text-a")
	assert.Contains(t, out, "text-b")
	assert.Contains(t, out, "description")
	assert.Contains(t, out, "author.name")
}

func TestSkillDiff_TwoSkills_NotFound(t *testing.T) {
	cc := testRegistry(t)

	_, err := runCmd(t, cc, "skill", "create", "exists", "--description", "A skill")
	require.NoError(t, err)

	_, err = runCmd(t, cc, "skill", "diff", "exists", "nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "nonexistent")
}

// --- skill diff: registry vs platform ---

func TestSkillDiff_RegistryVsPlatform_Identical(t *testing.T) {
	cc, _, _ := testRegistryWithDirs(t)
	home := t.TempDir()
	project := t.TempDir()
	withTestDetector(t, cc, home, project)

	_, err := runCmd(t, cc, "skill", "create", "diff-install", "--description", "A test skill")
	require.NoError(t, err)

	_, err = runCmd(t, cc, "skill", "install", "diff-install", "--platform", "claude-code")
	require.NoError(t, err)

	out, err := runCmd(t, cc, "skill", "diff", "diff-install",
		"--platform", "claude-code", "--scope", "user", "--json")
	require.NoError(t, err)

	var result output.SkillDiffResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	assert.True(t, result.Identical)
	assert.Equal(t, "diff-install", result.LeftName)
	assert.Equal(t, "diff-install", result.RightName)
	assert.Contains(t, result.LeftSource, "registry")
	assert.Contains(t, result.RightSource, "platform")
}

func TestSkillDiff_RegistryVsPlatform_Drifted(t *testing.T) {
	cc, _, _ := testRegistryWithDirs(t)
	home := t.TempDir()
	project := t.TempDir()
	withTestDetector(t, cc, home, project)

	_, err := runCmd(t, cc, "skill", "create", "drift-skill", "--description", "Original")
	require.NoError(t, err)

	_, err = runCmd(t, cc, "skill", "install", "drift-skill", "--platform", "claude-code")
	require.NoError(t, err)

	// Modify the installed copy on the platform to simulate drift
	installedPath := filepath.Join(home, ".claude", "skills", "drift-skill", "SKILL.md")
	driftedContent := `---
name: drift-skill
description: Modified on platform
metadata:
  author:
    name: skern
    type: human
  version: "0.2.0"
---

## Modified Instructions

This was changed on the platform.
`
	require.NoError(t, os.WriteFile(installedPath, []byte(driftedContent), 0o644))

	out, err := runCmd(t, cc, "skill", "diff", "drift-skill",
		"--platform", "claude-code", "--scope", "user", "--json")
	require.NoError(t, err)

	var result output.SkillDiffResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	assert.False(t, result.Identical)
	assert.True(t, result.BodyDiff)

	fieldMap := make(map[string]output.FieldDiff)
	for _, f := range result.Fields {
		fieldMap[f.Field] = f
	}

	assert.Contains(t, fieldMap, "description")
	assert.Contains(t, fieldMap, "version")
}

func TestSkillDiff_RegistryVsPlatform_TextOutput(t *testing.T) {
	cc, _, _ := testRegistryWithDirs(t)
	home := t.TempDir()
	project := t.TempDir()
	withTestDetector(t, cc, home, project)

	_, err := runCmd(t, cc, "skill", "create", "text-diff", "--description", "A skill")
	require.NoError(t, err)

	_, err = runCmd(t, cc, "skill", "install", "text-diff", "--platform", "claude-code")
	require.NoError(t, err)

	out, err := runCmd(t, cc, "skill", "diff", "text-diff",
		"--platform", "claude-code", "--scope", "user")
	require.NoError(t, err)
	assert.Contains(t, out, "identical")
}

func TestSkillDiff_RegistryVsPlatform_NotInstalled(t *testing.T) {
	cc, _, _ := testRegistryWithDirs(t)
	home := t.TempDir()
	project := t.TempDir()
	withTestDetector(t, cc, home, project)

	_, err := runCmd(t, cc, "skill", "create", "not-installed", "--description", "A skill")
	require.NoError(t, err)

	_, err = runCmd(t, cc, "skill", "diff", "not-installed",
		"--platform", "claude-code", "--scope", "user")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not installed")
}

func TestSkillDiff_MissingPlatformFlag(t *testing.T) {
	cc := testRegistry(t)

	_, err := runCmd(t, cc, "skill", "create", "missing-plat", "--description", "A skill")
	require.NoError(t, err)

	_, err = runCmd(t, cc, "skill", "diff", "missing-plat")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--platform")
}

func TestSkillDiff_PlatformAll_Rejected(t *testing.T) {
	cc, _, _ := testRegistryWithDirs(t)
	home := t.TempDir()
	project := t.TempDir()
	withTestDetector(t, cc, home, project)

	_, err := runCmd(t, cc, "skill", "create", "all-plat", "--description", "A skill")
	require.NoError(t, err)

	_, err = runCmd(t, cc, "skill", "diff", "all-plat",
		"--platform", "all", "--scope", "user")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "specific platform")
}

func TestSkillDiff_NoArgs(t *testing.T) {
	cc := testRegistry(t)

	_, err := runCmd(t, cc, "skill", "diff")
	assert.Error(t, err)
}

func TestSkillDiff_RegistryVsPlatform_DefaultScope(t *testing.T) {
	cc, _, _ := testRegistryWithDirs(t)
	home := t.TempDir()
	project := t.TempDir()
	withTestDetector(t, cc, home, project)

	_, err := runCmd(t, cc, "skill", "create", "default-scope", "--description", "A skill")
	require.NoError(t, err)

	_, err = runCmd(t, cc, "skill", "install", "default-scope", "--platform", "claude-code")
	require.NoError(t, err)

	// Omit --scope; should default to user
	out, err := runCmd(t, cc, "skill", "diff", "default-scope",
		"--platform", "claude-code", "--json")
	require.NoError(t, err)

	var result output.SkillDiffResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	assert.True(t, result.Identical)
}
