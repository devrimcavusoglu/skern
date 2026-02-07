package registry

import (
	"path/filepath"
	"strings"

	"github.com/devrimcavusoglu/scribe/internal/skill"
)

// DiscoveredSkill pairs a skill with its scope and filesystem path.
type DiscoveredSkill struct {
	Skill skill.Skill `json:"skill"`
	Scope skill.Scope `json:"scope"`
	Path  string      `json:"path"`
}

// ListAll returns skills from both user and project scopes.
func (r *Registry) ListAll() ([]DiscoveredSkill, error) {
	var result []DiscoveredSkill

	for _, scope := range []skill.Scope{skill.ScopeUser, skill.ScopeProject} {
		skills, err := r.List(scope)
		if err != nil {
			return nil, err
		}
		dir := r.scopeDir(scope)
		for _, s := range skills {
			result = append(result, DiscoveredSkill{
				Skill: s,
				Scope: scope,
				Path:  filepath.Join(dir, s.Name),
			})
		}
	}

	return result, nil
}

// Search finds skills whose names contain the query (case-insensitive).
func (r *Registry) Search(query string) ([]DiscoveredSkill, error) {
	all, err := r.ListAll()
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
