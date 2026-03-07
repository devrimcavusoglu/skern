package cli

import (
	"encoding/json"
	"testing"

	"github.com/devrimcavusoglu/skern/internal/output"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSkillRecommend_Create_NoSkills(t *testing.T) {
	cc := testRegistry(t)

	out, err := runCmd(t, cc, "skill", "recommend", "format go source code", "--json")
	require.NoError(t, err)

	var result output.SkillRecommendResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	assert.Equal(t, "format go source code", result.Query)
	assert.Equal(t, output.RecommendCreate, result.Action)
	assert.Equal(t, "No existing skills match your needs.", result.Reason)
	assert.Equal(t, "format-go-source-code", result.SuggestedName)
	assert.Equal(t, 0, result.Count)
}

func TestSkillRecommend_Reuse(t *testing.T) {
	cc := testRegistry(t)

	// Create a skill — name and description both contain "code-review"
	_, err := runCmd(t, cc, "skill", "create", "code-review", "--description", "code review")
	require.NoError(t, err)

	// Query that exactly matches both name and description
	out, err := runCmd(t, cc, "skill", "recommend", "code-review", "--json")
	require.NoError(t, err)

	var result output.SkillRecommendResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	assert.Equal(t, output.RecommendReuse, result.Action,
		"expected reuse but got %s (score: %.4f)", result.Action, func() float64 {
			if len(result.Matches) > 0 {
				return result.Matches[0].Score
			}
			return 0
		}())
	assert.NotEmpty(t, result.Matches)
	assert.Equal(t, "code-review", result.Matches[0].Name)
}

func TestSkillRecommend_Create_LowMatch(t *testing.T) {
	cc := testRegistry(t)

	// Create a skill that won't match the query
	_, err := runCmd(t, cc, "skill", "create", "deploy-app", "--description", "Deploys applications to production")
	require.NoError(t, err)

	out, err := runCmd(t, cc, "skill", "recommend", "format go source code", "--json")
	require.NoError(t, err)

	var result output.SkillRecommendResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	assert.Equal(t, output.RecommendCreate, result.Action)
}

func TestSkillRecommend_Threshold(t *testing.T) {
	cc := testRegistry(t)

	_, err := runCmd(t, cc, "skill", "create", "go-fmt", "--description", "Format Go source files")
	require.NoError(t, err)

	// Very high threshold should yield CREATE
	out, err := runCmd(t, cc, "skill", "recommend", "format go code", "--threshold", "0.99", "--json")
	require.NoError(t, err)

	var result output.SkillRecommendResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	assert.Equal(t, output.RecommendCreate, result.Action)
	assert.Equal(t, 0, result.Count)
}

func TestSkillRecommend_ScopedSearch(t *testing.T) {
	cc := testRegistry(t)

	_, err := runCmd(t, cc, "skill", "create", "user-lint", "--description", "Lints user code", "--scope", "user")
	require.NoError(t, err)
	_, err = runCmd(t, cc, "skill", "create", "proj-lint", "--description", "Lints project code", "--scope", "project")
	require.NoError(t, err)

	// Search only user scope
	out, err := runCmd(t, cc, "skill", "recommend", "lint", "--scope", "user", "--json")
	require.NoError(t, err)

	var result output.SkillRecommendResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	// Should only contain user-scoped skill
	for _, m := range result.Matches {
		assert.Equal(t, "user", m.Scope)
	}
}

func TestSkillRecommend_TextOutput(t *testing.T) {
	cc := testRegistry(t)

	out, err := runCmd(t, cc, "skill", "recommend", "deploy to production")
	require.NoError(t, err)
	assert.Contains(t, out, "Recommendation:")
	assert.Contains(t, out, "CREATE")
	assert.Contains(t, out, "deploy to production")
}

func TestSkillRecommend_TextOutput_WithMatch(t *testing.T) {
	cc := testRegistry(t)

	_, err := runCmd(t, cc, "skill", "create", "code-review", "--description", "code review")
	require.NoError(t, err)

	out, err := runCmd(t, cc, "skill", "recommend", "code review")
	require.NoError(t, err)
	assert.Contains(t, out, "Recommendation:")
	assert.Contains(t, out, "REUSE")
}

func TestSkillRecommend_NameOverride(t *testing.T) {
	cc := testRegistry(t)

	out, err := runCmd(t, cc, "skill", "recommend", "format go source code", "--name", "gofmt-runner", "--json")
	require.NoError(t, err)

	var result output.SkillRecommendResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	assert.Equal(t, output.RecommendCreate, result.Action)
	assert.Equal(t, "gofmt-runner", result.SuggestedName, "agent-provided name should override auto-generated")
}

func TestSkillRecommend_NameOverride_TextOutput(t *testing.T) {
	cc := testRegistry(t)

	out, err := runCmd(t, cc, "skill", "recommend", "format go source code", "--name", "gofmt-runner")
	require.NoError(t, err)
	assert.Contains(t, out, "gofmt-runner")
	assert.Contains(t, out, "skern skill create")
}

func TestSkillRecommend_NameOverride_InvalidName(t *testing.T) {
	cc := testRegistry(t)

	_, err := runCmd(t, cc, "skill", "recommend", "test query", "--name", "INVALID_NAME")
	assert.Error(t, err)
}

func TestSkillRecommend_InvalidScope(t *testing.T) {
	cc := testRegistry(t)

	_, err := runCmd(t, cc, "skill", "recommend", "test", "--scope", "invalid")
	assert.Error(t, err)
}

func TestSkillRecommend_MissingArg(t *testing.T) {
	cc := testRegistry(t)

	_, err := runCmd(t, cc, "skill", "recommend")
	assert.Error(t, err)
}
