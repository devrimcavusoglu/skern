package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/devrimcavusoglu/skern/internal/output"
	"github.com/devrimcavusoglu/skern/internal/registry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testRegistry returns a CommandContext with a temp registry.
func testRegistry(t *testing.T) *CommandContext {
	t.Helper()
	userDir := filepath.Join(t.TempDir(), "user-skills")
	projectDir := filepath.Join(t.TempDir(), "project-skills")

	return &CommandContext{
		NewRegistry: func() (*registry.Registry, error) {
			return registry.New(userDir, projectDir), nil
		},
		NewDetector: defaultNewDetector,
	}
}

func runCmd(t *testing.T, cc *CommandContext, args ...string) (string, error) {
	t.Helper()
	cmd := newRootCmd(cc)
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return buf.String(), err
}

// --- skill create ---

func TestSkillCreate(t *testing.T) {
	cc := testRegistry(t)

	_, err := runCmd(t, cc, "skill", "create", "my-skill", "--description", "A test skill")
	require.NoError(t, err)
}

func TestSkillCreate_JSON(t *testing.T) {
	cc := testRegistry(t)

	out, err := runCmd(t, cc, "skill", "create", "my-skill", "--json")
	require.NoError(t, err)

	var result output.SkillCreateResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	assert.Equal(t, "my-skill", result.Name)
	assert.Equal(t, "user", result.Scope)
	assert.NotEmpty(t, result.Path)
}

func TestSkillCreate_ProjectScope(t *testing.T) {
	cc := testRegistry(t)

	out, err := runCmd(t, cc, "skill", "create", "proj-skill", "--scope", "project", "--json")
	require.NoError(t, err)

	var result output.SkillCreateResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	assert.Equal(t, "project", result.Scope)
}

func TestSkillCreate_InvalidName(t *testing.T) {
	cc := testRegistry(t)

	_, err := runCmd(t, cc, "skill", "create", "INVALID_NAME")
	assert.Error(t, err)
}

func TestSkillCreate_Duplicate(t *testing.T) {
	cc := testRegistry(t)

	_, err := runCmd(t, cc, "skill", "create", "dup-skill")
	require.NoError(t, err)

	_, err = runCmd(t, cc, "skill", "create", "dup-skill")
	assert.Error(t, err)
}

func TestSkillCreate_WithAuthor(t *testing.T) {
	cc := testRegistry(t)

	out, err := runCmd(t, cc, "skill", "create", "authored-skill",
		"--author", "alice", "--author-type", "human",
		"--json")
	require.NoError(t, err)

	var result output.SkillCreateResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	assert.Equal(t, "authored-skill", result.Name)
}

// --- skill list ---

func TestSkillList_Empty(t *testing.T) {
	cc := testRegistry(t)

	out, err := runCmd(t, cc, "skill", "list", "--json")
	require.NoError(t, err)

	var result output.SkillListResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	assert.Equal(t, 0, result.Count)
}

func TestSkillList(t *testing.T) {
	cc := testRegistry(t)

	_, err := runCmd(t, cc, "skill", "create", "skill-a")
	require.NoError(t, err)
	_, err = runCmd(t, cc, "skill", "create", "skill-b")
	require.NoError(t, err)

	out, err := runCmd(t, cc, "skill", "list", "--json")
	require.NoError(t, err)

	var result output.SkillListResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	assert.Equal(t, 2, result.Count)
}

func TestSkillList_Scoped(t *testing.T) {
	cc := testRegistry(t)

	_, err := runCmd(t, cc, "skill", "create", "user-skill", "--scope", "user")
	require.NoError(t, err)
	_, err = runCmd(t, cc, "skill", "create", "proj-skill", "--scope", "project")
	require.NoError(t, err)

	out, err := runCmd(t, cc, "skill", "list", "--scope", "user", "--json")
	require.NoError(t, err)

	var result output.SkillListResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	assert.Equal(t, 1, result.Count)
	assert.Equal(t, "user-skill", result.Skills[0].Name)
}

func TestSkillList_Text(t *testing.T) {
	cc := testRegistry(t)

	_, err := runCmd(t, cc, "skill", "create", "my-skill")
	require.NoError(t, err)

	out, err := runCmd(t, cc, "skill", "list")
	require.NoError(t, err)
	assert.Contains(t, out, "my-skill")
}

// --- skill show ---

