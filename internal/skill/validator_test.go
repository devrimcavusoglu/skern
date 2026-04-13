package skill

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidate_ValidSkill(t *testing.T) {
	s := &Skill{
		Name:        "my-skill",
		Description: "Use when you need to complete tasks with step-by-step guidance",
		Body: `## Overview

This skill provides step-by-step guidance for completing tasks.

## When to Use

- When you need structured task completion

## Core Pattern

- First, analyze the input carefully and validate the data
- Then, process the data accordingly and format the results
- Finally, return the formatted output to the user

## Quick Reference

- Analyze → Process → Format → Return

## Common Mistakes

- Skipping validation of input data
`,
		Metadata: Metadata{
			Author:  Author{Name: "alice", Type: "human"},
			Version: "1.0.0",
		},
	}

	issues := Validate(s)
	assert.Empty(t, issues)
}

func TestValidate_EmptyName(t *testing.T) {
	s := &Skill{
		Name:        "",
		Description: "A description",
		Body:        "Some body",
		Metadata:    Metadata{Author: Author{Name: "alice", Type: "human"}},
	}

	issues := Validate(s)
	require.NotEmpty(t, issues)
	assert.Equal(t, "name", issues[0].Field)
	assert.Equal(t, SeverityError, issues[0].Severity)
}

func TestValidate_InvalidName(t *testing.T) {
	s := &Skill{
		Name:        "INVALID_NAME",
		Description: "A description",
		Body:        "Some body",
		Metadata:    Metadata{Author: Author{Name: "alice", Type: "human"}},
	}

	issues := Validate(s)
	hasNameError := false
	for _, i := range issues {
		if i.Field == "name" && i.Severity == SeverityError {
			hasNameError = true
		}
	}
	assert.True(t, hasNameError)
}

func TestValidate_EmptyDescription(t *testing.T) {
	s := &Skill{
		Name:     "my-skill",
		Body:     "Some body",
		Metadata: Metadata{Author: Author{Name: "alice", Type: "human"}},
	}

	issues := Validate(s)
	hasDescError := false
	for _, i := range issues {
		if i.Field == "description" && i.Severity == SeverityError {
			hasDescError = true
		}
	}
	assert.True(t, hasDescError)
}

func TestValidate_DescriptionTooLong(t *testing.T) {
	s := &Skill{
		Name:        "my-skill",
		Description: strings.Repeat("a", 1025),
		Body:        "Some body",
		Metadata:    Metadata{Author: Author{Name: "alice", Type: "human"}},
	}

	issues := Validate(s)
	hasDescError := false
	for _, i := range issues {
		if i.Field == "description" && i.Severity == SeverityError {
			hasDescError = true
		}
	}
	assert.True(t, hasDescError)
}

func TestValidate_Description1024OK(t *testing.T) {
	s := &Skill{
		Name:        "my-skill",
		Description: strings.Repeat("a", 1024),
		Body:        "Some body",
		Metadata:    Metadata{Author: Author{Name: "alice", Type: "human"}},
	}

	issues := Validate(s)
	for _, i := range issues {
		if i.Field == "description" && i.Severity == SeverityError {
			t.Errorf("unexpected description error: %s", i.Message)
		}
	}
}

func TestValidate_EmptyBody(t *testing.T) {
	s := &Skill{
		Name:        "my-skill",
		Description: "A description",
		Metadata:    Metadata{Author: Author{Name: "alice", Type: "human"}},
	}

	issues := Validate(s)
	hasBodyError := false
	for _, i := range issues {
		if i.Field == "body" && i.Severity == SeverityError {
			hasBodyError = true
		}
	}
	assert.True(t, hasBodyError)
}

func TestValidate_EmptyAllowedTool(t *testing.T) {
	s := &Skill{
		Name:         "my-skill",
		Description:  "A description",
		Body:         "Some body",
		AllowedTools: []string{"valid-tool", ""},
		Metadata:     Metadata{Author: Author{Name: "alice", Type: "human"}},
	}

	issues := Validate(s)
	hasToolError := false
	for _, i := range issues {
		if i.Field == "allowed-tools" && i.Severity == SeverityError {
			hasToolError = true
		}
	}
	assert.True(t, hasToolError)
}

