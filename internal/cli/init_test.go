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

func TestInit(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() { _ = os.Chdir(origDir) })

	out, err := runCmd(t, nil, "init")
	require.NoError(t, err)
	assert.Contains(t, out, "Initialized")

	// Verify directories exist
	info, err := os.Stat(filepath.Join(dir, ".skern", "skills"))
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestInit_Idempotent(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() { _ = os.Chdir(origDir) })

	// First init
	_, err := runCmd(t, nil, "init")
	require.NoError(t, err)

	// Second init should succeed with "already initialized" message
	out, err := runCmd(t, nil, "init")
	require.NoError(t, err)
	assert.Contains(t, out, "Already initialized")
}

func TestInit_JSON(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() { _ = os.Chdir(origDir) })

	out, err := runCmd(t, nil, "init", "--json")
	require.NoError(t, err)

	var result output.InitResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	assert.True(t, result.Created)
	assert.NotEmpty(t, result.Path)
}

func TestInit_JSON_AlreadyExists(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() { _ = os.Chdir(origDir) })

	// First init
	_, err := runCmd(t, nil, "init", "--json")
	require.NoError(t, err)

	// Second init
	out, err := runCmd(t, nil, "init", "--json")
	require.NoError(t, err)

	var result output.InitResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	assert.False(t, result.Created)
}
