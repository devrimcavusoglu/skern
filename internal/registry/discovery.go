package registry

import (
	"path/filepath"
	"sort"
	"strings"

	"github.com/devrimcavusoglu/skern/internal/overlap"
	"github.com/devrimcavusoglu/skern/internal/skill"
)

// DiscoveredSkill pairs a skill with its scope and filesystem path.
type DiscoveredSkill struct {
	Skill skill.Skill `json:"skill"`
	Scope skill.Scope `json:"scope"`
	Path  string      `json:"path"`
}

// ListAll returns skills from both user and project scopes.
func (r *Registry) ListAll() ([]DiscoveredSkill, []ParseWarning, error) {
	var result []DiscoveredSkill
	var allWarnings []ParseWarning

	for _, scope := range []skill.Scope{skill.ScopeUser, skill.ScopeProject} {
		skills, warnings, err := r.List(scope)
		if err != nil {
			return nil, nil, err
		}
		allWarnings = append(allWarnings, warnings...)
		dir := r.scopeDir(scope)
		for _, s := range skills {
			result = append(result, DiscoveredSkill{
				Skill: s,
				Scope: scope,
				Path:  filepath.Join(dir, s.Name),
			})
		}
	}

	return result, allWarnings, nil
}

// ScoredSkill pairs a discovered skill with a relevance score.
type ScoredSkill struct {
	DiscoveredSkill
	Score float64
}

// FuzzySearch finds skills matching the query using fuzzy name and description similarity.
// Results are filtered by the given threshold and sorted by score descending.
func (r *Registry) FuzzySearch(query string, threshold float64) ([]ScoredSkill, error) {
	all, _, err := r.ListAll()
	if err != nil {
		return nil, err
	}

	var results []ScoredSkill
	for _, d := range all {
		score := overlap.ScoreAll(overlap.SearchWeights,
			query, query, query, nil,
			d.Skill.Name, d.Skill.Description, d.Skill.Body, d.Skill.AllowedTools,
		)
		if score >= threshold {
			results = append(results, ScoredSkill{
				DiscoveredSkill: d,
				Score:           score,
			})
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	return results, nil
}

// Search finds skills whose names contain the query (case-insensitive).
func (r *Registry) Search(query string) ([]DiscoveredSkill, error) {
	all, _, err := r.ListAll()
	if err != nil {
		return nil, err
	}

	q := strings.ToLower(query)
	var matches []DiscoveredSkill
	for _, d := range all {
		if strings.Contains(strings.ToLower(d.Skill.Name), q) {
			matches = append(matches, d)
		}
	}

	return matches, nil
}
