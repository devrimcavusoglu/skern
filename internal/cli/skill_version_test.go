package cli

import (
	"encoding/json"
	"testing"

	"github.com/devrimcavusoglu/skern/internal/output"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- skill version (show) ---

func TestSkillVersion_Show(t *testing.T) {
	cc := testRegistry(t)

	_, err := runCmd(t, cc, "skill", "create", "ver-skill", "--description", "A skill")
	require.NoError(t, err)

	out, err := runCmd(t, cc, "skill", "version", "ver-skill", "--json")
	require.NoError(t, err)

	var result output.SkillVersionResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	assert.Equal(t, "ver-skill", result.Name)
	assert.Equal(t, "0.1.0", result.Version)
	assert.Equal(t, "user", result.Scope)
	assert.False(t, result.Bumped)
}

func TestSkillVersion_Show_Text(t *testing.T) {
	cc := testRegistry(t)

	_, err := runCmd(t, cc, "skill", "create", "ver-text", "--description", "A skill")
	require.NoError(t, err)

	out, err := runCmd(t, cc, "skill", "version", "ver-text")
	require.NoError(t, err)
	assert.Contains(t, out, "0.1.0")
}

func TestSkillVersion_NotFound(t *testing.T) {
	cc := testRegistry(t)

	_, err := runCmd(t, cc, "skill", "version", "nonexistent")
	assert.Error(t, err)
}

// --- skill version --bump ---

func TestSkillVersion_BumpPatch(t *testing.T) {
	cc := testRegistry(t)

	_, err := runCmd(t, cc, "skill", "create", "bump-patch", "--description", "A skill")
	require.NoError(t, err)

	out, err := runCmd(t, cc, "skill", "version", "bump-patch", "--bump", "patch", "--json")
	require.NoError(t, err)

	var result output.SkillVersionResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	assert.Equal(t, "bump-patch", result.Name)
	assert.Equal(t, "0.1.1", result.Version)
	assert.Equal(t, "0.1.0", result.PreviousVersion)
	assert.True(t, result.Bumped)

	// Verify the change persisted
	showOut, err := runCmd(t, cc, "skill", "show", "bump-patch", "--json")
	require.NoError(t, err)

	var showResult output.SkillResult
	require.NoError(t, json.Unmarshal([]byte(showOut), &showResult))
	assert.Equal(t, "0.1.1", showResult.Version)
}

func TestSkillVersion_BumpMinor(t *testing.T) {
	cc := testRegistry(t)

	_, err := runCmd(t, cc, "skill", "create", "bump-minor", "--description", "A skill")
	require.NoError(t, err)

	out, err := runCmd(t, cc, "skill", "version", "bump-minor", "--bump", "minor", "--json")
	require.NoError(t, err)

	var result output.SkillVersionResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	assert.Equal(t, "0.2.0", result.Version)
	assert.Equal(t, "0.1.0", result.PreviousVersion)
	assert.True(t, result.Bumped)
}

func TestSkillVersion_BumpMajor(t *testing.T) {
	cc := testRegistry(t)

	_, err := runCmd(t, cc, "skill", "create", "bump-major", "--description", "A skill")
	require.NoError(t, err)

	out, err := runCmd(t, cc, "skill", "version", "bump-major", "--bump", "major", "--json")
	require.NoError(t, err)

	var result output.SkillVersionResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	assert.Equal(t, "1.0.0", result.Version)
	assert.Equal(t, "0.1.0", result.PreviousVersion)
	assert.True(t, result.Bumped)
}

func TestSkillVersion_BumpInvalidLevel(t *testing.T) {
	cc := testRegistry(t)

	_, err := runCmd(t, cc, "skill", "create", "bump-bad", "--description", "A skill")
	require.NoError(t, err)

	_, err = runCmd(t, cc, "skill", "version", "bump-bad", "--bump", "invalid")
	assert.Error(t, err)
}

func TestSkillVersion_BumpText(t *testing.T) {
	cc := testRegistry(t)

	_, err := runCmd(t, cc, "skill", "create", "bump-text", "--description", "A skill")
	require.NoError(t, err)

	out, err := runCmd(t, cc, "skill", "version", "bump-text", "--bump", "patch")
	require.NoError(t, err)
	assert.Contains(t, out, "Bumped")
	assert.Contains(t, out, "0.1.0")
	assert.Contains(t, out, "0.1.1")
}

func TestSkillVersion_MultipleBumps(t *testing.T) {
	cc := testRegistry(t)

	_, err := runCmd(t, cc, "skill", "create", "multi-bump", "--description", "A skill")
	require.NoError(t, err)

	// Bump patch twice
	_, err = runCmd(t, cc, "skill", "version", "multi-bump", "--bump", "patch")
	require.NoError(t, err)
	_, err = runCmd(t, cc, "skill", "version", "multi-bump", "--bump", "patch")
	require.NoError(t, err)

	out, err := runCmd(t, cc, "skill", "version", "multi-bump", "--json")
	require.NoError(t, err)

	var result output.SkillVersionResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	assert.Equal(t, "0.1.2", result.Version)
}

func TestSkillVersion_Scoped(t *testing.T) {
	cc := testRegistry(t)

	_, err := runCmd(t, cc, "skill", "create", "scoped-ver", "--scope", "project", "--description", "A skill")
	require.NoError(t, err)

	out, err := runCmd(t, cc, "skill", "version", "scoped-ver", "--scope", "project", "--json")
	require.NoError(t, err)

	var result output.SkillVersionResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	assert.Equal(t, "project", result.Scope)
	assert.Equal(t, "0.1.0", result.Version)
}

// --- skill create --version ---

func TestSkillCreate_WithVersion(t *testing.T) {
	cc := testRegistry(t)

	_, err := runCmd(t, cc, "skill", "create", "versioned-skill",
		"--description", "A skill", "--version", "0.2.0")
	require.NoError(t, err)

	out, err := runCmd(t, cc, "skill", "show", "versioned-skill", "--json")
	require.NoError(t, err)

	var result output.SkillResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	assert.Equal(t, "0.2.0", result.Version)
}

func TestSkillCreate_WithVersion_JSON(t *testing.T) {
	cc := testRegistry(t)

	out, err := runCmd(t, cc, "skill", "create", "ver-json-skill",
		"--description", "A skill", "--version", "1.0.0", "--json")
	require.NoError(t, err)

	var createResult output.SkillCreateResult
	require.NoError(t, json.Unmarshal([]byte(out), &createResult))
	assert.Equal(t, "ver-json-skill", createResult.Name)

	// Verify version was set
	showOut, err := runCmd(t, cc, "skill", "show", "ver-json-skill", "--json")
	require.NoError(t, err)

	var showResult output.SkillResult
	require.NoError(t, json.Unmarshal([]byte(showOut), &showResult))
	assert.Equal(t, "1.0.0", showResult.Version)
}

func TestSkillCreate_WithVersion_Invalid(t *testing.T) {
	cc := testRegistry(t)

	_, err := runCmd(t, cc, "skill", "create", "bad-ver-skill",
		"--description", "A skill", "--version", "bad")
	assert.Error(t, err)
}

func TestSkillCreate_DefaultVersion(t *testing.T) {
	cc := testRegistry(t)

	_, err := runCmd(t, cc, "skill", "create", "default-ver",
		"--description", "A skill")
	require.NoError(t, err)

	out, err := runCmd(t, cc, "skill", "show", "default-ver", "--json")
	require.NoError(t, err)

	var result output.SkillResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	assert.Equal(t, "0.1.0", result.Version)
}
