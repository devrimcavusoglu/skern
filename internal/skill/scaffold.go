package skill

// DefaultBody returns the default body content for a new skill.
// The template follows the writing-skills guidelines: Overview, When to Use,
// Core Pattern, Quick Reference, and Common Mistakes sections.
func DefaultBody() string {
	return `## Overview

<!-- 1-2 sentences: the core principle or technique this skill provides. -->

TODO: Describe the core principle of this skill.

## When to Use

<!-- Triggering conditions — symptoms and use cases that signal this skill is needed. -->

- TODO: Add triggering conditions

## Core Pattern

<!-- The main technique or pattern. Use before/after examples for techniques. -->

TODO: Describe the core pattern or technique.

## Quick Reference

<!-- Scannable summary — a table or bullet list for fast lookup. -->

- TODO: Add quick reference items

## Common Mistakes

<!-- Frequent errors and their fixes. -->

- TODO: Add common mistakes and how to avoid them
`
}

// DefaultDescription returns the default placeholder description for a new skill.
// Follows the writing-skills guideline: start with "Use when..." to describe
// triggering conditions, not a workflow summary.
func DefaultDescription() string {
	return "Use when TODO: describe the triggering conditions for this skill.\n"
}

// NewSkill creates a new Skill with sensible defaults.
func NewSkill(name, description, authorName, authorType, authorPlatform string) *Skill {
	return NewSkillWithBody(name, description, authorName, authorType, authorPlatform, "")
}

// NewSkillWithBody creates a new Skill with a custom body. If body is empty, DefaultBody() is used.
func NewSkillWithBody(name, description, authorName, authorType, authorPlatform, body string) *Skill {
	if description == "" {
		description = DefaultDescription()
	}

	if body == "" {
		body = DefaultBody()
	}

	author := Author{
		Name:     authorName,
		Type:     authorType,
		Platform: authorPlatform,
	}

	return &Skill{
		Name:        name,
		Description: description,
		Metadata: Metadata{
			Author:  author,
			Version: "0.0.1",
		},
		Body: body,
	}
}
