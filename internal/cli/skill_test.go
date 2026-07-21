package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/devrimcavusoglu/skern/internal/output"
	"github.com/devrimcavusoglu/skern/internal/registry"
	"github.com/devrimcavusoglu/skern/internal/skill"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func requireParsedManifest(t *testing.T, path string) *skill.Skill {
	t.Helper()
	s, err := skill.ParseManifest(path)
	require.NoError(t, err, "parsing manifest %s", path)
	return s
}

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

func TestSkillValidate_HintsJSON(t *testing.T) {
	cc := testRegistry(t)

	// Default body is short (~8 words) — triggers body-too-short hint
	_, err := runCmd(t, cc, "skill", "create", "hint-skill", "--description", "A test skill", "--author", "alice")
	require.NoError(t, err)

	out, err := runCmd(t, cc, "skill", "validate", "hint-skill", "--json")
	require.NoError(t, err)

	var result output.SkillValidateResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	assert.True(t, result.Valid, "hints should not make skill invalid")
	assert.Equal(t, 0, result.Errors)
	assert.Greater(t, result.Hints, 0, "should have at least one hint")
}

func TestSkillValidate_HintsText(t *testing.T) {
	cc := testRegistry(t)

	_, err := runCmd(t, cc, "skill", "create", "hint-text", "--description", "A test skill", "--author", "alice")
	require.NoError(t, err)

	out, err := runCmd(t, cc, "skill", "validate", "hint-text")
	require.NoError(t, err)
	assert.Contains(t, out, "hint(s)")
	assert.Contains(t, out, "~")
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

// --from-template only accepts a directory; a file path must error with a
// message that points the user at the parent directory.
func TestSkillCreate_FromTemplate_FilePath_Errors(t *testing.T) {
	cc := testRegistry(t)

	tmplDir := t.TempDir()
	tmplPath := filepath.Join(tmplDir, "SKILL.md")
	require.NoError(t, os.WriteFile(tmplPath, []byte("---\nname: x\ndescription: y\nmetadata:\n  author:\n    name: a\n    type: human\n  version: \"0.1.0\"\n---\nbody"), 0o644))

	_, err := runCmd(t, cc, "skill", "create", "tmpl-skill", "--from-template", tmplPath)
	require.Error(t, err)
	msg := err.Error()
	assert.Contains(t, msg, "must point to a skill directory")
	assert.Contains(t, msg, "pass the parent directory instead")
}

func TestSkillCreate_FromTemplate_NotFound(t *testing.T) {
	cc := testRegistry(t)

	_, err := runCmd(t, cc, "skill", "create", "tmpl-fail", "--from-template", "/nonexistent/template-dir")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "reading template")
}

// Regression for #83: passing a skill *directory* as --from-template must
// copy sibling assets (references/, templates/, VENDORED.md, ...) into the
// new skill alongside SKILL.md.
func TestSkillCreate_FromTemplate_Directory_CopiesSiblings(t *testing.T) {
	cc := testRegistry(t)

	srcDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "SKILL.md"), []byte(`---
name: source-template
description: Use when copying siblings is required.
metadata:
  author:
    name: alice
    type: human
  version: "0.1.0"
---

## Overview

Source body.
`), 0o644))
	require.NoError(t, os.MkdirAll(filepath.Join(srcDir, "references"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "references", "architecture.md"), []byte("# Architecture\n"), 0o644))
	require.NoError(t, os.MkdirAll(filepath.Join(srcDir, "templates"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "templates", "example.txt"), []byte("Example.\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "VENDORED.md"), []byte("# Vendored\n"), 0o644))

	out, err := runCmd(t, cc, "skill", "create", "my-templated", "--from-template", srcDir, "--json")
	require.NoError(t, err)

	var result output.SkillCreateResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))

	// Frontmatter from the directory's SKILL.md is preserved (covers #82
	// in the directory mode too).
	parsed := requireParsedManifest(t, filepath.Join(result.Path, "SKILL.md"))
	assert.Equal(t, "my-templated", parsed.Name)
	assert.Equal(t, "Use when copying siblings is required.", parsed.Description)
	assert.Equal(t, "0.1.0", parsed.Metadata.Version)

	// Siblings copied verbatim.
	for _, rel := range []string{"references/architecture.md", "templates/example.txt", "VENDORED.md"} {
		_, err := os.Stat(filepath.Join(result.Path, rel))
		assert.NoError(t, err, "expected sibling %q to be copied", rel)
	}

	// Sibling content is byte-identical to the source.
	got, err := os.ReadFile(filepath.Join(result.Path, "references", "architecture.md"))
	require.NoError(t, err)
	assert.Equal(t, "# Architecture\n", string(got))
}

