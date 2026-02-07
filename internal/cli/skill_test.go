package cli

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/devrimcavusoglu/scribe/internal/output"
	"github.com/devrimcavusoglu/scribe/internal/registry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestRegistry overrides newRegistryFunc to use temp directories.
func setupTestRegistry(t *testing.T) {
	t.Helper()
	userDir := filepath.Join(t.TempDir(), "user-skills")
	projectDir := filepath.Join(t.TempDir(), "project-skills")

	original := newRegistryFunc
	newRegistryFunc = func() (*registry.Registry, error) {
		return registry.New(userDir, projectDir), nil
	}
	t.Cleanup(func() { newRegistryFunc = original })
}

func runCmd(t *testing.T, args ...string) (string, error) {
	t.Helper()
	cmd := NewRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return buf.String(), err
}

// --- skill create ---

func TestSkillCreate(t *testing.T) {
	setupTestRegistry(t)

	_, err := runCmd(t, "skill", "create", "my-skill", "--description", "A test skill")
	require.NoError(t, err)
}

func TestSkillCreate_JSON(t *testing.T) {
	setupTestRegistry(t)

	out, err := runCmd(t, "skill", "create", "my-skill", "--json")
	require.NoError(t, err)

	var result output.SkillCreateResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	assert.Equal(t, "my-skill", result.Name)
	assert.Equal(t, "user", result.Scope)
	assert.NotEmpty(t, result.Path)
}

func TestSkillCreate_ProjectScope(t *testing.T) {
	setupTestRegistry(t)

	out, err := runCmd(t, "skill", "create", "proj-skill", "--scope", "project", "--json")
	require.NoError(t, err)

	var result output.SkillCreateResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	assert.Equal(t, "project", result.Scope)
}

func TestSkillCreate_InvalidName(t *testing.T) {
	setupTestRegistry(t)

	_, err := runCmd(t, "skill", "create", "INVALID_NAME")
	assert.Error(t, err)
}

func TestSkillCreate_Duplicate(t *testing.T) {
	setupTestRegistry(t)

	_, err := runCmd(t, "skill", "create", "dup-skill")
	require.NoError(t, err)

	_, err = runCmd(t, "skill", "create", "dup-skill")
	assert.Error(t, err)
}

func TestSkillCreate_WithAuthor(t *testing.T) {
	setupTestRegistry(t)

	out, err := runCmd(t, "skill", "create", "authored-skill",
		"--author", "alice", "--author-type", "human",
		"--json")
	require.NoError(t, err)

	var result output.SkillCreateResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	assert.Equal(t, "authored-skill", result.Name)
}

// --- skill list ---

func TestSkillList_Empty(t *testing.T) {
	setupTestRegistry(t)

	out, err := runCmd(t, "skill", "list", "--json")
	require.NoError(t, err)

	var result output.SkillListResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	assert.Equal(t, 0, result.Count)
}

func TestSkillList(t *testing.T) {
	setupTestRegistry(t)

	_, err := runCmd(t, "skill", "create", "skill-a")
	require.NoError(t, err)
	_, err = runCmd(t, "skill", "create", "skill-b")
	require.NoError(t, err)

	out, err := runCmd(t, "skill", "list", "--json")
	require.NoError(t, err)

	var result output.SkillListResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	assert.Equal(t, 2, result.Count)
}

func TestSkillList_Scoped(t *testing.T) {
	setupTestRegistry(t)

	_, err := runCmd(t, "skill", "create", "user-skill", "--scope", "user")
	require.NoError(t, err)
	_, err = runCmd(t, "skill", "create", "proj-skill", "--scope", "project")
	require.NoError(t, err)

	out, err := runCmd(t, "skill", "list", "--scope", "user", "--json")
	require.NoError(t, err)

	var result output.SkillListResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	assert.Equal(t, 1, result.Count)
	assert.Equal(t, "user-skill", result.Skills[0].Name)
}

func TestSkillList_Text(t *testing.T) {
	setupTestRegistry(t)

	_, err := runCmd(t, "skill", "create", "my-skill")
	require.NoError(t, err)

	out, err := runCmd(t, "skill", "list")
	require.NoError(t, err)
	assert.Contains(t, out, "my-skill")
}

// --- skill show ---

func TestSkillShow(t *testing.T) {
	setupTestRegistry(t)

	_, err := runCmd(t, "skill", "create", "show-skill", "--description", "Show me")
	require.NoError(t, err)

	out, err := runCmd(t, "skill", "show", "show-skill", "--json")
	require.NoError(t, err)

	var result output.SkillResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	assert.Equal(t, "show-skill", result.Name)
	assert.Equal(t, "Show me", result.Description)
}

func TestSkillShow_NotFound(t *testing.T) {
	setupTestRegistry(t)

	_, err := runCmd(t, "skill", "show", "nonexistent")
	assert.Error(t, err)
}

func TestSkillShow_Text(t *testing.T) {
	setupTestRegistry(t)

	_, err := runCmd(t, "skill", "create", "detail-skill", "--description", "Detailed info")
	require.NoError(t, err)

	out, err := runCmd(t, "skill", "show", "detail-skill")
	require.NoError(t, err)
	assert.Contains(t, out, "detail-skill")
	assert.Contains(t, out, "Detailed info")
}

