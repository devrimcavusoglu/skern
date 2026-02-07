package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/devrimcavusoglu/scribe/internal/output"
	"github.com/devrimcavusoglu/scribe/internal/platform"
	"github.com/devrimcavusoglu/scribe/internal/registry"
	"github.com/devrimcavusoglu/scribe/internal/skill"
)

// newRegistryFunc creates a Registry. Overridable in tests.
var newRegistryFunc = defaultNewRegistry

// newDetectorFunc creates a platform Detector. Overridable in tests.
var newDetectorFunc = defaultNewDetector

func defaultNewRegistry() (*registry.Registry, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("determining home directory: %w", err)
	}

	userDir := filepath.Join(home, ".scribe", "skills")
	projectDir := filepath.Join(".", ".scribe", "skills")

	return registry.New(userDir, projectDir), nil
}

func defaultNewDetector() (*platform.Detector, error) {
	return platform.NewDetector()
}

// parseScope converts a scope string flag to a skill.Scope.
func parseScope(scopeStr string) (skill.Scope, error) {
	switch scopeStr {
	case "user":
		return skill.ScopeUser, nil
	case "project":
		return skill.ScopeProject, nil
	default:
		return "", &ValidationError{Message: fmt.Sprintf("invalid scope %q: must be \"user\" or \"project\"", scopeStr)}
	}
}

// toSkillResult converts a skill.Skill into an output.SkillResult.
func toSkillResult(s *skill.Skill, scope string, path string) output.SkillResult {
	return output.SkillResult{
		Name:        s.Name,
		Description: strings.TrimSpace(s.Description),
		Version:     s.Metadata.Version,
		Author: output.AuthorResult{
			Name:     s.Metadata.Author.Name,
			Type:     s.Metadata.Author.Type,
			Platform: s.Metadata.Author.Platform,
		},
		Scope:        scope,
		Path:         path,
		AllowedTools: s.AllowedTools,
	}
}

// toDiscoveredSkillResult converts a DiscoveredSkill into an output.SkillResult.
func toDiscoveredSkillResult(d registry.DiscoveredSkill) output.SkillResult {
	return toSkillResult(&d.Skill, string(d.Scope), d.Path)
}

// formatSkillTable formats a list of skills as a text table.
func formatSkillTable(skills []output.SkillResult) string {
	if len(skills) == 0 {
		return "No skills found.\n"
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("%-30s %-10s %-40s\n", "NAME", "SCOPE", "DESCRIPTION"))
	for _, s := range skills {
		desc := s.Description
		if len(desc) > 40 {
			desc = desc[:37] + "..."
		}
		b.WriteString(fmt.Sprintf("%-30s %-10s %-40s\n", s.Name, s.Scope, desc))
	}
	return b.String()
}

// formatSkillShow formats a single skill for detailed display.
func formatSkillShow(s output.SkillResult) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Name:        %s\n", s.Name))
	b.WriteString(fmt.Sprintf("Description: %s\n", s.Description))
	b.WriteString(fmt.Sprintf("Version:     %s\n", s.Version))
	b.WriteString(fmt.Sprintf("Author:      %s (%s)", s.Author.Name, s.Author.Type))
	if s.Author.Platform != "" {
		b.WriteString(fmt.Sprintf(" [%s]", s.Author.Platform))
	}
	b.WriteString("\n")
	if s.Scope != "" {
		b.WriteString(fmt.Sprintf("Scope:       %s\n", s.Scope))
	}
	if s.Path != "" {
		b.WriteString(fmt.Sprintf("Path:        %s\n", s.Path))
	}
	if len(s.AllowedTools) > 0 {
		b.WriteString(fmt.Sprintf("Tools:       %s\n", strings.Join(s.AllowedTools, ", ")))
	}
	return b.String()
}

// formatSearchResults formats search results for text output.
func formatSearchResults(query string, results []output.SkillResult) string {
	if len(results) == 0 {
		return fmt.Sprintf("No skills matching %q found.\n", query)
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("Found %d skill(s) matching %q:\n\n", len(results), query))
	b.WriteString(formatSkillTable(results))
	return b.String()
}

// resolveSkill finds a skill by name, searching the specified scope or both scopes.
func resolveSkill(reg *registry.Registry, name, scopeStr string) (*skill.Skill, string, skill.Scope, error) {
	if scopeStr != "" {
		scope, err := parseScope(scopeStr)
		if err != nil {
			return nil, "", "", err
		}
		s, path, err := reg.Get(name, scope)
		if err != nil {
			return nil, "", "", err
		}
		return s, path, scope, nil
	}

	// Search project first, then user
	for _, scope := range []skill.Scope{skill.ScopeProject, skill.ScopeUser} {
		s, path, err := reg.Get(name, scope)
		if err == nil {
			return s, path, scope, nil
		}
	}

	return nil, "", "", fmt.Errorf("skill %q not found", name)
}
