package skill

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseManifest_AllFields(t *testing.T) {
	content := `---
name: my-skill
description: |
  A test skill for unit testing.
allowed-tools:
  - tool-a
  - tool-b
metadata:
  author:
    name: alice
    type: human
  version: "0.1.0"
  modified-by:
    - name: codex-cli
      type: agent
      platform: codex-cli
      date: "2025-07-15T10:30:00Z"
---

## Instructions

Do the thing.
`

	path := filepath.Join(t.TempDir(), "SKILL.md")
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))

	s, err := ParseManifest(path)
	require.NoError(t, err)

	assert.Equal(t, "my-skill", s.Name)
	assert.Contains(t, s.Description, "A test skill")
	assert.Equal(t, []string{"tool-a", "tool-b"}, s.AllowedTools)
	assert.Equal(t, "alice", s.Metadata.Author.Name)
	assert.Equal(t, "human", s.Metadata.Author.Type)
	assert.Equal(t, "0.1.0", s.Metadata.Version)
	assert.Len(t, s.Metadata.ModifiedBy, 1)
	assert.Equal(t, "codex-cli", s.Metadata.ModifiedBy[0].Name)
	assert.Contains(t, s.Body, "## Instructions")
	assert.Contains(t, s.Body, "Do the thing.")
}

func TestParseManifest_Minimal(t *testing.T) {
	content := `---
name: minimal
description: A minimal skill.
metadata:
  author:
    name: bob
    type: human
  version: "0.1.0"
---

## Instructions

TODO
`

	path := filepath.Join(t.TempDir(), "SKILL.md")
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))

	s, err := ParseManifest(path)
	require.NoError(t, err)

	assert.Equal(t, "minimal", s.Name)
	assert.Equal(t, "A minimal skill.", s.Description)
	assert.Nil(t, s.AllowedTools)
	assert.Equal(t, "bob", s.Metadata.Author.Name)
}

func TestParseManifest_MissingDelimiters(t *testing.T) {
	content := `name: bad
description: no delimiters
`

	path := filepath.Join(t.TempDir(), "SKILL.md")
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))

	_, err := ParseManifest(path)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "delimiter")
}

func TestParseManifest_MissingClosingDelimiter(t *testing.T) {
	content := `---
name: bad
description: missing closing
`

	path := filepath.Join(t.TempDir(), "SKILL.md")
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))

	_, err := ParseManifest(path)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "closing")
}

func TestParseManifest_MalformedYAML(t *testing.T) {
	content := `---
name: [invalid
  bad: yaml: ::
---

Body
`

	path := filepath.Join(t.TempDir(), "SKILL.md")
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))

	_, err := ParseManifest(path)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "YAML")
}

func TestParseManifest_EmptyFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "SKILL.md")
	require.NoError(t, os.WriteFile(path, []byte(""), 0o644))

	_, err := ParseManifest(path)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty")
}

func TestParseManifest_FileNotFound(t *testing.T) {
	_, err := ParseManifest("/nonexistent/SKILL.md")
	assert.Error(t, err)
}

func TestWriteManifest(t *testing.T) {
	s := &Skill{
		Name:        "test-skill",
		Description: "A test skill.\n",
		Metadata: Metadata{
			Author:  Author{Name: "alice", Type: "human"},
			Version: "0.1.0",
		},
		Body: "## Instructions\n\nDo the thing.\n",
	}

	path := filepath.Join(t.TempDir(), "SKILL.md")
	require.NoError(t, WriteManifest(s, path))

	data, err := os.ReadFile(path)
	require.NoError(t, err)

	content := string(data)
	assert.Contains(t, content, "---\n")
	assert.Contains(t, content, "name: test-skill")
	assert.Contains(t, content, "## Instructions")
}

func TestManifest_Roundtrip(t *testing.T) {
	original := &Skill{
		Name:         "roundtrip-skill",
		Description:  "Round-trip test.\n",
		AllowedTools: []string{"tool-x"},
		Metadata: Metadata{
			Author:  Author{Name: "charlie", Type: "agent", Platform: "claude-code"},
			Version: "1.0.0",
			ModifiedBy: []ModifiedByEntry{
				{Name: "dave", Type: "human", Date: "2025-01-01T00:00:00Z"},
			},
		},
		Body: "## Instructions\n\nStep 1.\n",
	}

	path := filepath.Join(t.TempDir(), "SKILL.md")
	require.NoError(t, WriteManifest(original, path))

	parsed, err := ParseManifest(path)
	require.NoError(t, err)

	assert.Equal(t, original.Name, parsed.Name)
	assert.Equal(t, original.Description, parsed.Description)
	assert.Equal(t, original.AllowedTools, parsed.AllowedTools)
	assert.Equal(t, original.Metadata.Author, parsed.Metadata.Author)
	assert.Equal(t, original.Metadata.Version, parsed.Metadata.Version)
	assert.Equal(t, original.Metadata.ModifiedBy, parsed.Metadata.ModifiedBy)
	assert.Equal(t, original.Body, parsed.Body)
}

func TestManifest_Roundtrip_Tags(t *testing.T) {
	original := &Skill{
		Name:        "tagged-skill",
		Description: "A skill with tags.\n",
		Tags:        []string{"code-review", "testing"},
		Metadata: Metadata{
			Author:  Author{Name: "alice", Type: "human"},
			Version: "0.1.0",
		},
		Body: "## Instructions\n\nReview code.\n",
	}

	path := filepath.Join(t.TempDir(), "SKILL.md")
	require.NoError(t, WriteManifest(original, path))

	parsed, err := ParseManifest(path)
	require.NoError(t, err)

	assert.Equal(t, original.Tags, parsed.Tags)
}

func TestParseManifest_WithTags(t *testing.T) {
	content := `---
name: my-skill
description: A tagged skill.
tags:
  - devops
  - ci-cd
metadata:
  author:
    name: bob
    type: human
  version: "0.1.0"
---

## Instructions

Deploy things.
`

	path := filepath.Join(t.TempDir(), "SKILL.md")
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))

	s, err := ParseManifest(path)
	require.NoError(t, err)

	assert.Equal(t, []string{"devops", "ci-cd"}, s.Tags)
}

func TestParseManifest_NoTags(t *testing.T) {
	content := `---
name: my-skill
description: No tags.
metadata:
  author:
    name: bob
    type: human
  version: "0.1.0"
---

Body.
`

	path := filepath.Join(t.TempDir(), "SKILL.md")
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))

	s, err := ParseManifest(path)
	require.NoError(t, err)

	assert.Nil(t, s.Tags)
}
