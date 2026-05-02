package skill

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScopeThreshold(t *testing.T) {
	tests := []struct {
		name  string
		scope Scope
		want  int
	}{
		{"user scope", ScopeUser, UserScopeSkillCountWarn},
		{"project scope", ScopeProject, ProjectScopeSkillCountWarn},
		{"empty scope falls back to user", Scope(""), UserScopeSkillCountWarn},
		{"unknown scope falls back to user", Scope("weird"), UserScopeSkillCountWarn},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, ScopeThreshold(tt.scope))
		})
	}
}

func TestPlatformThreshold(t *testing.T) {
	tests := []struct {
		name  string
		scope Scope
		want  int
	}{
		{"project scope uses tighter threshold", ScopeProject, ProjectScopeSkillCountWarn},
		{"user scope uses platform threshold", ScopeUser, PlatformSkillCountWarn},
		{"empty scope falls back to platform default", Scope(""), PlatformSkillCountWarn},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, PlatformThreshold(tt.scope))
		})
	}
}

func TestThresholdInvariants(t *testing.T) {
	// Project must always be tighter than (or equal to) user thresholds —
	// the dynamic-loading model assumes project-scope working sets are
	// smaller.
	assert.LessOrEqual(t, ProjectScopeSkillCountWarn, UserScopeSkillCountWarn,
		"project threshold should not exceed user threshold")
	assert.LessOrEqual(t, ProjectScopeSkillCountWarn, PlatformSkillCountWarn,
		"project threshold should not exceed platform threshold")
}
