package registry

import (
	"testing"

	"github.com/devrimcavusoglu/scribe/internal/skill"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFuzzySearch_NoSkills(t *testing.T) {
	reg := New(t.TempDir(), t.TempDir())

	results, err := reg.FuzzySearch("anything", 0.3)
	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestFuzzySearch_FindsByName(t *testing.T) {
	reg := New(t.TempDir(), t.TempDir())

	// Create skills
	_, err := reg.Create(skill.NewSkill("code-review", "reviews code for errors", "", "", ""), skill.ScopeUser)
	require.NoError(t, err)
	_, err = reg.Create(skill.NewSkill("deploy-app", "deploys applications to production", "", "", ""), skill.ScopeUser)
	require.NoError(t, err)

	results, err := reg.FuzzySearch("code-review", 0.3)
	require.NoError(t, err)
	require.NotEmpty(t, results)
	assert.Equal(t, "code-review", results[0].Skill.Name)
	assert.GreaterOrEqual(t, results[0].Score, 0.3)
}

func TestFuzzySearch_FindsByDescription(t *testing.T) {
	reg := New(t.TempDir(), t.TempDir())

	_, err := reg.Create(skill.NewSkill("go-fmt", "format go source files using gofmt", "", "", ""), skill.ScopeUser)
	require.NoError(t, err)
	_, err = reg.Create(skill.NewSkill("deploy-app", "deploys applications to production", "", "", ""), skill.ScopeUser)
	require.NoError(t, err)

	results, err := reg.FuzzySearch("format go source code", 0.1)
	require.NoError(t, err)
	require.NotEmpty(t, results)
	// The go-fmt skill should score highest due to description overlap
	assert.Equal(t, "go-fmt", results[0].Skill.Name)
}

func TestFuzzySearch_RespectsThreshold(t *testing.T) {
	reg := New(t.TempDir(), t.TempDir())

	_, err := reg.Create(skill.NewSkill("deploy-app", "deploys applications", "", "", ""), skill.ScopeUser)
	require.NoError(t, err)

	// Very high threshold should filter out most results
	results, err := reg.FuzzySearch("completely unrelated query", 0.99)
	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestFuzzySearch_SortedByScore(t *testing.T) {
	reg := New(t.TempDir(), t.TempDir())

	_, err := reg.Create(skill.NewSkill("code-review", "reviews code for errors", "", "", ""), skill.ScopeUser)
	require.NoError(t, err)
	_, err = reg.Create(skill.NewSkill("code-format", "formats code files", "", "", ""), skill.ScopeUser)
	require.NoError(t, err)
	_, err = reg.Create(skill.NewSkill("deploy-app", "deploys applications", "", "", ""), skill.ScopeUser)
	require.NoError(t, err)

	results, err := reg.FuzzySearch("code review", 0.1)
	require.NoError(t, err)

	// Verify results are sorted by score descending
	for i := 1; i < len(results); i++ {
		assert.GreaterOrEqual(t, results[i-1].Score, results[i].Score,
			"results should be sorted by score descending")
	}
}

func TestFuzzySearch_BothScopes(t *testing.T) {
	reg := New(t.TempDir(), t.TempDir())

	_, err := reg.Create(skill.NewSkill("user-lint", "lints code", "", "", ""), skill.ScopeUser)
	require.NoError(t, err)
	_, err = reg.Create(skill.NewSkill("proj-lint", "lints project code", "", "", ""), skill.ScopeProject)
	require.NoError(t, err)

	results, err := reg.FuzzySearch("lint", 0.1)
	require.NoError(t, err)

	// Should find skills from both scopes
	scopes := map[skill.Scope]bool{}
	for _, r := range results {
		scopes[r.Scope] = true
	}
	assert.True(t, scopes[skill.ScopeUser], "should include user-scope skills")
	assert.True(t, scopes[skill.ScopeProject], "should include project-scope skills")
}