func TestValidate_MissingAuthorName(t *testing.T) {
	s := &Skill{
		Name:        "my-skill",
		Description: "A description",
		Body:        "Some body",
		Metadata:    Metadata{Author: Author{Type: "human"}},
	}

	issues := Validate(s)
	hasAuthorWarn := false
	for _, i := range issues {
		if i.Field == "metadata.author.name" && i.Severity == SeverityWarning {
			hasAuthorWarn = true
		}
	}
	assert.True(t, hasAuthorWarn)
}

func TestValidate_InvalidAuthorType(t *testing.T) {
	s := &Skill{
		Name:        "my-skill",
		Description: "A description",
		Body:        "Some body",
		Metadata:    Metadata{Author: Author{Name: "alice", Type: "bot"}},
	}

	issues := Validate(s)
	hasTypeError := false
	for _, i := range issues {
		if i.Field == "metadata.author.type" && i.Severity == SeverityError {
			hasTypeError = true
		}
	}
	assert.True(t, hasTypeError)
}

func TestValidate_AgentWithoutPlatform(t *testing.T) {
	s := &Skill{
		Name:        "my-skill",
		Description: "A description",
		Body:        "Some body",
		Metadata:    Metadata{Author: Author{Name: "claude", Type: "agent"}},
	}

	issues := Validate(s)
	hasPlatformWarn := false
	for _, i := range issues {
		if i.Field == "metadata.author.platform" && i.Severity == SeverityWarning {
			hasPlatformWarn = true
		}
	}
	assert.True(t, hasPlatformWarn)
}

func TestValidate_VersionFormat(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		wantWarn bool
	}{
		{"valid 0.1.0", "0.1.0", false},
		{"valid 1.0.0", "1.0.0", false},
		{"valid 12.34.56", "12.34.56", false},
		{"single number", "1", true},
		{"two parts", "1.2", true},
		{"non-numeric parts", "a.b.c", true},
		{"mixed non-numeric", "1.2.three", true},
		{"v prefix", "v1.0.0", true},
		{"leading zeros", "01.02.03", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Skill{
				Name:        "my-skill",
				Description: "A description",
				Body:        "Some body",
				Metadata:    Metadata{Author: Author{Name: "alice", Type: "human"}, Version: tt.version},
			}

			issues := Validate(s)
			hasVersionWarn := false
			for _, i := range issues {
				if i.Field == "metadata.version" && i.Severity == SeverityWarning {
					hasVersionWarn = true
				}
			}
			assert.Equal(t, tt.wantWarn, hasVersionWarn, "version %q", tt.version)
		})
	}
}

func TestHasErrors(t *testing.T) {
	tests := []struct {
		name   string
		issues []ValidationIssue
		want   bool
	}{
		{"no issues", nil, false},
		{"warnings only", []ValidationIssue{{Severity: SeverityWarning}}, false},
		{"hints only", []ValidationIssue{{Severity: SeverityHint}}, false},
		{"has error", []ValidationIssue{{Severity: SeverityError}}, true},
		{"mixed", []ValidationIssue{{Severity: SeverityWarning}, {Severity: SeverityError}}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, HasErrors(tt.issues))
		})
	}
}

func TestValidationIssue_String(t *testing.T) {
	issue := ValidationIssue{
		Field:    "name",
		Severity: SeverityError,
		Message:  "name is invalid",
	}
	assert.Equal(t, "[error] name: name is invalid", issue.String())
}

// Lint style tests

func TestLintStyle_BodyTooShort(t *testing.T) {
	s := &Skill{
		Name:        "my-skill",
		Description: "Use when you need a valid skill description",
		Body:        "## Instructions\n\nDo something useful.",
	}

	issues := lintStyle(s)
	hasBodyShortHint := false
	for _, i := range issues {
		if i.Field == "body" && i.Severity == SeverityHint && strings.Contains(i.Message, "words") {
			hasBodyShortHint = true
		}
	}
	assert.True(t, hasBodyShortHint)
}

