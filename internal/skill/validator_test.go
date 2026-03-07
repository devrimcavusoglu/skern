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
		Description: "A valid skill description",
		Body:        "## Instructions\n\nDo something useful.",
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
		if i.Field == "description" {
			t.Errorf("unexpected description issue: %s", i.Message)
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
