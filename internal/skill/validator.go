package skill

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Severity represents the severity of a validation issue.
type Severity string

// Severity constants for validation issues.
const (
	SeverityError   Severity = "error"
	SeverityWarning Severity = "warning"
	SeverityHint    Severity = "hint"
)

// ValidationIssue represents a single validation problem found in a skill.
type ValidationIssue struct {
	Field    string   `json:"field"`
	Severity Severity `json:"severity"`
	Message  string   `json:"message"`
}

func (v ValidationIssue) String() string {
	return fmt.Sprintf("[%s] %s: %s", v.Severity, v.Field, v.Message)
}

// Validate checks a Skill against all validation rules and returns any issues found.
func Validate(s *Skill) []ValidationIssue {
	var issues []ValidationIssue

	issues = append(issues, validateName(s.Name)...)
	issues = append(issues, validateDescription(s.Description)...)
	issues = append(issues, validateBody(s.Body)...)
	issues = append(issues, validateAllowedTools(s.AllowedTools)...)
	issues = append(issues, validateMetadata(s.Metadata)...)
	issues = append(issues, lintStyle(s)...)

	return issues
}

// HasErrors returns true if any issues have error severity.
func HasErrors(issues []ValidationIssue) bool {
	for _, issue := range issues {
		if issue.Severity == SeverityError {
			return true
		}
	}
	return false
}

func validateName(name string) []ValidationIssue {
	if err := ValidateName(name); err != nil {
		return []ValidationIssue{{
			Field:    "name",
			Severity: SeverityError,
			Message:  err.Error(),
		}}
	}
	return nil
}

func validateDescription(desc string) []ValidationIssue {
	var issues []ValidationIssue

	trimmed := strings.TrimSpace(desc)
	if trimmed == "" {
		issues = append(issues, ValidationIssue{
			Field:    "description",
			Severity: SeverityError,
			Message:  "description is required",
		})
		return issues
	}

	if len(trimmed) > 1024 {
		issues = append(issues, ValidationIssue{
			Field:    "description",
			Severity: SeverityError,
			Message:  fmt.Sprintf("description exceeds 1024 characters (%d)", len(trimmed)),
		})
	}

	return issues
}

func validateBody(body string) []ValidationIssue {
	if strings.TrimSpace(body) == "" {
		return []ValidationIssue{{
			Field:    "body",
			Severity: SeverityError,
			Message:  "SKILL.md body content is required",
		}}
	}
	return nil
}

func validateAllowedTools(tools []string) []ValidationIssue {
	var issues []ValidationIssue
	for i, tool := range tools {
		if strings.TrimSpace(tool) == "" {
			issues = append(issues, ValidationIssue{
				Field:    "allowed-tools",
				Severity: SeverityError,
				Message:  fmt.Sprintf("allowed-tools[%d] is empty", i),
			})
		}
	}
	return issues
}

// ValidateFolder checks that file references in the skill body exist on disk.
// Missing references produce warnings, not errors.
func ValidateFolder(s *Skill, skillDir string) []ValidationIssue {
	refs := ExtractFileReferences(s.Body)
	var issues []ValidationIssue

	for _, ref := range refs {
		path := filepath.Join(skillDir, ref)
		if _, err := os.Stat(path); err != nil {
			issues = append(issues, ValidationIssue{
				Field:    "folder",
				Severity: SeverityWarning,
				Message:  fmt.Sprintf("referenced file %q not found in skill directory", ref),
			})
		}
	}

	return issues
}

// Stylistic lint thresholds.
const (
	lintBodyMinWords = 20
	lintDescMinWords = 3
)

// lintStyle performs stylistic quality checks on a skill.
// Issues use SeverityHint to distinguish from structural errors/warnings.
func lintStyle(s *Skill) []ValidationIssue {
	var issues []ValidationIssue

	// Body too short
	bodyWords := len(strings.Fields(strings.TrimSpace(s.Body)))
	if bodyWords > 0 && bodyWords < lintBodyMinWords {
		issues = append(issues, ValidationIssue{
			Field:    "body",
			Severity: SeverityHint,
			Message:  fmt.Sprintf("body has only %d words; consider adding more detailed instructions", bodyWords),
		})
	}

	// Description too vague (very short)
	descWords := len(strings.Fields(strings.TrimSpace(s.Description)))
	if descWords > 0 && descWords < lintDescMinWords {
		issues = append(issues, ValidationIssue{
			Field:    "description",
			Severity: SeverityHint,
			Message:  fmt.Sprintf("description has only %d word(s); consider being more specific", descWords),
		})
	}

	// Body lacks step-by-step guidance markers
	bodyLower := strings.ToLower(s.Body)
	hasSteps := strings.Contains(bodyLower, "step") ||
		strings.Contains(bodyLower, "1.") ||
		strings.Contains(bodyLower, "- ") ||
		strings.Contains(bodyLower, "* ")
	if bodyWords >= lintBodyMinWords && !hasSteps {
		issues = append(issues, ValidationIssue{
			Field:    "body",
			Severity: SeverityHint,
			Message:  "body lacks step-by-step structure; consider adding numbered steps or bullet points",
		})
	}

	return issues
}

var semverRegex = regexp.MustCompile(`^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)$`)

func validateMetadata(m Metadata) []ValidationIssue {
	var issues []ValidationIssue

	if m.Author.Name == "" {
		issues = append(issues, ValidationIssue{
			Field:    "metadata.author.name",
			Severity: SeverityWarning,
			Message:  "author name is not set",
		})
	}

	if m.Author.Type != "" && m.Author.Type != "human" && m.Author.Type != "agent" {
		issues = append(issues, ValidationIssue{
			Field:    "metadata.author.type",
			Severity: SeverityError,
			Message:  fmt.Sprintf("author type %q is invalid: must be \"human\" or \"agent\"", m.Author.Type),
		})
	}

	if m.Author.Type == "agent" && m.Author.Platform == "" {
		issues = append(issues, ValidationIssue{
			Field:    "metadata.author.platform",
			Severity: SeverityWarning,
			Message:  "author platform should be set when author type is \"agent\"",
		})
	}

	if m.Version != "" {
		if !semverRegex.MatchString(m.Version) {
			issues = append(issues, ValidationIssue{
				Field:    "metadata.version",
				Severity: SeverityWarning,
				Message:  fmt.Sprintf("version %q does not follow semver format (expected X.Y.Z)", m.Version),
			})
		}
	}

	return issues
}