func TestSkillShow(t *testing.T) {
	cc := testRegistry(t)

	_, err := runCmd(t, cc, "skill", "create", "show-skill", "--description", "Show me")
	require.NoError(t, err)

	out, err := runCmd(t, cc, "skill", "show", "show-skill", "--json")
	require.NoError(t, err)

	var result output.SkillResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	assert.Equal(t, "show-skill", result.Name)
	assert.Equal(t, "Show me", result.Description)
}

func TestSkillShow_NotFound(t *testing.T) {
	cc := testRegistry(t)

	_, err := runCmd(t, cc, "skill", "show", "nonexistent")
	assert.Error(t, err)
}

func TestSkillShow_Text(t *testing.T) {
	cc := testRegistry(t)

	_, err := runCmd(t, cc, "skill", "create", "detail-skill", "--description", "Detailed info")
	require.NoError(t, err)

	out, err := runCmd(t, cc, "skill", "show", "detail-skill")
	require.NoError(t, err)
	assert.Contains(t, out, "detail-skill")
	assert.Contains(t, out, "Detailed info")
}

// --- skill search ---

func TestSkillSearch(t *testing.T) {
	cc := testRegistry(t)

	_, err := runCmd(t, cc, "skill", "create", "code-review")
	require.NoError(t, err)
	_, err = runCmd(t, cc, "skill", "create", "code-format")
	require.NoError(t, err)
	_, err = runCmd(t, cc, "skill", "create", "deploy-app")
	require.NoError(t, err)

	out, err := runCmd(t, cc, "skill", "search", "code", "--json")
	require.NoError(t, err)

	var result output.SkillSearchResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	assert.Equal(t, "code", result.Query)
	assert.Equal(t, 2, result.Count)
}

func TestSkillSearch_NoMatch(t *testing.T) {
	cc := testRegistry(t)

	_, err := runCmd(t, cc, "skill", "create", "my-skill")
	require.NoError(t, err)

	out, err := runCmd(t, cc, "skill", "search", "nonexistent", "--json")
	require.NoError(t, err)

	var result output.SkillSearchResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	assert.Equal(t, 0, result.Count)
}

func TestSkillSearch_Text(t *testing.T) {
	cc := testRegistry(t)

	_, err := runCmd(t, cc, "skill", "create", "find-me")
	require.NoError(t, err)

	out, err := runCmd(t, cc, "skill", "search", "find")
	require.NoError(t, err)
	assert.Contains(t, out, "find-me")
}

// --- skill remove ---

func TestSkillRemove(t *testing.T) {
	cc := testRegistry(t)

	_, err := runCmd(t, cc, "skill", "create", "remove-me")
	require.NoError(t, err)

	out, err := runCmd(t, cc, "skill", "remove", "remove-me", "--json")
	require.NoError(t, err)

	var result output.SkillRemoveResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	assert.Equal(t, "remove-me", result.Name)
	assert.Equal(t, "user", result.Scope)

	// Verify it's gone
	_, err = runCmd(t, cc, "skill", "show", "remove-me")
	assert.Error(t, err)
}

func TestSkillRemove_NotFound(t *testing.T) {
	cc := testRegistry(t)

	_, err := runCmd(t, cc, "skill", "remove", "nonexistent")
	assert.Error(t, err)
}

func TestSkillRemove_Text(t *testing.T) {
	cc := testRegistry(t)

	_, err := runCmd(t, cc, "skill", "create", "bye-skill")
	require.NoError(t, err)

	out, err := runCmd(t, cc, "skill", "remove", "bye-skill")
	require.NoError(t, err)
	assert.Contains(t, out, "Removed")
	assert.Contains(t, out, "bye-skill")
}

func TestSkillRemove_InvalidName(t *testing.T) {
	cc := testRegistry(t)

	_, err := runCmd(t, cc, "skill", "remove", "INVALID")
	assert.Error(t, err)
}

// --- skill validate ---

func TestSkillValidate(t *testing.T) {
	cc := testRegistry(t)

	_, err := runCmd(t, cc, "skill", "create", "valid-skill", "--description", "A valid skill", "--author", "alice")
	require.NoError(t, err)

	out, err := runCmd(t, cc, "skill", "validate", "valid-skill", "--json")
	require.NoError(t, err)

	var result output.SkillValidateResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	assert.Equal(t, "valid-skill", result.Name)
	assert.True(t, result.Valid)
	assert.Equal(t, 0, result.Errors)
}

