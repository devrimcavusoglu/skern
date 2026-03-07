// Package skill defines domain types and logic for Agent Skills.
package skill

import (
	"fmt"
	"regexp"
)

// Scope represents where a skill is stored.
type Scope string

// Scope constants for skill storage locations.
const (
	ScopeUser    Scope = "user"
	ScopeProject Scope = "project"
)

// nameRegex validates skill names: lowercase alphanumeric with hyphens, 1-64 chars.
var nameRegex = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

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

// ValidateName checks that a skill name matches the required pattern.
func ValidateName(name string) error {
	if len(name) == 0 {
		return fmt.Errorf("skill name cannot be empty")
	}
	if len(name) > 64 {
		return fmt.Errorf("skill name cannot exceed 64 characters")
	}
	if !nameRegex.MatchString(name) {
		return fmt.Errorf("skill name %q is invalid: must match [a-z0-9]+(-[a-z0-9]+)* (lowercase alphanumeric with hyphens)", name)
	}
	return nil
}