// Explicit CLI flags must override template-derived values; unset flags must
// leave template values intact.
func TestSkillCreate_FromTemplate_CLIOverridesTemplate(t *testing.T) {
	cc := testRegistry(t)

	srcDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "SKILL.md"), []byte(`---
name: source-template
description: Use when the template description applies.
tags:
  - inherited
metadata:
  author:
    name: alice
    type: human
  version: "0.1.0"
---

## Overview
body
`), 0o644))

	out, err := runCmd(t, cc, "skill", "create", "my-templated",
		"--from-template", srcDir,
		"--description", "Use when the CLI override applies.",
		"--tags", "override1,override2",
		"--json",
	)
	require.NoError(t, err)

	var result output.SkillCreateResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))

	parsed := requireParsedManifest(t, filepath.Join(result.Path, "SKILL.md"))
	assert.Equal(t, "Use when the CLI override applies.", parsed.Description, "CLI --description should win")
	assert.Equal(t, []string{"override1", "override2"}, parsed.Tags, "CLI --tags should win")
	// Author was not overridden — template value preserved.
	assert.Equal(t, "alice", parsed.Metadata.Author.Name)
	// Version not set on CLI — template value preserved.
	assert.Equal(t, "0.1.0", parsed.Metadata.Version)
}

// A directory passed via --from-template that lacks SKILL.md must surface
// a clear error rather than silently producing a default skill.
func TestSkillCreate_FromTemplate_DirectoryMissingManifest(t *testing.T) {
	cc := testRegistry(t)

	srcDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "README.md"), []byte("not a skill\n"), 0o644))

	_, err := runCmd(t, cc, "skill", "create", "my-templated", "--from-template", srcDir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "SKILL.md")
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

// --- skill edit ---

func TestSkillEdit_Description(t *testing.T) {
	cc := testRegistry(t)

	_, err := runCmd(t, cc, "skill", "create", "edit-desc",
		"--description", "Original description", "--author", "alice")
	require.NoError(t, err)

	out, err := runCmd(t, cc, "skill", "edit", "edit-desc",
		"--description", "Updated description", "--json")
	require.NoError(t, err)

	var result output.SkillEditResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	assert.Equal(t, "edit-desc", result.Name)
	assert.Contains(t, result.Updated, "description")

	// Verify the change persisted
	showOut, err := runCmd(t, cc, "skill", "show", "edit-desc", "--json")
	require.NoError(t, err)

	var showResult output.SkillResult
	require.NoError(t, json.Unmarshal([]byte(showOut), &showResult))
	assert.Equal(t, "Updated description", showResult.Description)
}

func TestSkillEdit_Version(t *testing.T) {
	cc := testRegistry(t)

	_, err := runCmd(t, cc, "skill", "create", "edit-ver",
		"--description", "A skill", "--author", "alice")
	require.NoError(t, err)

	_, err = runCmd(t, cc, "skill", "edit", "edit-ver", "--version", "2.0.0")
	require.NoError(t, err)

	showOut, err := runCmd(t, cc, "skill", "show", "edit-ver", "--json")
	require.NoError(t, err)

	var showResult output.SkillResult
	require.NoError(t, json.Unmarshal([]byte(showOut), &showResult))
	assert.Equal(t, "2.0.0", showResult.Version)
}