func TestLintStyle_DescriptionTooVague(t *testing.T) {
	s := &Skill{
		Name:        "my-skill",
		Description: "Deploy",
		Body:        "## Instructions\n\nThis skill provides step-by-step guidance for completing tasks:\n\n- First, analyze the input carefully and validate the data\n- Then, process the data accordingly and format the results\n- Finally, return the formatted output to the user",
	}

	issues := lintStyle(s)
	hasDescVague := false
	for _, i := range issues {
		if i.Field == "description" && i.Severity == SeverityHint && strings.Contains(i.Message, "word") {
			hasDescVague = true
		}
	}
	assert.True(t, hasDescVague)
}

func TestLintStyle_BodyLacksSteps(t *testing.T) {
	s := &Skill{
		Name:        "my-skill",
		Description: "Use when you need to do something specific",
		Body:        strings.Repeat("word ", 25),
	}

	issues := lintStyle(s)
	hasStepHint := false
	for _, i := range issues {
		if i.Field == "body" && i.Severity == SeverityHint && strings.Contains(i.Message, "step-by-step") {
			hasStepHint = true
		}
	}
	assert.True(t, hasStepHint)
}

func TestLintStyle_NoHints(t *testing.T) {
	s := &Skill{
		Name:        "my-skill",
		Description: "Use when you need step-by-step task guidance",
		Body: `## Overview

This skill provides step-by-step guidance for completing tasks.

## When to Use

- When you need structured task completion

## Core Pattern

- First, analyze the input carefully and validate the data
- Then, process the data accordingly and format the results

## Quick Reference

- Analyze → Process → Format → Return

## Common Mistakes

- Skipping validation of input data
`,
	}

	issues := lintStyle(s)
	assert.Empty(t, issues)
}

func TestLintStyle_DescriptionMissingTriggerPrefix(t *testing.T) {
	s := &Skill{
		Name:        "my-skill",
		Description: "A skill that helps with task management",
		Body:        "## Instructions\n\nSome body content here.",
	}

	issues := lintStyle(s)
	hasTriggerHint := false
	for _, i := range issues {
		if i.Field == "description" && i.Severity == SeverityHint && strings.Contains(i.Message, "Use when") {
			hasTriggerHint = true
		}
	}
	assert.True(t, hasTriggerHint, "should hint that description should start with 'Use when'")
}

func TestLintStyle_DescriptionWithTriggerPrefix(t *testing.T) {
	tests := []struct {
		name string
		desc string
	}{
		{"use when", "Use when deploying services to production"},
		{"use for", "Use for formatting code according to team standards"},
		{"use to", "Use to validate API responses before processing"},
		{"trigger when", "Trigger when a new pull request is opened"},
		{"apply when", "Apply when refactoring legacy code modules"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Skill{
				Name:        "my-skill",
				Description: tt.desc,
				Body:        "Short body.",
			}

			issues := lintStyle(s)
			for _, i := range issues {
				if i.Field == "description" && strings.Contains(i.Message, "Use when") {
					t.Errorf("should not hint about trigger prefix for description %q", tt.desc)
				}
			}
		})
	}
}

func TestLintStyle_BodyMissingRecommendedSections(t *testing.T) {
	s := &Skill{
		Name:        "my-skill",
		Description: "Use when you need to do something specific",
		Body:        "## Instructions\n\nThis skill provides step-by-step guidance for completing tasks:\n\n- First, analyze the input carefully and validate the data\n- Then, process the data accordingly and format the results\n- Finally, return the formatted output to the user",
	}

	issues := lintStyle(s)
	hasSectionHint := false
	for _, i := range issues {
		if i.Field == "body" && i.Severity == SeverityHint && strings.Contains(i.Message, "missing recommended sections") {
			hasSectionHint = true
			assert.Contains(t, i.Message, "when to use")
			assert.Contains(t, i.Message, "common mistakes")
		}
	}
	assert.True(t, hasSectionHint, "should hint about missing recommended sections")
}

func TestLintStyle_BodyWithAllRecommendedSections(t *testing.T) {
	s := &Skill{
		Name:        "my-skill",
		Description: "Use when you need structured guidance",
		Body: `## When to Use

- Structured tasks

## Core Pattern

- Step-by-step approach

## Quick Reference

- Key points here

## Common Mistakes

- Forgetting validation
`,
	}

	issues := lintStyle(s)
	for _, i := range issues {
		if i.Field == "body" && strings.Contains(i.Message, "missing recommended sections") {
			t.Errorf("should not hint about missing sections when all are present: %s", i.Message)
		}
	}
}
