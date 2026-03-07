package registry

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/devrimcavusoglu/skern/internal/skill"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestRegistry(t *testing.T) *Registry {
	t.Helper()
	userDir := filepath.Join(t.TempDir(), "user-skills")
	projectDir := filepath.Join(t.TempDir(), "project-skills")
	return New(userDir, projectDir)
}

func TestRegistry_Create(t *testing.T) {
	reg := newTestRegistry(t)
	s := skill.NewSkill("test-skill", "A test skill.", "alice", "human", "")

	path, err := reg.Create(s, skill.ScopeUser)
	require.NoError(t, err)
	assert.DirExists(t, path)
	assert.FileExists(t, filepath.Join(path, "SKILL.md"))
}

func TestRegistry_Create_InvalidName(t *testing.T) {
	reg := newTestRegistry(t)
	s := skill.NewSkill("INVALID", "Desc.", "", "", "")

	_, err := reg.Create(s, skill.ScopeUser)
	assert.Error(t, err)
}

func TestRegistry_Create_Duplicate(t *testing.T) {
	reg := newTestRegistry(t)
	s := skill.NewSkill("dup-skill", "Desc.", "", "", "")

	_, err := reg.Create(s, skill.ScopeUser)
	require.NoError(t, err)

	_, err = reg.Create(s, skill.ScopeUser)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestRegistry_Get(t *testing.T) {
	reg := newTestRegistry(t)
	s := skill.NewSkill("get-skill", "A getter.", "bob", "human", "")

	_, err := reg.Create(s, skill.ScopeUser)
	require.NoError(t, err)

	got, path, err := reg.Get("get-skill", skill.ScopeUser)
	require.NoError(t, err)
	assert.Equal(t, "get-skill", got.Name)
	assert.NotEmpty(t, path)
}

func TestRegistry_Get_NotFound(t *testing.T) {
	reg := newTestRegistry(t)

	_, _, err := reg.Get("nonexistent", skill.ScopeUser)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestRegistry_Remove(t *testing.T) {
	reg := newTestRegistry(t)
	s := skill.NewSkill("remove-me", "To be removed.", "", "", "")

	path, err := reg.Create(s, skill.ScopeUser)
	require.NoError(t, err)
	assert.DirExists(t, path)

	err = reg.Remove("remove-me", skill.ScopeUser)
	require.NoError(t, err)
	assert.NoDirExists(t, path)
}

func TestRegistry_Remove_NotFound(t *testing.T) {
	reg := newTestRegistry(t)

	err := reg.Remove("nonexistent", skill.ScopeUser)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestRegistry_List(t *testing.T) {
	reg := newTestRegistry(t)

	// Create two skills
	s1 := skill.NewSkill("skill-a", "First.", "", "", "")
	s2 := skill.NewSkill("skill-b", "Second.", "", "", "")

	_, err := reg.Create(s1, skill.ScopeUser)
	require.NoError(t, err)
	_, err = reg.Create(s2, skill.ScopeUser)
	require.NoError(t, err)

	skills, _, err := reg.List(skill.ScopeUser)
	require.NoError(t, err)
	assert.Len(t, skills, 2)
}

func TestRegistry_List_Empty(t *testing.T) {
	reg := newTestRegistry(t)

	skills, _, err := reg.List(skill.ScopeUser)
	require.NoError(t, err)
	assert.Empty(t, skills)
}

func TestRegistry_Exists(t *testing.T) {
	reg := newTestRegistry(t)
	s := skill.NewSkill("exist-skill", "Exists.", "", "", "")

	_, err := reg.Create(s, skill.ScopeUser)
	require.NoError(t, err)

	assert.True(t, reg.Exists("exist-skill", skill.ScopeUser))
	assert.False(t, reg.Exists("nonexistent", skill.ScopeUser))
}

func TestRegistry_ProjectScope(t *testing.T) {
	reg := newTestRegistry(t)
	s := skill.NewSkill("project-skill", "In project.", "", "", "")

	path, err := reg.Create(s, skill.ScopeProject)
	require.NoError(t, err)
	assert.DirExists(t, path)

	got, _, err := reg.Get("project-skill", skill.ScopeProject)
	require.NoError(t, err)
	assert.Equal(t, "project-skill", got.Name)

	// Should not exist in user scope
	assert.False(t, reg.Exists("project-skill", skill.ScopeUser))
}

// Discovery tests

func TestRegistry_ListAll(t *testing.T) {
	reg := newTestRegistry(t)

	s1 := skill.NewSkill("user-skill", "User scope.", "", "", "")
	s2 := skill.NewSkill("project-skill", "Project scope.", "", "", "")

	_, err := reg.Create(s1, skill.ScopeUser)
	require.NoError(t, err)
	_, err = reg.Create(s2, skill.ScopeProject)
	require.NoError(t, err)

	all, _, err := reg.ListAll()
	require.NoError(t, err)
	assert.Len(t, all, 2)

	// Verify scopes are correct
	scopes := map[string]skill.Scope{}
	for _, d := range all {
		scopes[d.Skill.Name] = d.Scope
	}
	assert.Equal(t, skill.ScopeUser, scopes["user-skill"])
	assert.Equal(t, skill.ScopeProject, scopes["project-skill"])
}

func TestRegistry_ListAll_Empty(t *testing.T) {
	reg := newTestRegistry(t)

	all, _, err := reg.ListAll()
	require.NoError(t, err)
	assert.Empty(t, all)
}

func TestRegistry_Search(t *testing.T) {
	reg := newTestRegistry(t)

	s1 := skill.NewSkill("code-review", "Reviews code.", "", "", "")
	s2 := skill.NewSkill("code-format", "Formats code.", "", "", "")
	s3 := skill.NewSkill("deploy-app", "Deploys app.", "", "", "")

	for _, s := range []*skill.Skill{s1, s2, s3} {
		_, err := reg.Create(s, skill.ScopeUser)
		require.NoError(t, err)
	}

	results, err := reg.Search("code")
	require.NoError(t, err)
	assert.Len(t, results, 2)
}

func TestRegistry_Search_CaseInsensitive(t *testing.T) {
	reg := newTestRegistry(t)

	s := skill.NewSkill("my-skill", "A skill.", "", "", "")
	_, err := reg.Create(s, skill.ScopeUser)
	require.NoError(t, err)

	results, err := reg.Search("MY-SKILL")
	require.NoError(t, err)
	assert.Len(t, results, 1)
}

func TestRegistry_Search_NoMatch(t *testing.T) {
	reg := newTestRegistry(t)

	s := skill.NewSkill("my-skill", "A skill.", "", "", "")
	_, err := reg.Create(s, skill.ScopeUser)
	require.NoError(t, err)

	results, err := reg.Search("nonexistent")
	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestRegistry_Search_MultiScope(t *testing.T) {
	userDir := filepath.Join(t.TempDir(), "user-skills")
	projectDir := filepath.Join(t.TempDir(), "project-skills")
	reg := New(userDir, projectDir)

	s1 := skill.NewSkill("test-user", "User.", "", "", "")
	s2 := skill.NewSkill("test-project", "Project.", "", "", "")

	_, err := reg.Create(s1, skill.ScopeUser)
	require.NoError(t, err)
	_, err = reg.Create(s2, skill.ScopeProject)
	require.NoError(t, err)

	results, err := reg.Search("test")
	require.NoError(t, err)
	assert.Len(t, results, 2)
}

func TestRegistry_List_SkipsInvalidWithWarning(t *testing.T) {
	reg := newTestRegistry(t)

	// Create a valid skill
	s := skill.NewSkill("valid-skill", "Valid.", "", "", "")
	_, err := reg.Create(s, skill.ScopeUser)
	require.NoError(t, err)

	// Create an invalid directory (no SKILL.md)
	invalidDir := filepath.Join(reg.userDir, "invalid-dir")
	require.NoError(t, os.MkdirAll(invalidDir, 0o755))

	skills, warnings, err := reg.List(skill.ScopeUser)
	require.NoError(t, err)
	assert.Len(t, skills, 1)
	assert.Equal(t, "valid-skill", skills[0].Name)

	// Should have a parse warning for the invalid directory
	require.Len(t, warnings, 1)
	assert.Equal(t, "invalid-dir", warnings[0].Name)
	assert.NotEmpty(t, warnings[0].Error)
}