func TestSkillEdit_Author(t *testing.T) {
	cc := testRegistry(t)

	_, err := runCmd(t, cc, "skill", "create", "edit-auth",
		"--description", "A skill", "--author", "alice")
	require.NoError(t, err)

	_, err = runCmd(t, cc, "skill", "edit", "edit-auth",
		"--author", "bob", "--author-type", "agent", "--author-platform", "claude-code")
	require.NoError(t, err)

	showOut, err := runCmd(t, cc, "skill", "show", "edit-auth", "--json")
	require.NoError(t, err)

	var showResult output.SkillResult
	require.NoError(t, json.Unmarshal([]byte(showOut), &showResult))
	assert.Equal(t, "bob", showResult.Author.Name)
	assert.Equal(t, "agent", showResult.Author.Type)
	assert.Equal(t, "claude-code", showResult.Author.Platform)
}

func TestSkillEdit_ModifiedBy(t *testing.T) {
	cc := testRegistry(t)

	_, err := runCmd(t, cc, "skill", "create", "edit-mod",
		"--description", "A skill", "--author", "alice")
	require.NoError(t, err)

	editOut, err := runCmd(t, cc, "skill", "edit", "edit-mod",
		"--description", "New desc",
		"--modified-by", "claude", "--modified-by-type", "agent", "--modified-by-platform", "claude-code",
		"--json")
	require.NoError(t, err)

	var editResult output.SkillEditResult
	require.NoError(t, json.Unmarshal([]byte(editOut), &editResult))
	assert.Contains(t, editResult.Updated, "description")
	assert.Contains(t, editResult.Updated, "modified-by")

	showOut, err := runCmd(t, cc, "skill", "show", "edit-mod", "--json")
	require.NoError(t, err)

	var showResult output.SkillResult
	require.NoError(t, json.Unmarshal([]byte(showOut), &showResult))
	require.Len(t, showResult.ModifiedBy, 1)
	assert.Equal(t, "claude", showResult.ModifiedBy[0].Name)
	assert.Equal(t, "agent", showResult.ModifiedBy[0].Type)
	assert.Equal(t, "claude-code", showResult.ModifiedBy[0].Platform)
	assert.NotEmpty(t, showResult.ModifiedBy[0].Date)
}

func TestSkillEdit_NotFound(t *testing.T) {
	cc := testRegistry(t)

	_, err := runCmd(t, cc, "skill", "edit", "nonexistent", "--description", "x")
	assert.Error(t, err)
}

func TestSkillEdit_TextOutput(t *testing.T) {
	cc := testRegistry(t)

	_, err := runCmd(t, cc, "skill", "create", "edit-text",
		"--description", "A skill", "--author", "alice")
	require.NoError(t, err)

	out, err := runCmd(t, cc, "skill", "edit", "edit-text",
		"--description", "Better desc", "--version", "1.0.0")
	require.NoError(t, err)
	assert.Contains(t, out, "Updated")
	assert.Contains(t, out, "description")
	assert.Contains(t, out, "version")
}

// --- skill tags ---

func TestSkillCreate_WithTags(t *testing.T) {
	cc := testRegistry(t)

	_, err := runCmd(t, cc, "skill", "create", "tagged-skill",
		"--description", "A tagged skill",
		"--tags", "devops,testing",
		"--json")
	require.NoError(t, err)

	out, err := runCmd(t, cc, "skill", "show", "tagged-skill", "--json")
	require.NoError(t, err)

	var result output.SkillResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	assert.Equal(t, []string{"devops", "testing"}, result.Tags)
}

func TestSkillCreate_WithTags_Show(t *testing.T) {
	cc := testRegistry(t)

	_, err := runCmd(t, cc, "skill", "create", "tagged-text",
		"--description", "A tagged skill",
		"--tags", "ci,deploy")
	require.NoError(t, err)

	out, err := runCmd(t, cc, "skill", "show", "tagged-text")
	require.NoError(t, err)
	assert.Contains(t, out, "Tags:")
	assert.Contains(t, out, "ci")
	assert.Contains(t, out, "deploy")
}