func TestSkillValidate_NotFound(t *testing.T) {
	cc := testRegistry(t)

	_, err := runCmd(t, cc, "skill", "validate", "nonexistent")
	assert.Error(t, err)
}

func TestSkillValidate_Text(t *testing.T) {
	cc := testRegistry(t)

	_, err := runCmd(t, cc, "skill", "create", "text-skill", "--description", "A skill", "--author", "alice")
	require.NoError(t, err)

	out, err := runCmd(t, cc, "skill", "validate", "text-skill")
	require.NoError(t, err)
	assert.Contains(t, out, "valid")
}

// --- skill create with overlap ---

func TestSkillCreate_OverlapWarn(t *testing.T) {
	cc := testRegistry(t)

	// Create first skill
	_, err := runCmd(t, cc, "skill", "create", "code-review", "--description", "Reviews code")
	require.NoError(t, err)

	// Create similar skill — should succeed with warning
	out, err := runCmd(t, cc, "skill", "create", "code-reviewer", "--description", "Reviews code changes")
	require.NoError(t, err)
	assert.Contains(t, out, "similar")
}

func TestSkillCreate_Force(t *testing.T) {
	cc := testRegistry(t)

	_, err := runCmd(t, cc, "skill", "create", "my-tool", "--description", "Does things")
	require.NoError(t, err)

	// Even with high overlap, --force should allow creation
	_, err = runCmd(t, cc, "skill", "create", "my-tools", "--description", "Does things", "--force")
	require.NoError(t, err)
}

// --- Validation error exit code ---

func TestValidationError_ExitCode(t *testing.T) {
	// Execute() returns exit code 2 for validation errors
	// We test via the error type directly
	ve := &ValidationError{Message: "test"}
	assert.Equal(t, "test", ve.Error())
}

// --- completion ---

func TestCompletion_Bash(t *testing.T) {
	out, err := runCmd(t, nil, "completion", "bash")
	require.NoError(t, err)
	assert.NotEmpty(t, out)
	assert.Contains(t, out, "bash")
}

func TestCompletion_Zsh(t *testing.T) {
	out, err := runCmd(t, nil, "completion", "zsh")
	require.NoError(t, err)
	assert.NotEmpty(t, out)
}

func TestCompletion_Fish(t *testing.T) {
	out, err := runCmd(t, nil, "completion", "fish")
	require.NoError(t, err)
	assert.NotEmpty(t, out)
}

func TestCompletion_Invalid(t *testing.T) {
	_, err := runCmd(t, nil, "completion", "powershell")
	assert.Error(t, err)
}

// --- from-template ---

func TestSkillCreate_FromTemplate(t *testing.T) {
	cc := testRegistry(t)

	// Write a template file
	tmplDir := t.TempDir()
	tmplPath := filepath.Join(tmplDir, "template.md")
	require.NoError(t, os.WriteFile(tmplPath, []byte("## Custom Instructions\n\nDo something custom.\n"), 0o644))

	out, err := runCmd(t, cc, "skill", "create", "tmpl-skill", "--from-template", tmplPath, "--json")
	require.NoError(t, err)

	var result output.SkillCreateResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	assert.Equal(t, "tmpl-skill", result.Name)

	// Verify the body was used by reading the created SKILL.md
	skillMd, err := os.ReadFile(filepath.Join(result.Path, "SKILL.md"))
	require.NoError(t, err)
	assert.Contains(t, string(skillMd), "Custom Instructions")
	assert.Contains(t, string(skillMd), "Do something custom")
}

func TestSkillCreate_FromTemplate_NotFound(t *testing.T) {
	cc := testRegistry(t)

	_, err := runCmd(t, cc, "skill", "create", "tmpl-fail", "--from-template", "/nonexistent/template.md")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "reading template")
}

// --- dedup hints in list ---

func TestSkillList_DedupHints(t *testing.T) {
	cc := testRegistry(t)

	// Create two similar skills
	_, err := runCmd(t, cc, "skill", "create", "code-review", "--description", "Reviews code")
	require.NoError(t, err)
	_, err = runCmd(t, cc, "skill", "create", "code-reviewer", "--description", "Reviews code changes", "--force")
	require.NoError(t, err)

	out, err := runCmd(t, cc, "skill", "list", "--json")
	require.NoError(t, err)

	var result output.SkillListResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	assert.Equal(t, 2, result.Count)
	assert.NotEmpty(t, result.Duplicates, "should have duplicate hints for similar skills")
	assert.Equal(t, "code-review", result.Duplicates[0].SkillA)
	assert.Equal(t, "code-reviewer", result.Duplicates[0].SkillB)
}

