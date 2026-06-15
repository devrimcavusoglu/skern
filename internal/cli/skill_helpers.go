package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/devrimcavusoglu/skern/internal/output"
	"github.com/devrimcavusoglu/skern/internal/platform"
	"github.com/devrimcavusoglu/skern/internal/registry"
	"github.com/devrimcavusoglu/skern/internal/skill"
)

func defaultNewRegistry() (*registry.Registry, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("determining home directory: %w", err)
	}

	userDir := filepath.Join(home, ".skern", "skills")
	projectDir := filepath.Join(".", ".skern", "skills")

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
	var modifiedBy []output.ModifiedByResult
	for _, m := range s.Metadata.ModifiedBy {
		modifiedBy = append(modifiedBy, output.ModifiedByResult{
			Name:     m.Name,
			Type:     m.Type,
			Platform: m.Platform,
			Date:     m.Date,
		})
	}

	return output.SkillResult{
		Name:        s.Name,
		Description: strings.TrimSpace(s.Description),
		Version:     s.Metadata.Version,
		Author: output.AuthorResult{
			Name:     s.Metadata.Author.Name,
			Type:     s.Metadata.Author.Type,
			Platform: s.Metadata.Author.Platform,
		},
		Tags:         s.Tags,
		Scope:        scope,
		Path:         path,
		AllowedTools: s.AllowedTools,
		ModifiedBy:   modifiedBy,
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
	fmt.Fprintf(&b, "%-30s %-10s %-7s %-40s\n", "NAME", "SCOPE", "FILES", "DESCRIPTION")
	for _, s := range skills {
		desc := s.Description
		if len(desc) > 40 {
			desc = desc[:37] + "..."
		}
		fileCount := "-"
		if len(s.Files) > 0 {
			fileCount = fmt.Sprintf("%d", len(s.Files))
		}
		fmt.Fprintf(&b, "%-30s %-10s %-7s %-40s\n", s.Name, s.Scope, fileCount, desc)
	}
	return b.String()
}

// formatSkillShow formats a single skill for detailed display.
func formatSkillShow(s output.SkillResult) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Name:        %s\n", s.Name)
	fmt.Fprintf(&b, "Description: %s\n", s.Description)
	fmt.Fprintf(&b, "Version:     %s\n", s.Version)
	fmt.Fprintf(&b, "Author:      %s (%s)", s.Author.Name, s.Author.Type)
	if s.Author.Platform != "" {
		fmt.Fprintf(&b, " [%s]", s.Author.Platform)
	}
	b.WriteString("\n")
	if s.Scope != "" {
		fmt.Fprintf(&b, "Scope:       %s\n", s.Scope)
	}
	if s.Path != "" {
		fmt.Fprintf(&b, "Path:        %s\n", s.Path)
	}
	if len(s.Tags) > 0 {
		fmt.Fprintf(&b, "Tags:        %s\n", strings.Join(s.Tags, ", "))
	}
	if len(s.AllowedTools) > 0 {
		fmt.Fprintf(&b, "Tools:       %s\n", strings.Join(s.AllowedTools, ", "))
	}
	if len(s.Files) > 0 {
		b.WriteString("Files:\n")
		for _, f := range s.Files {
			fmt.Fprintf(&b, "  - %s\n", f)
		}
	}
	if len(s.ModifiedBy) > 0 {
		b.WriteString("Modified-by:\n")
		for _, m := range s.ModifiedBy {
			entry := fmt.Sprintf("  - %s (%s)", m.Name, m.Type)
			if m.Platform != "" {
				entry += fmt.Sprintf(" [%s]", m.Platform)
			}
			if m.Date != "" {
				entry += fmt.Sprintf(" on %s", m.Date)
			}
			b.WriteString(entry + "\n")
		}
	}
	return b.String()
}

// formatDedupHints formats duplicate hints for text output.
func formatDedupHints(hints []output.DuplicateHint) string {
	var b strings.Builder
	b.WriteString("\nPotential duplicates:\n")
	for _, h := range hints {
		fmt.Fprintf(&b, "  - %s <-> %s (score: %.2f)\n", h.SkillA, h.SkillB, h.Score)
	}
	return b.String()
}

