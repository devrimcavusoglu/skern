package skill

// Capacity thresholds for the number of skills in a registry scope or installed
// on a platform. These are advisory by default; agents and humans see warnings
// when counts approach or exceed the threshold, but operations still proceed
// unless `--enforce-budget` is passed on a mutating command.
//
// Per #52 (D4): per-platform AND per-scope counts are tracked separately. The
// project-scope threshold is intentionally lower than user-scope because a
// project's working set should stay focused.
const (
	// ProjectScopeSkillCountWarn is the count at which project-scope skill
	// registries (.skern/skills/) trigger a warning.
	ProjectScopeSkillCountWarn = 20

	// UserScopeSkillCountWarn is the count at which user-scope skill
	// registries (~/.skern/skills/) trigger a warning.
	UserScopeSkillCountWarn = 50

	// PlatformSkillCountWarn is the count at which a single platform's
	// installed-skills directory triggers a warning. Applies per (platform,
	// scope) pair. Mirrors the user-scope registry threshold by default.
	PlatformSkillCountWarn = 50
)

// ScopeThreshold returns the registry-scope threshold for a given scope.
func ScopeThreshold(scope Scope) int {
	if scope == ScopeProject {
		return ProjectScopeSkillCountWarn
	}
	return UserScopeSkillCountWarn
}

// PlatformThreshold returns the per-platform-per-scope threshold.
//
// Project-scope installations get the lower threshold because a project's
// active skill set on any one platform should be tighter than a user-level
// global set.
func PlatformThreshold(scope Scope) int {
	if scope == ScopeProject {
		return ProjectScopeSkillCountWarn
	}
	return PlatformSkillCountWarn
}
