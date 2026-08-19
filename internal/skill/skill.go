// Package skill defines domain types and logic for Agent Skills.
package skill

import (
	"fmt"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
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

// Author represents the creator of a skill. Extra carries any unmodeled key
// nested under metadata.author (see Metadata.Extra).
type Author struct {
	Name     string         `yaml:"name" json:"name"`
	Type     string         `yaml:"type" json:"type"`
	Platform string         `yaml:"platform,omitempty" json:"platform,omitempty"`
	Extra    map[string]any `yaml:",inline" json:"-"`
}

// ModifiedByEntry records a modification to a skill. Extra carries any
// unmodeled key nested in a metadata.modified-by entry (see Metadata.Extra).
type ModifiedByEntry struct {
	Name     string         `yaml:"name" json:"name"`
	Type     string         `yaml:"type" json:"type"`
	Platform string         `yaml:"platform,omitempty" json:"platform,omitempty"`
	Date     string         `yaml:"date" json:"date"`
	Extra    map[string]any `yaml:",inline" json:"-"`
}

// Metadata holds provenance information for a skill.
//
// Extra carries every metadata.* key skern does not model (for example a
// consumer's own `phases` or `tags` list). Keys land here on parse and are
// written back by WriteManifest, so an extended skill contract survives a
// round trip through skern. The map is nil when a manifest has no such keys.
// Author and ModifiedByEntry carry their own Extra for keys nested one level
// deeper, so no key anywhere under metadata is dropped.
type Metadata struct {
	Author     Author            `yaml:"author" json:"author"`
	Version    string            `yaml:"version" json:"version"`
	ModifiedBy []ModifiedByEntry `yaml:"modified-by,omitempty" json:"modified_by,omitempty"`
	Extra      map[string]any    `yaml:",inline" json:"-"`
}

// InstallConfig is the author-owned `install:` frontmatter block controlling
// how a skill is copied onto a platform. The registry always keeps the whole
// skill directory; these settings only shape the installed copy.
type InstallConfig struct {
	// Exclude lists glob patterns (relative to the skill directory, slash
	// separated) for files and directories that stay in the registry but are
	// not copied on install — evaluation corpora, fixtures, development
	// scratch. See MatchExclude for the matching rules. A single string is
	// accepted in YAML as shorthand for a one-element list.
	Exclude StringList `yaml:"exclude,omitempty" json:"exclude,omitempty"`
	// Extra carries unmodeled keys under install: (see Metadata.Extra).
	Extra map[string]any `yaml:",inline" json:"-"`
}

// IsZero lets yaml omit an empty install block on write.
func (c InstallConfig) IsZero() bool { return len(c.Exclude) == 0 && len(c.Extra) == 0 }

// StringList is a []string that also accepts a bare scalar in YAML
// (`exclude: eval` means `exclude: [eval]`). It always marshals as a
// sequence.
type StringList []string

// UnmarshalYAML implements yaml.Unmarshaler.
func (l *StringList) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind == yaml.ScalarNode {
		var one string
		if err := node.Decode(&one); err != nil {
			return err
		}
		*l = StringList{one}
		return nil
	}
	var many []string
	if err := node.Decode(&many); err != nil {
		return err
	}
	*l = StringList(many)
	return nil
}

// Skill represents an Agent Skill with frontmatter and body content.
//
// Extra carries every top-level frontmatter key skern does not model
// (`compatibility`, `license`, `handoffs`, …). See Metadata.Extra for the
// round-trip guarantee; the same applies here.
type Skill struct {
	Name         string         `yaml:"name" json:"name"`
	Description  string         `yaml:"description" json:"description"`
	Tags         []string       `yaml:"tags,omitempty" json:"tags,omitempty"`
	AllowedTools []string       `yaml:"allowed-tools,omitempty" json:"allowed_tools,omitempty"`
	Metadata     Metadata       `yaml:"metadata" json:"metadata"`
	Install      InstallConfig  `yaml:"install,omitempty" json:"install,omitzero"`
	Extra        map[string]any `yaml:",inline" json:"-"`
	Body         string         `yaml:"-" json:"-"`
}

// tagPartRegex validates one side of a tag: lowercase alphanumeric segments
// joined by hyphens, matching the shape rule nameRegex applies to names.
var tagPartRegex = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

// ValidateTag checks that a tag is either a flat tag ("code-review") or a
// categorical tag ("lang:python", "topic:ci-cd"): lowercase alphanumeric
// segments joined by hyphens, with at most one colon separating category and
// value. Uppercase is rejected so stored tags have a single canonical form;
// the tag *filters* remain case-insensitive, so a query may be typed in any
// case and skills with legacy hand-edited uppercase tags still match.
func ValidateTag(tag string) error {
	parts := strings.Split(tag, ":")
	if len(parts) > 2 {
		return fmt.Errorf("tag %q is invalid: at most one \":\" (separating category and value) is allowed", tag)
	}
	for _, p := range parts {
		if !tagPartRegex.MatchString(p) {
			return fmt.Errorf("tag %q is invalid: tags must be lowercase alphanumeric segments joined by hyphens, optionally as \"category:value\"", tag)
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
