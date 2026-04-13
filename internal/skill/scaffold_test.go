package skill

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSkill_NameOnly(t *testing.T) {
	s := NewSkill("my-skill", "", "", "", "")

	assert.Equal(t, "my-skill", s.Name)
	assert.Contains(t, s.Description, "Use when")
	assert.Contains(t, s.Description, "TODO")
	assert.Equal(t, "0.0.1", s.Metadata.Version)
	assert.Contains(t, s.Body, "## Overview")
	assert.Contains(t, s.Body, "## When to Use")
	assert.Contains(t, s.Body, "## Core Pattern")
	assert.Contains(t, s.Body, "## Quick Reference")
	assert.Contains(t, s.Body, "## Common Mistakes")
}

func TestNewSkill_WithDescription(t *testing.T) {
	s := NewSkill("my-skill", "A custom description.", "", "", "")

	assert.Equal(t, "A custom description.", s.Description)
}

func TestNewSkill_WithHumanAuthor(t *testing.T) {
	s := NewSkill("my-skill", "Desc.", "alice", "human", "")

	assert.Equal(t, "alice", s.Metadata.Author.Name)
	assert.Equal(t, "human", s.Metadata.Author.Type)
	assert.Empty(t, s.Metadata.Author.Platform)
}

func TestNewSkill_WithAgentAuthor(t *testing.T) {
	s := NewSkill("my-skill", "Desc.", "claude", "agent", "claude-code")

	assert.Equal(t, "claude", s.Metadata.Author.Name)
	assert.Equal(t, "agent", s.Metadata.Author.Type)
	assert.Equal(t, "claude-code", s.Metadata.Author.Platform)
}

func TestDefaultBody(t *testing.T) {
	body := DefaultBody()
	assert.Contains(t, body, "## Overview")
	assert.Contains(t, body, "## When to Use")
	assert.Contains(t, body, "## Core Pattern")
	assert.Contains(t, body, "## Quick Reference")
	assert.Contains(t, body, "## Common Mistakes")
	assert.Contains(t, body, "TODO")
}

func TestDefaultDescription(t *testing.T) {
	desc := DefaultDescription()
	assert.True(t, len(desc) > 0)
	assert.Contains(t, desc, "Use when")
	assert.Contains(t, desc, "TODO")
}
