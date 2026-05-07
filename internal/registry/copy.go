package registry

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/devrimcavusoglu/skern/internal/skill"
)

// CopySiblings recursively copies every entry from srcDir into dstDir,
// skipping the top-level SKILL.md (the manifest is written separately by
// the caller). Used by `skill create --from-template <dir>` to populate
// companion files (references/, templates/, VENDORED.md, ...) alongside
// the freshly written manifest.
//
// dstDir must already exist. Symlinks are not followed; a symlink at the
// source becomes a regular file copy of its target's content (via os.Open
// which follows the link). Non-regular non-directory entries (devices,
// sockets) are skipped.
func CopySiblings(srcDir, dstDir string) error {
	return filepath.WalkDir(srcDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}

		// Skip the top-level SKILL.md — already written by WriteManifest.
		if rel == skill.ManifestFile {
			return nil
		}

		dstPath := filepath.Join(dstDir, rel)

		if d.IsDir() {
			return os.MkdirAll(dstPath, 0o755)
		}

		info, err := d.Info()
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return nil
		}

		return copyFile(path, dstPath, info.Mode().Perm())
	})
}

func copyFile(src, dst string, perm os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("opening %s: %w", src, err)
	}
	defer func() { _ = in.Close() }()

	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return fmt.Errorf("creating parent dir for %s: %w", dst, err)
	}

	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return fmt.Errorf("creating %s: %w", dst, err)
	}
	defer func() { _ = out.Close() }()

	if _, err := io.Copy(out, in); err != nil {
		return fmt.Errorf("copying %s to %s: %w", src, dst, err)
	}
	return nil
}