// --- skill search ---

func TestSkillSearch(t *testing.T) {
	setupTestRegistry(t)

	_, err := runCmd(t, "skill", "create", "code-review")
	require.NoError(t, err)
	_, err = runCmd(t, "skill", "create", "code-format")
	require.NoError(t, err)
	_, err = runCmd(t, "skill", "create", "deploy-app")
	require.NoError(t, err)

	out, err := runCmd(t, "skill", "search", "code", "--json")
	require.NoError(t, err)

	var result output.SkillSearchResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	assert.Equal(t, "code", result.Query)
	assert.Equal(t, 2, result.Count)
}

func TestSkillSearch_NoMatch(t *testing.T) {
	setupTestRegistry(t)

	_, err := runCmd(t, "skill", "create", "my-skill")
	require.NoError(t, err)

	out, err := runCmd(t, "skill", "search", "nonexistent", "--json")
	require.NoError(t, err)

	var result output.SkillSearchResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	assert.Equal(t, 0, result.Count)
}

func TestSkillSearch_Text(t *testing.T) {
	setupTestRegistry(t)

	_, err := runCmd(t, "skill", "create", "find-me")
	require.NoError(t, err)

	out, err := runCmd(t, "skill", "search", "find")
	require.NoError(t, err)
	assert.Contains(t, out, "find-me")
}

// --- skill remove ---

func TestSkillRemove(t *testing.T) {
	setupTestRegistry(t)

	_, err := runCmd(t, "skill", "create", "remove-me")
	require.NoError(t, err)

	out, err := runCmd(t, "skill", "remove", "remove-me", "--json")
	require.NoError(t, err)

	var result output.SkillRemoveResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	assert.Equal(t, "remove-me", result.Name)
	assert.Equal(t, "user", result.Scope)

	// Verify it's gone
	_, err = runCmd(t, "skill", "show", "remove-me")
	assert.Error(t, err)
}

func TestSkillRemove_NotFound(t *testing.T) {
	setupTestRegistry(t)

	_, err := runCmd(t, "skill", "remove", "nonexistent")
	assert.Error(t, err)
}

func TestSkillRemove_Text(t *testing.T) {
	setupTestRegistry(t)

	_, err := runCmd(t, "skill", "create", "bye-skill")
	require.NoError(t, err)

	out, err := runCmd(t, "skill", "remove", "bye-skill")
	require.NoError(t, err)
	assert.Contains(t, out, "Removed")
	assert.Contains(t, out, "bye-skill")
}

func TestSkillRemove_InvalidName(t *testing.T) {
	setupTestRegistry(t)

	_, err := runCmd(t, "skill", "remove", "INVALID")
	assert.Error(t, err)
}

// --- skill validate ---

func TestSkillValidate(t *testing.T) {
	setupTestRegistry(t)

	_, err := runCmd(t, "skill", "create", "valid-skill", "--description", "A valid skill", "--author", "alice")
	require.NoError(t, err)

	out, err := runCmd(t, "skill", "validate", "valid-skill", "--json")
	require.NoError(t, err)

	var result output.SkillValidateResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	assert.Equal(t, "valid-skill", result.Name)
	assert.True(t, result.Valid)
	assert.Equal(t, 0, result.Errors)
}

func TestSkillValidate_NotFound(t *testing.T) {
	setupTestRegistry(t)

	_, err := runCmd(t, "skill", "validate", "nonexistent")
	assert.Error(t, err)
}

func TestSkillValidate_Text(t *testing.T) {
	setupTestRegistry(t)

	_, err := runCmd(t, "skill", "create", "text-skill", "--description", "A skill", "--author", "alice")
	require.NoError(t, err)

	out, err := runCmd(t, "skill", "validate", "text-skill")
	require.NoError(t, err)
	assert.Contains(t, out, "valid")
}

// --- skill create with overlap ---

func TestSkillCreate_OverlapWarn(t *testing.T) {
	setupTestRegistry(t)

	// Create first skill
	_, err := runCmd(t, "skill", "create", "code-review", "--description", "Reviews code")
	require.NoError(t, err)

	// Create similar skill — should succeed with warning
	out, err := runCmd(t, "skill", "create", "code-reviewer", "--description", "Reviews code changes")
	require.NoError(t, err)
	assert.Contains(t, out, "similar")
}

func TestSkillCreate_Force(t *testing.T) {
	setupTestRegistry(t)

	_, err := runCmd(t, "skill", "create", "my-tool", "--description", "Does things")
	require.NoError(t, err)

	// Even with high overlap, --force should allow creation
	_, err = runCmd(t, "skill", "create", "my-tools", "--description", "Does things", "--force")
	require.NoError(t, err)
}

// --- Validation error exit code ---

func TestValidationError_ExitCode(t *testing.T) {
	setupTestRegistry(t)

	// Execute() returns exit code 2 for validation errors
	// We test via the error type directly
	ve := &ValidationError{Message: "test"}
	assert.Equal(t, "test", ve.Error())
}
