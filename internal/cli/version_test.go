package cli

import (
	"bytes"
	"runtime/debug"
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

func TestInitVersionFromBuildInfo(t *testing.T) {
	tests := []struct {
		name            string
		initialVersion  string
		buildInfo       *debug.BuildInfo
		buildInfoOk     bool
		expectedVersion string
		expectedCommit  string
		expectedDate    string
	}{
		{
			name:            "ldflags already set",
			initialVersion:  "v1.0.0",
			buildInfo:       &debug.BuildInfo{Main: debug.Module{Version: "v2.0.0"}},
			buildInfoOk:     true,
			expectedVersion: "v1.0.0",
			expectedCommit:  "none",
			expectedDate:    "unknown",
		},
		{
			name:            "build info unavailable",
			initialVersion:  "dev",
			buildInfo:       nil,
			buildInfoOk:     false,
			expectedVersion: "dev",
			expectedCommit:  "none",
			expectedDate:    "unknown",
		},
		{
			name:           "go install with version and vcs info",
			initialVersion: "dev",
			buildInfo: &debug.BuildInfo{
				Main: debug.Module{Version: "v0.1.1"},
				Settings: []debug.BuildSetting{
					{Key: "vcs.revision", Value: "abc123def456"},
					{Key: "vcs.time", Value: "2026-01-15T10:30:00Z"},
				},
			},
			buildInfoOk:     true,
			expectedVersion: "v0.1.1",
			expectedCommit:  "abc123def456",
			expectedDate:    "2026-01-15T10:30:00Z",
		},
		{
			name:           "devel version is ignored",
			initialVersion: "dev",
			buildInfo: &debug.BuildInfo{
				Main: debug.Module{Version: "(devel)"},
				Settings: []debug.BuildSetting{
					{Key: "vcs.revision", Value: "abc123"},
				},
			},
			buildInfoOk:     true,
			expectedVersion: "dev",
			expectedCommit:  "abc123",
			expectedDate:    "unknown",
		},
		{
			name:           "version set but no vcs info",
			initialVersion: "dev",
			buildInfo: &debug.BuildInfo{
				Main: debug.Module{Version: "v0.2.0"},
			},
			buildInfoOk:     true,
			expectedVersion: "v0.2.0",
			expectedCommit:  "none",
			expectedDate:    "unknown",
		},
		{
			name:           "empty vcs values are ignored",
			initialVersion: "dev",
			buildInfo: &debug.BuildInfo{
				Main: debug.Module{Version: "v0.3.0"},
				Settings: []debug.BuildSetting{
					{Key: "vcs.revision", Value: ""},
					{Key: "vcs.time", Value: ""},
				},
			},
			buildInfoOk:     true,
			expectedVersion: "v0.3.0",
			expectedCommit:  "none",
			expectedDate:    "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore globals
			origVersion, origCommit, origDate := Version, Commit, Date
			t.Cleanup(func() {
				Version, Commit, Date = origVersion, origCommit, origDate
			})

			Version = tt.initialVersion
			Commit = "none"
			Date = "unknown"

			reader := func() (*debug.BuildInfo, bool) {
				return tt.buildInfo, tt.buildInfoOk
			}
			initVersionFromBuildInfo(reader)

			assert.Equal(t, tt.expectedVersion, Version)
			assert.Equal(t, tt.expectedCommit, Commit)
			assert.Equal(t, tt.expectedDate, Date)
		})
	}
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
