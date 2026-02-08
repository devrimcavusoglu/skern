package skill

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSuggestName(t *testing.T) {
	tests := []struct {
		name  string
		query string
		want  string
	}{
		{"simple", "format go source code", "format-go-source-code"},
		{"with special chars", "lint & format markdown!", "lint-format-markdown"},
		{"uppercase", "Run Database Migrations", "run-database-migrations"},
		{"extra spaces", "  build  docker  image  ", "build-docker-image"},
		{"already slugified", "code-review", "code-review"},
		{"with underscores", "run_unit_tests", "run-unit-tests"},
		{"empty string", "", ""},
		{"only special chars", "!@#$%", ""},
		{"single word", "deploy", "deploy"},
		{"with numbers", "python3 linter", "python3-linter"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SuggestName(tt.query)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestSuggestName_TruncatesLongNames(t *testing.T) {
	// Create a query that would produce a name longer than 64 chars
	long := "this is a very long query string that should definitely be truncated to fit within the sixty four character limit"
	result := SuggestName(long)
	assert.LessOrEqual(t, len(result), 64)
	assert.NotEmpty(t, result)
	// Should not end with a hyphen
	assert.NotEqual(t, "-", string(result[len(result)-1]))
}

func TestSuggestName_ValidResult(t *testing.T) {
	// Any non-empty result should pass ValidateName
	queries := []string{
		"format go code",
		"run tests",
		"deploy to production",
		"lint markdown files",
	}
	for _, q := range queries {
		result := SuggestName(q)
		if result != "" {
			assert.NoError(t, ValidateName(result), "SuggestName(%q) = %q should be valid", q, result)
		}
	}
}
