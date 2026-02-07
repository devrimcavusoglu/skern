package cli

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVersionCommand(t *testing.T) {
	Version = "0.0.1"
	Commit = "abc1234"
	Date = "2026-02-07"

	cmd := NewRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"version"})

	err := cmd.Execute()
	require.NoError(t, err)
}

func TestVersionCommand_JSON(t *testing.T) {
	Version = "0.0.1"
	Commit = "abc1234"
	Date = "2026-02-07"

	cmd := NewRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"version", "--json"})

	err := cmd.Execute()
	require.NoError(t, err)
}

func TestRootCommand_Help(t *testing.T) {
	cmd := NewRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "Agent Skills")
}

func TestRootCommand_UnknownSubcommand(t *testing.T) {
	cmd := NewRootCmd()
	cmd.SetArgs([]string{"nonexistent"})

	err := cmd.Execute()
	assert.Error(t, err)
}
