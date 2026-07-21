// Package skill defines domain types and logic for Agent Skills.
package skill

import (
	"fmt"
	"regexp"
	"strings"
)

// Scope represents where a skill is stored.
type Scope string

// ManifestFile is the canonical filename for a skill manifest.
const ManifestFile = "SKILL.md"

// Scope constants for skill storage locations.
const (
	ScopeUser    Scope = "user"
	ScopeProject Scope = "project"
)

// nameRegex validates skill names: lowercase alphanumeric segments joined by
// hyphens or dots as separators, 1-64 chars. Dots enable namespace-style names
// (e.g. "myorg.bootstrap") so installers can preserve their prefix across
// platforms that map directory names to slash commands.
var nameRegex = regexp.MustCompile(`^[a-z0-9]+([.-][a-z0-9]+)*$`)

// Author represents the creator of a skill.
type Author struct {
	Name     string `yaml:"name" json:"name"`
	Type     string `yaml:"type" json:"type"`
	Platform string `yaml:"platform,omitempty" json:"platform,omitempty"`
}

// ModifiedByEntry records a modification to a skill.
type ModifiedByEntry struct {
	Name     string `yaml:"name" json:"name"`
	Type     string `yaml:"type" json:"type"`
	Platform string `yaml:"platform,omitempty" json:"platform,omitempty"`
	Date     string `yaml:"date" json:"date"`
}

// Metadata holds provenance information for a skill.
type Metadata struct {
	Author     Author            `yaml:"author" json:"author"`
	Version    string            `yaml:"version" json:"version"`
	ModifiedBy []ModifiedByEntry `yaml:"modified-by,omitempty" json:"modified_by,omitempty"`
}

// Skill represents an Agent Skill with frontmatter and body content.
type Skill struct {
	Name         string   `yaml:"name" json:"name"`
	Description  string   `yaml:"description" json:"description"`
	Tags         []string `yaml:"tags,omitempty" json:"tags,omitempty"`
	AllowedTools []string `yaml:"allowed-tools,omitempty" json:"allowed_tools,omitempty"`
	Metadata     Metadata `yaml:"metadata" json:"metadata"`
	Body         string   `yaml:"-" json:"-"`
}

// tagPartRegex validates one side of a tag: alphanumeric segments joined by
// hyphens. Uppercase is allowed because tag matching is case-insensitive.
var tagPartRegex = regexp.MustCompile(`^[a-zA-Z0-9]+(-[a-zA-Z0-9]+)*$`)

// ValidateTag checks that a tag is either a flat tag ("code-review") or a
// categorical tag ("lang:python", "topic:ci-cd"): alphanumeric segments joined
// by hyphens, with at most one colon separating category and value.
func ValidateTag(tag string) error {
	parts := strings.Split(tag, ":")
	if len(parts) > 2 {
		return fmt.Errorf("tag %q is invalid: at most one \":\" (separating category and value) is allowed", tag)
	}
	for _, p := range parts {
		if !tagPartRegex.MatchString(p) {
			return fmt.Errorf("tag %q is invalid: tags may only contain alphanumeric characters and hyphens (segments joined by hyphens, optionally as \"category:value\")", tag)
		}
	}
	return nil
}

// ValidateName checks that a skill name matches the required pattern.
func ValidateName(name string) error {
	if len(name) == 0 {
		return fmt.Errorf("skill name cannot be empty")
	}
	if len(name) > 64 {
		return fmt.Errorf("skill name cannot exceed 64 characters")
	}
	if !nameRegex.MatchString(name) {
		return fmt.Errorf("skill name %q is invalid: must match [a-z0-9]+([.-][a-z0-9]+)* (lowercase alphanumeric segments joined by hyphens or dots)", name)
	}
	return nil
}
