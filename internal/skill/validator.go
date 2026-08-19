package skill

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
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
	issues = append(issues, validateTags(s.Tags)...)
	issues = append(issues, validateAllowedTools(s.AllowedTools)...)
	issues = append(issues, validateMetadata(s.Metadata)...)
	issues = append(issues, validateInstall(s.Install)...)
	issues = append(issues, lintStyle(s)...)

	return issues
}

// validateInstall checks the `install.exclude` patterns: every entry must be
// a relative, well-formed glob that cannot match SKILL.md.
func validateInstall(c InstallConfig) []ValidationIssue {
	var issues []ValidationIssue
	for i, pattern := range c.Exclude {
		if err := ValidateExcludePattern(pattern); err != nil {
			field := "install.exclude"
			if len(c.Exclude) > 1 {
				field = fmt.Sprintf("install.exclude[%d]", i)
			}
			issues = append(issues, ValidationIssue{
				Field:    field,
				Severity: SeverityError,
				Message:  err.Error(),
			})
		}
	}
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

// validateTags checks each tag against the tag charset rule. Surrounding
// whitespace is tolerated (the tag filters trim it too).
func validateTags(tags []string) []ValidationIssue {
	var issues []ValidationIssue
	for i, tag := range tags {
		if err := ValidateTag(strings.TrimSpace(tag)); err != nil {
			issues = append(issues, ValidationIssue{
				Field:    "tags",
				Severity: SeverityError,
				Message:  fmt.Sprintf("tags[%d]: %s", i, err),
			})
		}
	}
	return issues
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
			continue
		}
		// The file exists in the registry but would be missing from every
		// installed copy — almost certainly a mistake in install.exclude.
		if MatchExclude(s.Install.Exclude, ref) {
			issues = append(issues, ValidationIssue{
				Field:    "install.exclude",
				Severity: SeverityWarning,
				Message:  fmt.Sprintf("referenced file %q is excluded from install by install.exclude", ref),
			})
		}
	}

	// A pattern that matches nothing on disk is usually a typo (or stale
	// after a rename) — worth a warning, not an error.
	for _, pattern := range s.Install.Exclude {
		if ValidateExcludePattern(pattern) != nil {
			continue // reported as an error by Validate
		}
		matched, err := ExcludedFiles(skillDir, []string{pattern})
		if err == nil && len(matched) == 0 {
			issues = append(issues, ValidationIssue{
				Field:    "install.exclude",
				Severity: SeverityWarning,
				Message:  fmt.Sprintf("exclude pattern %q matches no file in the skill directory", pattern),
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

// descriptionTriggerPrefixes are the recommended prefixes for skill descriptions.
// Descriptions should start with a triggering condition (e.g., "Use when...") to
// optimize agent discovery (Claude Search Optimization). Summaries of what the
// skill does cause agents to follow the description shortcut instead of reading
// the full SKILL.md.
var descriptionTriggerPrefixes = []string{
	"use when",
	"use for",
	"use to",
	"trigger when",
	"apply when",
}

// recommendedBodySections lists the sections recommended by the writing-skills
// guidelines. These improve discoverability, scannability, and quality.
var recommendedBodySections = []string{
	"overview",
	"when to use",
	"core pattern",
	"quick reference",
	"common mistakes",
}

// hasMarkdownHeading checks whether the body contains a markdown heading (any level)
// whose text matches the given section name (case-insensitive).
func hasMarkdownHeading(body, section string) bool {
	sectionLower := strings.ToLower(section)
	for _, line := range strings.Split(body, "\n") {
		trimmed := strings.TrimSpace(line)
		// Strip leading '#' characters
		stripped := strings.TrimLeft(trimmed, "#")
		if stripped == trimmed {
			continue // no '#' prefix — not a heading
		}
		// Must have at least one space after '#'s
		if len(stripped) == 0 || !unicode.IsSpace(rune(stripped[0])) {
			continue
		}
		headingText := strings.TrimSpace(stripped)
		if strings.ToLower(headingText) == sectionLower {
			return true
		}
	}
	return false
}

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
	descTrimmed := strings.TrimSpace(s.Description)
	descWords := len(strings.Fields(descTrimmed))
	if descWords > 0 && descWords < lintDescMinWords {
		issues = append(issues, ValidationIssue{
			Field:    "description",
			Severity: SeverityHint,
			Message:  fmt.Sprintf("description has only %d word(s); consider being more specific", descWords),
		})
	}

	// Description should start with a triggering condition (CSO guideline)
	if descWords >= lintDescMinWords {
		descLower := strings.ToLower(descTrimmed)
		hasTriggerPrefix := false
		for _, prefix := range descriptionTriggerPrefixes {
			if strings.HasPrefix(descLower, prefix) {
				hasTriggerPrefix = true
				break
			}
		}
		if !hasTriggerPrefix {
			issues = append(issues, ValidationIssue{
				Field:    "description",
				Severity: SeverityHint,
				Message:  "description should start with a trigger phrase (\"Use when\", \"Use for\", \"Use to\", \"Trigger when\", or \"Apply when\") to describe triggering conditions, not summarize the workflow",
			})
		}
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

	// Body should include recommended sections (writing-skills guideline)
	if bodyWords >= lintBodyMinWords {
		var missing []string
		for _, section := range recommendedBodySections {
			if !hasMarkdownHeading(s.Body, section) {
				missing = append(missing, section)
			}
		}
		if len(missing) > 0 {
			issues = append(issues, ValidationIssue{
				Field:    "body",
				Severity: SeverityHint,
				Message:  fmt.Sprintf("body is missing recommended sections: %s", strings.Join(missing, ", ")),
			})
		}
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
