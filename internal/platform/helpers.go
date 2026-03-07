package platform

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const manifestFile = "SKILL.md"

// installSkill copies a skill directory into the target base directory.
// It returns an error if the skill already exists at the destination.
func installSkill(sourceDir, skillName, targetBaseDir string) error {
	destDir := filepath.Join(targetBaseDir, skillName)

	if _, err := os.Stat(destDir); err == nil {
		return fmt.Errorf("skill %q already installed at %s", skillName, destDir)
	}

	if err := os.MkdirAll(targetBaseDir, 0o755); err != nil {
		return fmt.Errorf("creating skills directory: %w", err)
	}

	if err := copyDir(sourceDir, destDir); err != nil {
		// Clean up on failure
		_ = os.RemoveAll(destDir)
		return fmt.Errorf("copying skill: %w", err)
	}

	return nil
}

// uninstallSkill removes a skill directory from the target base directory.
func uninstallSkill(skillName, targetBaseDir string) error {
	destDir := filepath.Join(targetBaseDir, skillName)

	if _, err := os.Stat(destDir); os.IsNotExist(err) {
		return fmt.Errorf("skill %q not installed at %s", skillName, targetBaseDir)
	}

	return os.RemoveAll(destDir)
}

// listInstalledSkills reads directory entries and returns names of valid skill directories.
func listInstalledSkills(targetBaseDir string) ([]string, error) {
	entries, err := os.ReadDir(targetBaseDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading skills directory: %w", err)
	}

	var names []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		// Validate that SKILL.md exists
		manifestPath := filepath.Join(targetBaseDir, entry.Name(), manifestFile)
		if _, err := os.Stat(manifestPath); err == nil {
			names = append(names, entry.Name())
		}
	}

	return names, nil
}

// copyDir recursively copies a directory tree from src to dst.
func copyDir(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)

		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}

		return copyFile(path, target)
	})
}

// copyFile copies a single file, preserving permissions.
func copyFile(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}

	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode())
	if err != nil {
		return err
	}
	defer func() { _ = out.Close() }()

	if _, err = io.Copy(out, in); err != nil {
		return err
	}

	return out.Sync()
}