// formatParseWarnings formats parse warnings for text output.
func formatParseWarnings(warnings []registry.ParseWarning) string {
	var b strings.Builder
	b.WriteString("\nWarning: some skill directories could not be parsed:\n")
	for _, w := range warnings {
		fmt.Fprintf(&b, "  - %s: %s\n", w.Name, w.Error)
	}
	return b.String()
}

// formatSearchResults formats search results for text output.
func formatSearchResults(query string, results []output.SkillResult) string {
	if len(results) == 0 {
		return fmt.Sprintf("No skills matching %q found.\n", query)
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Found %d skill(s) matching %q:\n\n", len(results), query)
	b.WriteString(formatSkillTable(results))
	return b.String()
}

// hasTag checks if a tag list contains the given tag (case-insensitive).
func hasTag(tags []string, tag string) bool {
	t := strings.ToLower(tag)
	for _, v := range tags {
		if strings.ToLower(v) == t {
			return true
		}
	}
	return false
}

// parseCategoryFilters converts repeated --category flags into a namespace ->
// requested-values map. Each flag value has the form "category:value" and may
// carry a comma-separated value list ("lang:python,go"). Namespaces and values
// are lowercased so matching is case-insensitive, consistent with hasTag.
//
// Malformed input is a ValidationError (exit code 2): a value with no colon,
// an empty category name, or an empty value. Flat tags (no colon) are a
// different surface — they belong to --tag, not --category.
func parseCategoryFilters(raw []string) (map[string][]string, error) {
	filters := map[string][]string{}
	for _, entry := range raw {
		ns, valStr, found := strings.Cut(entry, ":")
		if !found {
			return nil, &ValidationError{Message: fmt.Sprintf("invalid --category %q: expected format \"category:value\"", entry)}
		}
		ns = strings.ToLower(strings.TrimSpace(ns))
		if ns == "" {
			return nil, &ValidationError{Message: fmt.Sprintf("invalid --category %q: category name must not be empty", entry)}
		}
		for _, v := range strings.Split(valStr, ",") {
			v = strings.ToLower(strings.TrimSpace(v))
			if v == "" {
				return nil, &ValidationError{Message: fmt.Sprintf("invalid --category %q: value must not be empty", entry)}
			}
			filters[ns] = append(filters[ns], v)
		}
	}
	return filters, nil
}

// matchesCategories reports whether a skill's tags satisfy the requested
// category filters. Semantics: OR within a category (any requested value
// matches), AND across categories (every requested category must be satisfied).
//
// A skill is "category-absent" for a namespace when none of its tags carry that
// namespace. By default an absent category fails the match (strict). When
// includeUntagged is set, an absent category is treated as "applies to all" and
// passes — but a category the skill *does* declare must still match a requested
// value. An empty filter set matches everything.
func matchesCategories(tags []string, filters map[string][]string, includeUntagged bool) bool {
	if len(filters) == 0 {
		return true
	}

	// Index the skill's categorical tags as namespace -> set of values.
	// Flat tags (no colon) and malformed tags (empty namespace/value) are
	// not categorical and are ignored here.
	skillCats := map[string]map[string]bool{}
	for _, t := range tags {
		ns, val, found := strings.Cut(t, ":")
		if !found {
			continue
		}
		ns = strings.ToLower(strings.TrimSpace(ns))
		val = strings.ToLower(strings.TrimSpace(val))
		if ns == "" || val == "" {
			continue
		}
		if skillCats[ns] == nil {
			skillCats[ns] = map[string]bool{}
		}
		skillCats[ns][val] = true
	}

	for ns, wanted := range filters {
		have, present := skillCats[ns]
		if !present {
			if includeUntagged {
				continue
			}
			return false
		}
		matched := false
		for _, w := range wanted {
			if have[w] {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}
	return true
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

	return nil, "", "", fmt.Errorf("skill %q not found (run 'skern skill list' to see available skills)", name)
}
