package overlap

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDescriptionSimilarity(t *testing.T) {
	tests := []struct {
		name   string
		a, b   string
		minSim float64
		maxSim float64
	}{
		{"identical", "reviews code for errors", "reviews code for errors", 1.0, 1.0},
		{"similar", "reviews code for errors", "checks code for bugs", 0.1, 0.6},
		{"different", "deploys applications", "formats documentation", 0.0, 0.1},
		{"empty a", "", "reviews code", 0.0, 0.0},
		{"empty b", "reviews code", "", 0.0, 0.0},
		{"both empty", "", "", 0.0, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sim := descriptionSimilarity(tt.a, tt.b)
			assert.GreaterOrEqual(t, sim, tt.minSim, "similarity %f below min %f", sim, tt.minSim)
			assert.LessOrEqual(t, sim, tt.maxSim, "similarity %f above max %f", sim, tt.maxSim)
		})
	}
}

func TestToolsOverlap(t *testing.T) {
	tests := []struct {
		name string
		a, b []string
		want float64
	}{
		{"identical", []string{"Bash", "Read"}, []string{"Bash", "Read"}, 1.0},
		{"partial", []string{"Bash", "Read"}, []string{"Bash", "Write"}, 1.0 / 3.0},
		{"no overlap", []string{"Bash"}, []string{"Write"}, 0.0},
		{"empty a", nil, []string{"Bash"}, 0.0},
		{"empty b", []string{"Bash"}, nil, 0.0},
		{"both empty", nil, nil, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toolsOverlap(tt.a, tt.b)
			assert.InDelta(t, tt.want, result, 0.01)
		})
	}
}

func TestExtractKeywords(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{"simple", "reviews code for errors", []string{"reviews", "code", "errors"}},
		{"with separators", "code-review/linting", []string{"code", "review", "linting"}},
		{"filters stop words", "a the and is to", nil},
		{"short words", "a b c de fg", []string{"de", "fg"}},
		{"empty", "", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractKeywords(tt.input)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestScore(t *testing.T) {
	// Identical skill scores 0.8 (name 0.5 + desc 0.3, tools 0.0 since candidate has no tools)
	score := Score("code-review", "reviews code for errors", "code-review", "reviews code for errors", nil)
	assert.GreaterOrEqual(t, score, 0.75)

	// Completely different should score low
	score = Score("deploy-app", "deploys to production", "format-docs", "formats documentation files", nil)
	assert.LessOrEqual(t, score, 0.3)
}

func TestScoreWithTools(t *testing.T) {
	score := ScoreWithTools(
		"code-review", "reviews code", []string{"Bash", "Read"},
		"code-review", "reviews code", []string{"Bash", "Read"},
	)
	assert.GreaterOrEqual(t, score, 0.9)
}

func TestScore_WeightsSum(t *testing.T) {
	// Verify weights sum to 1.0
	assert.InDelta(t, 1.0, nameWeight+descriptionWeight+toolsWeight, 0.001)
}