func TestSkillList_FilterByTag(t *testing.T) {
	cc := testRegistry(t)

	_, err := runCmd(t, cc, "skill", "create", "tool-a",
		"--description", "Tool A", "--tags", "devops")
	require.NoError(t, err)

	_, err = runCmd(t, cc, "skill", "create", "tool-b",
		"--description", "Tool B", "--tags", "testing")
	require.NoError(t, err)

	_, err = runCmd(t, cc, "skill", "create", "tool-c",
		"--description", "Tool C", "--tags", "devops,testing")
	require.NoError(t, err)

	// Filter by devops — should get tool-a and tool-c
	out, err := runCmd(t, cc, "skill", "list", "--tag", "devops", "--json")
	require.NoError(t, err)

	var result output.SkillListResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	assert.Equal(t, 2, result.Count)

	names := map[string]bool{}
	for _, s := range result.Skills {
		names[s.Name] = true
	}
	assert.True(t, names["tool-a"])
	assert.True(t, names["tool-c"])
}

// --- categorical-tag filter (#96) ---

func TestParseCategoryFilters(t *testing.T) {
	tests := []struct {
		name    string
		raw     []string
		want    map[string][]string
		wantErr bool
	}{
		{name: "empty", raw: nil, want: map[string][]string{}},
		{name: "single", raw: []string{"lang:python"}, want: map[string][]string{"lang": {"python"}}},
		{
			name: "comma list folds into one namespace",
			raw:  []string{"lang:python,go"},
			want: map[string][]string{"lang": {"python", "go"}},
		},
		{
			name: "repeated same namespace accumulates",
			raw:  []string{"lang:python", "lang:go"},
			want: map[string][]string{"lang": {"python", "go"}},
		},
		{
			name: "distinct namespaces",
			raw:  []string{"lang:python", "topic:testing"},
			want: map[string][]string{"lang": {"python"}, "topic": {"testing"}},
		},
		{
			name: "lowercased",
			raw:  []string{"Lang:Python"},
			want: map[string][]string{"lang": {"python"}},
		},
		{
			name: "duplicate values deduped",
			raw:  []string{"lang:python,python", "lang:Python"},
			want: map[string][]string{"lang": {"python"}},
		},
		{name: "no colon", raw: []string{"python"}, wantErr: true},
		{name: "empty namespace", raw: []string{":python"}, wantErr: true},
		{name: "comma in category name", raw: []string{",lang:python"}, wantErr: true},
		{name: "empty value", raw: []string{"lang:"}, wantErr: true},
		{name: "empty value in comma list", raw: []string{"lang:python,"}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseCategoryFilters(tt.raw)
			if tt.wantErr {
				require.Error(t, err)
				var ve *ValidationError
				assert.ErrorAs(t, err, &ve, "malformed --category must be a ValidationError (exit code 2)")
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMatchesCategories(t *testing.T) {
	tests := []struct {
		name            string
		tags            []string
		filters         map[string][]string
		includeUntagged bool
		want            bool
	}{
		{name: "empty filter matches everything", tags: []string{"lang:go"}, filters: map[string][]string{}, want: true},
		{name: "single match", tags: []string{"lang:python"}, filters: map[string][]string{"lang": {"python"}}, want: true},
		{name: "single miss", tags: []string{"lang:go"}, filters: map[string][]string{"lang": {"python"}}, want: false},
		{
			name:    "OR within category",
			tags:    []string{"lang:go"},
			filters: map[string][]string{"lang": {"python", "go"}},
			want:    true,
		},
		{
			name:    "AND across categories satisfied",
			tags:    []string{"lang:python", "topic:testing"},
			filters: map[string][]string{"lang": {"python"}, "topic": {"testing"}},
			want:    true,
		},
		{
			name:    "AND across categories one missing value fails",
			tags:    []string{"lang:python", "topic:docs"},
			filters: map[string][]string{"lang": {"python"}, "topic": {"testing"}},
			want:    false,
		},
		{
			name:    "category absent fails by default",
			tags:    []string{"lang:python"},
			filters: map[string][]string{"topic": {"testing"}},
			want:    false,
		},
		{
			name:            "category absent passes with includeUntagged",
			tags:            []string{"lang:python"},
			filters:         map[string][]string{"topic": {"testing"}},
			includeUntagged: true,
			want:            true,
		},
		{
			name:            "includeUntagged still requires a present category to match",
			tags:            []string{"lang:go", "topic:docs"},
			filters:         map[string][]string{"lang": {"python"}, "topic": {"docs"}},
			includeUntagged: true,
			want:            false,
		},
		{
			name:    "zero tags fails by default",
			tags:    nil,
			filters: map[string][]string{"lang": {"python"}},
			want:    false,
		},
		{
			name:            "zero tags passes with includeUntagged",
			tags:            nil,
			filters:         map[string][]string{"lang": {"python"}},
			includeUntagged: true,
			want:            true,
		},
		{
			name:    "flat tag is not categorical",
			tags:    []string{"python", "lang:go"},
			filters: map[string][]string{"lang": {"python"}},
			want:    false,
		},
		{
			name:    "case-insensitive match",
			tags:    []string{"Lang:Python"},
			filters: map[string][]string{"lang": {"python"}},
			want:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, matchesCategories(tt.tags, tt.filters, tt.includeUntagged))
		})
	}
}

// hasTag and matchesCategories share one normalization convention:
// case-insensitive, surrounding whitespace ignored on stored tags.
func TestHasTag_TrimAndCase(t *testing.T) {
	assert.True(t, hasTag([]string{" Featured "}, "featured"))
	assert.True(t, hasTag([]string{"featured"}, " FEATURED "))
	assert.False(t, hasTag([]string{"feat"}, "featured"))
}

// TestSkillList_FilterByCategory covers the --json contract for the categorical
// filter end to end: OR within a category, AND across categories, and the
// strict-by-default untagged handling.
func TestSkillList_FilterByCategory(t *testing.T) {
	cc := testRegistry(t)

	mk := func(name, desc, tags string) {
		t.Helper()
		_, err := runCmd(t, cc, "skill", "create", name, "--description", desc, "--tags", tags)
		require.NoError(t, err)
	}
	mk("py-test", "Python testing", "lang:python,topic:testing")
	mk("py-docs", "Python docs", "lang:python,topic:docs")
	mk("go-test", "Go testing", "lang:go,topic:testing")
	mk("untyped", "No categories", "misc")

	listNames := func(t *testing.T, args ...string) map[string]bool {
		t.Helper()
		out, err := runCmd(t, cc, append([]string{"skill", "list"}, args...)...)
		require.NoError(t, err)
		var result output.SkillListResult
		require.NoError(t, json.Unmarshal([]byte(out), &result))
		assert.Equal(t, len(result.Skills), result.Count)
		names := map[string]bool{}
		for _, s := range result.Skills {
			names[s.Name] = true
		}
		return names
	}

	// Single category value.
	got := listNames(t, "--category", "lang:python", "--json")
	assert.Equal(t, map[string]bool{"py-test": true, "py-docs": true}, got)

	// OR within a category.
	got = listNames(t, "--category", "lang:python,go", "--json")
	assert.Equal(t, map[string]bool{"py-test": true, "py-docs": true, "go-test": true}, got)

	// AND across categories.
	got = listNames(t, "--category", "lang:python", "--category", "topic:testing", "--json")
	assert.Equal(t, map[string]bool{"py-test": true}, got)

	// Strict by default: a skill with no tag in the category is excluded.
	got = listNames(t, "--category", "topic:testing", "--json")
	assert.Equal(t, map[string]bool{"py-test": true, "go-test": true}, got)

	// --include-untagged: category-absent skills now match that category.
	got = listNames(t, "--category", "topic:testing", "--include-untagged", "--json")
	assert.Equal(t, map[string]bool{"py-test": true, "go-test": true, "untyped": true}, got)
}

// TestSkillList_FilterByCategory_TextOutput drives the text (non-JSON)
// rendering path with a category filter.
func TestSkillList_FilterByCategory_TextOutput(t *testing.T) {
	cc := testRegistry(t)
	_, err := runCmd(t, cc, "skill", "create", "py-skill", "--description", "Python skill", "--tags", "lang:python")
	require.NoError(t, err)
	_, err = runCmd(t, cc, "skill", "create", "go-skill", "--description", "Go skill", "--tags", "lang:go")
	require.NoError(t, err)

	out, err := runCmd(t, cc, "skill", "list", "--category", "lang:python")
	require.NoError(t, err)
	assert.Contains(t, out, "py-skill")
	assert.NotContains(t, out, "go-skill")
}

func TestSkillList_FilterByCategory_Invalid(t *testing.T) {
	cc := testRegistry(t)
	_, err := runCmd(t, cc, "skill", "create", "x", "--description", "X", "--tags", "lang:go")
	require.NoError(t, err)

	_, err = runCmd(t, cc, "skill", "list", "--category", "python", "--json")
	require.Error(t, err)
	var ve *ValidationError
	assert.ErrorAs(t, err, &ve)
}

// TestSkillList_TagAndCategory confirms --tag and --category compose (AND).
func TestSkillList_TagAndCategory(t *testing.T) {
	cc := testRegistry(t)
	_, err := runCmd(t, cc, "skill", "create", "a", "--description", "A", "--tags", "featured,lang:go")
	require.NoError(t, err)
	_, err = runCmd(t, cc, "skill", "create", "b", "--description", "B", "--tags", "lang:go")
	require.NoError(t, err)

	out, err := runCmd(t, cc, "skill", "list", "--tag", "featured", "--category", "lang:go", "--json")
	require.NoError(t, err)
	var result output.SkillListResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	require.Equal(t, 1, result.Count)
	assert.Equal(t, "a", result.Skills[0].Name)
}

// TestSkillCreate_InvalidTag enforces the tag charset at the create boundary:
// alphanumeric segments joined by hyphens, at most one category:value colon.
func TestSkillCreate_InvalidTag(t *testing.T) {
	cc := testRegistry(t)
	for _, bad := range []string{"my_tag", "my tag", "a:b:c", "-tag", "tag-", "c++"} {
		_, err := runCmd(t, cc, "skill", "create", "x", "--description", "X", "--tags", bad)
		require.Error(t, err, "tag %q should be rejected", bad)
		var ve *ValidationError
		assert.ErrorAs(t, err, &ve, "invalid tag must be a ValidationError (exit code 2)")
	}
}

// TestSkillList_HyphenatedTags confirms hyphenated tags work end to end through
// both filters, and that space after a comma in --tags is normalized away.
func TestSkillList_HyphenatedTags(t *testing.T) {
	cc := testRegistry(t)
	_, err := runCmd(t, cc, "skill", "create", "review-helper", "--description", "Review helper",
		"--tags", "topic:code-review, this-is-another")
	require.NoError(t, err)

	out, err := runCmd(t, cc, "skill", "list", "--category", "topic:code-review", "--tag", "this-is-another", "--json")
	require.NoError(t, err)
	var result output.SkillListResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	require.Equal(t, 1, result.Count)
	assert.Equal(t, "review-helper", result.Skills[0].Name)
	assert.Equal(t, []string{"topic:code-review", "this-is-another"}, result.Skills[0].Tags)
}

// TestSkillList_WithPlatforms verifies that --with-platforms enriches each
// skill entry with the list of platforms where the skill is currently
// installed, scoped to the registry skill's scope.
func TestSkillList_WithPlatforms(t *testing.T) {
	cc, _, _ := testRegistryWithDirs(t)
	home := t.TempDir()
	project := t.TempDir()
	withTestDetector(t, cc, home, project)

	_, err := runCmd(t, cc, "skill", "create", "wp-installed", "--description", "Test")
	require.NoError(t, err)
	_, err = runCmd(t, cc, "skill", "create", "wp-not-installed", "--description", "Test")
	require.NoError(t, err)

	_, err = runCmd(t, cc, "skill", "install", "wp-installed", "--platform", "claude-code")
	require.NoError(t, err)

	out, err := runCmd(t, cc, "skill", "list", "--scope", "user", "--with-platforms", "--json")
	require.NoError(t, err)

	var result output.SkillListResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))

	byName := map[string]output.SkillResult{}
	for _, s := range result.Skills {
		byName[s.Name] = s
	}

	assert.Equal(t, []string{"claude-code"}, byName["wp-installed"].InstalledOn)
	assert.Empty(t, byName["wp-not-installed"].InstalledOn,
		"uninstalled skill should report empty platform list")
}