func TestSkillList_DedupHints_Text(t *testing.T) {
	cc := testRegistry(t)

	_, err := runCmd(t, cc, "skill", "create", "code-review", "--description", "Reviews code")
	require.NoError(t, err)
	_, err = runCmd(t, cc, "skill", "create", "code-reviewer", "--description", "Reviews code changes", "--force")
	require.NoError(t, err)

	out, err := runCmd(t, cc, "skill", "list")
	require.NoError(t, err)
	assert.Contains(t, out, "Potential duplicates")
	assert.Contains(t, out, "code-review")
	assert.Contains(t, out, "code-reviewer")
}

func TestSkillList_NoDedupHints(t *testing.T) {
	cc := testRegistry(t)

	_, err := runCmd(t, cc, "skill", "create", "alpha-skill", "--description", "Does alpha things")
	require.NoError(t, err)
	_, err = runCmd(t, cc, "skill", "create", "zeta-deploy", "--description", "Deploys to production")
	require.NoError(t, err)

	out, err := runCmd(t, cc, "skill", "list", "--json")
	require.NoError(t, err)

	var result output.SkillListResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	assert.Equal(t, 2, result.Count)
	assert.Empty(t, result.Duplicates, "should have no duplicate hints for dissimilar skills")
}

// --- skill list parse warnings ---

func TestSkillList_ParseWarnings_JSON(t *testing.T) {
	cc, userDir, _ := testRegistryWithDirs(t)

	// Create a valid skill
	_, err := runCmd(t, cc, "skill", "create", "good-skill", "--description", "Works fine")
	require.NoError(t, err)

	// Create a corrupt skill directory (no SKILL.md)
	require.NoError(t, os.MkdirAll(filepath.Join(userDir, "broken-skill"), 0o755))

	out, err := runCmd(t, cc, "skill", "list", "--json")
	require.NoError(t, err)

	var result output.SkillListResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	assert.Equal(t, 1, result.Count)
	require.Len(t, result.ParseWarnings, 1)
	assert.Equal(t, "broken-skill", result.ParseWarnings[0].Name)
	assert.NotEmpty(t, result.ParseWarnings[0].Error)
}

func TestSkillList_ParseWarnings_Text(t *testing.T) {
	cc, userDir, _ := testRegistryWithDirs(t)

	_, err := runCmd(t, cc, "skill", "create", "ok-skill", "--description", "Works fine")
	require.NoError(t, err)

	require.NoError(t, os.MkdirAll(filepath.Join(userDir, "bad-skill"), 0o755))

	out, err := runCmd(t, cc, "skill", "list")
	require.NoError(t, err)
	assert.Contains(t, out, "could not be parsed")
	assert.Contains(t, out, "bad-skill")
}

// --- author provenance (modified-by) ---

// --- skill show with files ---

