package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/devrimcavusoglu/skern/internal/cli/instructions"
	"github.com/devrimcavusoglu/skern/internal/output"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// withTempCwd switches the test process to a fresh temp dir and restores
// the original on cleanup. Init writes relative paths so tests must run
// from a clean cwd.
func withTempCwd(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	orig, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() { _ = os.Chdir(orig) })
	return dir
}

func TestInit(t *testing.T) {
	dir := withTempCwd(t)

	out, err := runCmd(t, nil, "init")
	require.NoError(t, err)
	assert.Contains(t, out, "Initialized")

	info, err := os.Stat(filepath.Join(dir, ".skern", "skills"))
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestInit_Idempotent(t *testing.T) {
	withTempCwd(t)

	_, err := runCmd(t, nil, "init")
	require.NoError(t, err)

	out, err := runCmd(t, nil, "init")
	require.NoError(t, err)
	assert.Contains(t, out, "Already initialized")
}

func TestInit_JSON(t *testing.T) {
	withTempCwd(t)

	out, err := runCmd(t, nil, "init", "--json")
	require.NoError(t, err)

	var result output.InitResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	assert.True(t, result.Created)
	assert.NotEmpty(t, result.Path)
	assert.Nil(t, result.Instructions, "Instructions should be omitted when not opted in")
}

func TestInit_JSON_AlreadyExists(t *testing.T) {
	withTempCwd(t)

	_, err := runCmd(t, nil, "init", "--json")
	require.NoError(t, err)

	out, err := runCmd(t, nil, "init", "--json")
	require.NoError(t, err)

	var result output.InitResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	assert.False(t, result.Created)
}

// --- Instruction-snippet flag flows ---

func TestInit_Instructions_NoTargetsFound(t *testing.T) {
	withTempCwd(t)

	out, err := runCmd(t, nil, "init", "--instructions", "--json")
	require.NoError(t, err)

	var result output.InitResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	require.NotNil(t, result.Instructions)
	assert.False(t, result.Instructions.ToolForming)
	assert.Empty(t, result.Instructions.Targets)
	assert.Empty(t, result.Instructions.Writes)
	assert.False(t, result.Instructions.Printed)
}

func TestInit_Instructions_DiscoversAndWrites(t *testing.T) {
	dir := withTempCwd(t)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte("# Project\n"), 0o644))

	out, err := runCmd(t, nil, "init", "--instructions", "--json")
	require.NoError(t, err)

	var result output.InitResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	require.NotNil(t, result.Instructions)
	require.Len(t, result.Instructions.Writes, 1)
	assert.Equal(t, filepath.Join("AGENTS.md"), result.Instructions.Writes[0].Path)
	assert.Equal(t, "appended", result.Instructions.Writes[0].Action)

	body, err := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
	require.NoError(t, err)
	bodyStr := string(body)
	assert.Contains(t, bodyStr, "# Project")
	assert.Contains(t, bodyStr, instructions.StartMarker)
	assert.Contains(t, bodyStr, "Skern (skill management)")
	assert.NotContains(t, bodyStr, "Tool-forming loop", "tool-forming should be off by default")
}

func TestInit_Instructions_WithToolFormingLoop(t *testing.T) {
	dir := withTempCwd(t)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte(""), 0o644))

	_, err := runCmd(t, nil, "init", "--instructions", "--tool-forming-loop", "--json")
	require.NoError(t, err)

	body, err := os.ReadFile(filepath.Join(dir, "CLAUDE.md"))
	require.NoError(t, err)
	assert.Contains(t, string(body), "Tool-forming loop")
	assert.Contains(t, string(body), "skern skill search")
}

func TestInit_Instructions_TargetCreatesNewFile(t *testing.T) {
	dir := withTempCwd(t)
	target := filepath.Join("subdir", "MY_AGENT.md")

	_, err := runCmd(t, nil, "init", "--target", target, "--json")
	require.NoError(t, err)

	body, err := os.ReadFile(filepath.Join(dir, target))
	require.NoError(t, err)
	assert.Contains(t, string(body), instructions.StartMarker)
	assert.Contains(t, string(body), instructions.EndMarker)
}

func TestInit_Instructions_TargetSkipsAutoDiscovery(t *testing.T) {
	dir := withTempCwd(t)
	// Create both a candidate file and an explicit target. Only the target
	// should be touched.
	require.NoError(t, os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte("orig"), 0o644))
	target := "EXPLICIT.md"
	require.NoError(t, os.WriteFile(filepath.Join(dir, target), []byte("explicit-orig"), 0o644))

	out, err := runCmd(t, nil, "init", "--target", target, "--json")
	require.NoError(t, err)

	var result output.InitResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	require.NotNil(t, result.Instructions)
	require.Len(t, result.Instructions.Writes, 1)
	assert.Equal(t, target, result.Instructions.Writes[0].Path)

	agents, err := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
	require.NoError(t, err)
	assert.Equal(t, "orig", string(agents), "auto-discovery candidate must not be touched when --target is set")
}

func TestInit_Instructions_PrintToStdout(t *testing.T) {
	dir := withTempCwd(t)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte("untouched"), 0o644))

	out, err := runCmd(t, nil, "init", "--print-instructions", "--tool-forming-loop")
	require.NoError(t, err)

	assert.Contains(t, out, instructions.StartMarker)
	assert.Contains(t, out, "Tool-forming loop")

	// Discovered file must NOT be modified in print mode.
	body, err := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
	require.NoError(t, err)
	assert.Equal(t, "untouched", string(body))
}

func TestInit_Instructions_IdempotentReRun(t *testing.T) {
	dir := withTempCwd(t)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte("# proj\n"), 0o644))

	_, err := runCmd(t, nil, "init", "--instructions", "--json")
	require.NoError(t, err)

	first, err := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
	require.NoError(t, err)

	out, err := runCmd(t, nil, "init", "--instructions", "--json")
	require.NoError(t, err)

	second, err := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
	require.NoError(t, err)
	assert.Equal(t, string(first), string(second), "second run must be a no-op when content unchanged")

	var result output.InitResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	require.NotNil(t, result.Instructions)
	require.Len(t, result.Instructions.Writes, 1)
	assert.Equal(t, "unchanged", result.Instructions.Writes[0].Action)
}

func TestInit_Instructions_UpdatesBlockOnToggleChange(t *testing.T) {
	dir := withTempCwd(t)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte(""), 0o644))

	// Round 1: opt out of tool-forming.
	_, err := runCmd(t, nil, "init", "--instructions", "--json")
	require.NoError(t, err)

	round1, err := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
	require.NoError(t, err)
	assert.NotContains(t, string(round1), "Tool-forming loop")

	// Round 2: opt in. The block should be replaced in place, not appended.
	_, err = runCmd(t, nil, "init", "--instructions", "--tool-forming-loop", "--json")
	require.NoError(t, err)

	round2, err := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
	require.NoError(t, err)
	round2Str := string(round2)
	assert.Contains(t, round2Str, "Tool-forming loop")
	assert.Equal(t, 1, strings.Count(round2Str, instructions.StartMarker),
		"block must be replaced, not duplicated, on re-run")
}

func TestInit_Instructions_NoPromptInJSONMode(t *testing.T) {
	withTempCwd(t)
	// Run --json with no instruction flags. Should not prompt (no stdin
	// input given, and with --json the writer never reaches the prompt
	// branch). Result.Instructions must be nil.
	out, err := runCmd(t, nil, "init", "--json")
	require.NoError(t, err)

	var result output.InitResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	assert.Nil(t, result.Instructions)
}