// TestSkillList_WithoutWithPlatforms confirms InstalledOn stays nil/empty
// when --with-platforms is not requested. JSON consumers rely on this to
// distinguish "queried, none" from "not queried".
func TestSkillList_WithoutWithPlatforms(t *testing.T) {
	cc, _, _ := testRegistryWithDirs(t)
	home := t.TempDir()
	project := t.TempDir()
	withTestDetector(t, cc, home, project)

	_, err := runCmd(t, cc, "skill", "create", "no-flag-skill", "--description", "Test")
	require.NoError(t, err)
	_, err = runCmd(t, cc, "skill", "install", "no-flag-skill", "--platform", "claude-code")
	require.NoError(t, err)

	out, err := runCmd(t, cc, "skill", "list", "--scope", "user", "--json")
	require.NoError(t, err)

	var result output.SkillListResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	require.Len(t, result.Skills, 1)
	assert.Empty(t, result.Skills[0].InstalledOn)
}

// TestSkillUninstall_BatchAndCapacity covers batch uninstall plus the post-op
// capacity snapshot showing the count drop.
func TestSkillUninstall_BatchAndCapacity(t *testing.T) {
	cc, _, _ := testRegistryWithDirs(t)
	home := t.TempDir()
	project := t.TempDir()
	withTestDetector(t, cc, home, project)

	for _, n := range []string{"u-a", "u-b"} {
		_, err := runCmd(t, cc, "skill", "create", n, "--description", "Test")
		require.NoError(t, err)
		_, err = runCmd(t, cc, "skill", "install", n, "--platform", "claude-code")
		require.NoError(t, err)
	}

	out, err := runCmd(t, cc, "skill", "uninstall",
		"u-a", "u-b", "--platform", "claude-code", "--json")
	require.NoError(t, err)

	var result output.SkillUninstallResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	require.Len(t, result.Skills, 2)
	for _, s := range result.Skills {
		assert.True(t, s.Success)
	}
	require.NotNil(t, result.Capacity)
	assert.Equal(t, 0, result.Capacity.Installed)
}

func TestSkillSearch_FilterByTag(t *testing.T) {
	cc := testRegistry(t)

	_, err := runCmd(t, cc, "skill", "create", "code-lint",
		"--description", "Lint code", "--tags", "code-quality")
	require.NoError(t, err)

	_, err = runCmd(t, cc, "skill", "create", "code-format",
		"--description", "Format code", "--tags", "formatting")
	require.NoError(t, err)

	// Search "code" but filter by code-quality — should get only code-lint
	out, err := runCmd(t, cc, "skill", "search", "code", "--tag", "code-quality", "--json")
	require.NoError(t, err)

	var result output.SkillSearchResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	assert.Equal(t, 1, result.Count)
	assert.Equal(t, "code-lint", result.Results[0].Name)
}
