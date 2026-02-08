package overlap

import (
	"testing"

	"github.com/devrimcavusoglu/scribe/internal/skill"
	"github.com/stretchr/testify/assert"
)

func TestLevenshtein(t *testing.T) {
	tests := []struct {
		name string
		a, b string
		want int
	}{
		{"identical", "abc", "abc", 0},
		{"empty both", "", "", 0},
		{"empty a", "", "abc", 3},
		{"empty b", "abc", "", 3},
		{"one char diff", "abc", "adc", 1},
		{"insertion", "abc", "abcd", 1},
		{"deletion", "abcd", "abc", 1},
		{"complete diff", "abc", "xyz", 3},
		{"kitten sitting", "kitten", "sitting", 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, levenshtein(tt.a, tt.b))
		})
	}
}

func TestNameSimilarity(t *testing.T) {
	tests := []struct {
		name   string
		a, b   string
		minSim float64
		maxSim float64
	}{
		{"identical", "code-review", "code-review", 1.0, 1.0},
		{"very similar", "code-review", "code-reviewer", 0.7, 1.0},
		{"shared prefix", "code-review", "code-format", 0.5, 0.8},
		{"contained", "review", "code-review", 0.5, 1.0},
		{"completely different", "deploy-app", "code-review", 0.0, 0.4},
		{"empty strings", "", "", 0.0, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sim := NameSimilarity(tt.a, tt.b)
			assert.GreaterOrEqual(t, sim, tt.minSim, "similarity %f below minimum %f", sim, tt.minSim)
			assert.LessOrEqual(t, sim, tt.maxSim, "similarity %f above maximum %f", sim, tt.maxSim)
		})
	}
}

func TestNameSimilarity_CaseInsensitive(t *testing.T) {
	sim := NameSimilarity("Code-Review", "code-review")
	assert.Equal(t, 1.0, sim)
}

func TestCheck_NoExisting(t *testing.T) {
	matches := Check("my-skill", "does things", nil, nil)
	assert.Empty(t, matches)
}

func TestCheck_NoMatches(t *testing.T) {
	existing := []skill.Skill{
		{Name: "deploy-app", Description: "deploys applications"},
		{Name: "lint-code", Description: "runs code linters"},
	}
	scopes := []skill.Scope{skill.ScopeUser, skill.ScopeUser}

	matches := Check("format-docs", "formats documentation", existing, scopes)
	assert.Empty(t, matches)
}

func TestCheck_FindsSimilarNames(t *testing.T) {
	existing := []skill.Skill{
		{Name: "code-review", Description: "reviews code"},
		{Name: "deploy-app", Description: "deploys applications"},
	}
	scopes := []skill.Scope{skill.ScopeUser, skill.ScopeUser}

	matches := Check("code-reviewer", "reviews code changes", existing, scopes)
	// Should find code-review as similar
	found := false
	for _, m := range matches {
		if m.Name == "code-review" {
			found = true
			assert.GreaterOrEqual(t, m.Score, WarnThreshold)
		}
	}
	assert.True(t, found, "expected to find code-review as similar")
}

func TestCheck_IdenticalName(t *testing.T) {
	existing := []skill.Skill{
		{Name: "my-skill", Description: "does things"},
	}
	scopes := []skill.Scope{skill.ScopeUser}

	matches := Check("my-skill", "does things", existing, scopes)
	assert.NotEmpty(t, matches)
	assert.Equal(t, "my-skill", matches[0].Name)
	// Identical name + same description scores 0.8 (0.5*1.0 + 0.3*1.0 + 0.2*0.0)
	// since no tools overlap occurs when candidate has no tools
	assert.GreaterOrEqual(t, matches[0].Score, WarnThreshold)
}

func TestMaxScore(t *testing.T) {
	matches := []Match{
		{Name: "a", Score: 0.5},
		{Name: "b", Score: 0.9},
		{Name: "c", Score: 0.7},
	}
	assert.Equal(t, 0.9, MaxScore(matches))
}

func TestMaxScore_Empty(t *testing.T) {
	assert.Equal(t, 0.0, MaxScore(nil))
}

func TestShouldBlock(t *testing.T) {
	tests := []struct {
		name    string
		matches []Match
		want    bool
	}{
		{"empty", nil, false},
		{"below threshold", []Match{{Score: 0.5}}, false},
		{"at threshold", []Match{{Score: 0.9}}, true},
		{"above threshold", []Match{{Score: 0.95}}, true},
		{"mixed", []Match{{Score: 0.5}, {Score: 0.95}}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, ShouldBlock(tt.matches))
		})
	}
}

func TestCommonPrefixLen(t *testing.T) {
	tests := []struct {
		name string
		a, b string
		want int
	}{
		{"identical", "abc", "abc", 3},
		{"prefix", "abc", "abd", 2},
		{"no common", "abc", "xyz", 0},
		{"empty", "", "abc", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, commonPrefixLen(tt.a, tt.b))
		})
	}
}

func TestCommonSuffixLen(t *testing.T) {
	tests := []struct {
		name string
		a, b string
		want int
	}{
		{"identical", "abc", "abc", 3},
		{"suffix", "abc", "xbc", 2},
		{"no common", "abc", "xyz", 0},
		{"empty", "", "abc", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, commonSuffixLen(tt.a, tt.b))
		})
	}
}