func TestSkillShow_WithFiles(t *testing.T) {
	cc := testRegistry(t)

	_, err := runCmd(t, cc, "skill", "create", "file-skill", "--description", "A skill with files")
	require.NoError(t, err)

	// Get the skill path
	showOut, err := runCmd(t, cc, "skill", "show", "file-skill", "--json")
	require.NoError(t, err)

	var initial output.SkillResult
	require.NoError(t, json.Unmarshal([]byte(showOut), &initial))

	// Add extra files to the skill directory
	scriptsDir := filepath.Join(initial.Path, "scripts")
	require.NoError(t, os.MkdirAll(scriptsDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(scriptsDir, "convert.py"), []byte("# python"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(initial.Path, "config.json"), []byte("{}"), 0o644))

	// Show again — should include files
	out, err := runCmd(t, cc, "skill", "show", "file-skill", "--json")
	require.NoError(t, err)

	var result output.SkillResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	assert.Len(t, result.Files, 2)
	assert.Contains(t, result.Files, "config.json")
	assert.Contains(t, result.Files, filepath.Join("scripts", "convert.py"))
}

// --- skill validate folder warning ---

func TestSkillValidate_FolderWarning(t *testing.T) {
	cc := testRegistry(t)

	_, err := runCmd(t, cc, "skill", "create", "ref-skill", "--description", "A skill with refs", "--author", "alice")
	require.NoError(t, err)

	// Get the skill path
	showOut, err := runCmd(t, cc, "skill", "show", "ref-skill", "--json")
	require.NoError(t, err)

	var initial output.SkillResult
	require.NoError(t, json.Unmarshal([]byte(showOut), &initial))

	// Overwrite SKILL.md with a body referencing a missing file
	skillMdPath := filepath.Join(initial.Path, "SKILL.md")
	content := `---
name: ref-skill
description: A skill with refs
metadata:
  author:
    name: alice
    type: human
  version: "0.1.0"
---
## Instructions

Run ` + "`scripts/run.py`" + ` to process data.
`
	require.NoError(t, os.WriteFile(skillMdPath, []byte(content), 0o644))

	// Validate — should warn about missing file
	out, err := runCmd(t, cc, "skill", "validate", "ref-skill", "--json")
	require.NoError(t, err)

	var result output.SkillValidateResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	assert.True(t, result.Valid, "should still be valid (warnings only)")
	assert.Equal(t, 1, result.Warns)

	// Find the folder warning
	found := false
	for _, issue := range result.Issues {
		if issue.Field == "folder" {
			found = true
			assert.Equal(t, "warning", issue.Severity)
			assert.Contains(t, issue.Message, "scripts/run.py")
		}
	}
	assert.True(t, found, "should have a folder warning")
}

func TestSkillShow_ModifiedBy(t *testing.T) {
	cc := testRegistry(t)

	// Create a skill and manually add modified-by entries
	_, err := runCmd(t, cc, "skill", "create", "prov-skill", "--description", "Provenance test", "--author", "alice")
	require.NoError(t, err)

	// Read the created SKILL.md and add modified-by to the frontmatter
	showOut, err := runCmd(t, cc, "skill", "show", "prov-skill", "--json")
	require.NoError(t, err)

	var result output.SkillResult
	require.NoError(t, json.Unmarshal([]byte(showOut), &result))

	// Write modified SKILL.md with modified-by
	skillMdPath := filepath.Join(result.Path, "SKILL.md")
	modifiedContent := `---
name: prov-skill
description: Provenance test
metadata:
  author:
    name: alice
    type: human
  version: "0.1.0"
  modified-by:
    - name: bob
      type: agent
      platform: claude-code
      date: "2025-01-15"
    - name: carol
      type: human
      date: "2025-02-01"
---
## Instructions

TODO: Add step-by-step instructions for the agent.
`
	require.NoError(t, os.WriteFile(skillMdPath, []byte(modifiedContent), 0o644))

	// Show the skill — JSON should include modified_by
	out, err := runCmd(t, cc, "skill", "show", "prov-skill", "--json")
	require.NoError(t, err)

	var updated output.SkillResult
	require.NoError(t, json.Unmarshal([]byte(out), &updated))
	require.Len(t, updated.ModifiedBy, 2)
	assert.Equal(t, "bob", updated.ModifiedBy[0].Name)
	assert.Equal(t, "agent", updated.ModifiedBy[0].Type)
	assert.Equal(t, "claude-code", updated.ModifiedBy[0].Platform)
	assert.Equal(t, "2025-01-15", updated.ModifiedBy[0].Date)
	assert.Equal(t, "carol", updated.ModifiedBy[1].Name)
}

func TestSkillShow_ModifiedBy_Text(t *testing.T) {
	cc := testRegistry(t)

	_, err := runCmd(t, cc, "skill", "create", "prov-text", "--description", "Provenance text test", "--author", "alice")
	require.NoError(t, err)

	showOut, err := runCmd(t, cc, "skill", "show", "prov-text", "--json")
	require.NoError(t, err)

	var result output.SkillResult
	require.NoError(t, json.Unmarshal([]byte(showOut), &result))

	skillMdPath := filepath.Join(result.Path, "SKILL.md")
	modifiedContent := `---
name: prov-text
description: Provenance text test
metadata:
  author:
    name: alice
    type: human
  version: "0.1.0"
  modified-by:
    - name: bob
      type: agent
      platform: claude-code
      date: "2025-01-15"
---
## Instructions

TODO: Add step-by-step instructions for the agent.
`
	require.NoError(t, os.WriteFile(skillMdPath, []byte(modifiedContent), 0o644))

	out, err := runCmd(t, cc, "skill", "show", "prov-text")
	require.NoError(t, err)
	assert.Contains(t, out, "Modified-by")
	assert.Contains(t, out, "bob")
	assert.Contains(t, out, "agent")
	assert.Contains(t, out, "claude-code")
	assert.Contains(t, out, "2025-01-15")
}
