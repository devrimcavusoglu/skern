package skill

import (
	"os"
	"path/filepath"
	"strings"
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

// The template from issue #100: unmodeled top-level keys (compatibility,
// handoffs), an unmodeled metadata.* key (phases), and metadata.tags (distinct
// from top-level tags) must all survive parse → write → parse.
const extraKeysManifest = `---
name: demo.plan
description: top level tags round trip check
tags: [workflow, planning]
compatibility: "requires a project checkout"
metadata:
  version: "2.0.0"
  phases: [plan]
  tags: [meta-only]
handoffs:
  - label: Review
    agent: demo.review
---

Body.
`

func TestParseManifest_ExtraKeysCaptured(t *testing.T) {
	s, err := ParseManifestFromBytes([]byte(extraKeysManifest))
	require.NoError(t, err)

	// Modeled keys still land in their fields, never in Extra.
	assert.Equal(t, "demo.plan", s.Name)
	assert.Equal(t, []string{"workflow", "planning"}, s.Tags)
	assert.Equal(t, "2.0.0", s.Metadata.Version)
	assert.NotContains(t, s.Extra, "name")
	assert.NotContains(t, s.Extra, "tags")
	assert.NotContains(t, s.Extra, "metadata")
	assert.NotContains(t, s.Metadata.Extra, "version")

	assert.Equal(t, "requires a project checkout", s.Extra["compatibility"])
	handoffs, ok := s.Extra["handoffs"].([]any)
	require.True(t, ok, "handoffs should decode as a sequence, got %T", s.Extra["handoffs"])
	require.Len(t, handoffs, 1)
	first, ok := handoffs[0].(map[string]any)
	require.True(t, ok, "handoff entry should decode as a string-keyed map, got %T", handoffs[0])
	assert.Equal(t, "Review", first["label"])
	assert.Equal(t, "demo.review", first["agent"])

	assert.Equal(t, []any{"plan"}, s.Metadata.Extra["phases"])
	assert.Equal(t, []any{"meta-only"}, s.Metadata.Extra["tags"])
}

func TestParseManifest_NoExtraKeysLeavesMapsNil(t *testing.T) {
	content := `---
name: plain
description: No unmodeled keys here.
metadata:
  author:
    name: bob
    type: human
  version: "0.1.0"
---

Body.
`
	s, err := ParseManifestFromBytes([]byte(content))
	require.NoError(t, err)
	assert.Nil(t, s.Extra)
	assert.Nil(t, s.Metadata.Extra)
}

func TestManifest_Roundtrip_ExtraKeys(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.md")
	require.NoError(t, os.WriteFile(src, []byte(extraKeysManifest), 0o644))

	s, err := ParseManifest(src)
	require.NoError(t, err)

	// Mutate a modeled field the way skill edit / version would, then write.
	s.Metadata.Version = "2.1.0"
	dst := filepath.Join(dir, "SKILL.md")
	require.NoError(t, WriteManifest(s, dst))

	written, err := os.ReadFile(dst)
	require.NoError(t, err)
	text := string(written)
	assert.Contains(t, text, "compatibility: requires a project checkout")
	assert.Contains(t, text, "handoffs:")
	assert.Contains(t, text, "label: Review")
	assert.Contains(t, text, "agent: demo.review")
	assert.Contains(t, text, "phases:")
	assert.Contains(t, text, "version: 2.1.0")

	// Modeled keys come first, extras after — and nothing is duplicated.
	assert.Less(t, strings.Index(text, "name: demo.plan"), strings.Index(text, "compatibility:"))
	assert.Equal(t, 1, strings.Count(text, "\ntags:"), "top-level tags must be written exactly once")

	// A second parse sees the same extras and the updated modeled field.
	again, err := ParseManifest(dst)
	require.NoError(t, err)
	assert.Equal(t, "2.1.0", again.Metadata.Version)
	assert.Equal(t, s.Extra, again.Extra)
	assert.Equal(t, s.Metadata.Extra, again.Metadata.Extra)
	assert.Equal(t, []string{"workflow", "planning"}, again.Tags)
	assert.Equal(t, []any{"meta-only"}, again.Metadata.Extra["tags"])
}

func TestWriteManifest_ExtraKeyCollisionIsError(t *testing.T) {
	s := &Skill{
		Name:        "clash",
		Description: "Extra shadows a modeled key.",
		Extra:       map[string]any{"description": "shadow", "zzz": 1},
		Metadata: Metadata{
			Version: "0.1.0",
			Extra:   map[string]any{"version": "9.9.9"},
		},
		Body: "Body.",
	}
	path := filepath.Join(t.TempDir(), "SKILL.md")

	// Must return an error, not panic (yaml.v3 panics on inline-key collisions).
	require.NotPanics(t, func() {
		err := WriteManifest(s, path)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "description")
		assert.Contains(t, err.Error(), "collide")
	})
	_, statErr := os.Stat(path)
	assert.True(t, os.IsNotExist(statErr), "nothing should be written on collision")

	// Fix the top-level clash; the metadata clash is reported next.
	delete(s.Extra, "description")
	err := WriteManifest(s, path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "metadata")
	assert.Contains(t, err.Error(), "version")

	delete(s.Metadata.Extra, "version")
	require.NoError(t, WriteManifest(s, path))
}

func TestRenderExtraValue(t *testing.T) {
	assert.Equal(t, "", RenderExtraValue(nil))
	assert.Equal(t, "plain", RenderExtraValue("plain"))
	assert.Equal(t, "42", RenderExtraValue(42))
	assert.Equal(t, "[plan, review]", RenderExtraValue([]any{"plan", "review"}))
	assert.Equal(t, "{agent: demo.review, label: Review}",
		RenderExtraValue(map[string]any{"label": "Review", "agent": "demo.review"}))
	// Nested collections stay on one line, and int vs int64 render the same.
	assert.Equal(t, RenderExtraValue(map[string]any{"n": 1}), RenderExtraValue(map[string]any{"n": int64(1)}))
	assert.Equal(t, "[{a: 1}]", RenderExtraValue([]any{map[string]any{"a": 1}}))
}
