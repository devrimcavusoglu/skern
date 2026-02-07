// Package registry provides filesystem-based CRUD operations for skill storage.
package registry

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/devrimcavusoglu/scribe/internal/skill"
)

const manifestFile = "SKILL.md"

// Registry manages skills stored on the filesystem.
type Registry struct {
	userDir    string
	projectDir string
}

// New creates a Registry with the given user and project directories.
func New(userDir, projectDir string) *Registry {
	return &Registry{
		userDir:    userDir,
		projectDir: projectDir,
	}
}

// Create writes a new skill to the given scope directory.
// Returns the path where the skill was created.
func (r *Registry) Create(s *skill.Skill, scope skill.Scope) (string, error) {
	if err := skill.ValidateName(s.Name); err != nil {
		return "", err
	}

	dir := r.scopeDir(scope)
	skillDir := filepath.Join(dir, s.Name)

	if _, err := os.Stat(skillDir); err == nil {
		return "", fmt.Errorf("skill %q already exists in %s scope", s.Name, scope)
	}

	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		return "", fmt.Errorf("creating skill directory: %w", err)
	}

	manifestPath := filepath.Join(skillDir, manifestFile)
	if err := skill.WriteManifest(s, manifestPath); err != nil {
		// Clean up on failure
		_ = os.RemoveAll(skillDir)
		return "", fmt.Errorf("writing manifest: %w", err)
	}

	return skillDir, nil
}

// Get reads a skill from the given scope by name.
// Returns the skill, its directory path, and any error.
func (r *Registry) Get(name string, scope skill.Scope) (*skill.Skill, string, error) {
	if err := skill.ValidateName(name); err != nil {
		return nil, "", err
	}

	dir := r.scopeDir(scope)
	skillDir := filepath.Join(dir, name)
	manifestPath := filepath.Join(skillDir, manifestFile)

	s, err := skill.ParseManifest(manifestPath)
	if err != nil {
		return nil, "", fmt.Errorf("skill %q not found in %s scope: %w", name, scope, err)
	}

	return s, skillDir, nil
}

// Remove deletes a skill directory from the given scope.
func (r *Registry) Remove(name string, scope skill.Scope) error {
	if err := skill.ValidateName(name); err != nil {
		return err
	}

	dir := r.scopeDir(scope)
	skillDir := filepath.Join(dir, name)

	if _, err := os.Stat(skillDir); os.IsNotExist(err) {
		return fmt.Errorf("skill %q not found in %s scope", name, scope)
	}

	return os.RemoveAll(skillDir)
}

// List returns all skills in the given scope.
func (r *Registry) List(scope skill.Scope) ([]skill.Skill, error) {
	dir := r.scopeDir(scope)

	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading %s skills directory: %w", scope, err)
	}

	var skills []skill.Skill
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		manifestPath := filepath.Join(dir, entry.Name(), manifestFile)
		s, err := skill.ParseManifest(manifestPath)
		if err != nil {
			continue // skip invalid entries
		}

		skills = append(skills, *s)
	}

	return skills, nil
}

// Exists checks whether a skill exists in the given scope.
func (r *Registry) Exists(name string, scope skill.Scope) bool {
	dir := r.scopeDir(scope)
	skillDir := filepath.Join(dir, name)
	_, err := os.Stat(filepath.Join(skillDir, manifestFile))
	return err == nil
}

func (r *Registry) scopeDir(scope skill.Scope) string {
	if scope == skill.ScopeProject {
		return r.projectDir
	}
	return r.userDir
}
