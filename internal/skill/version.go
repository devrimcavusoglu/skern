package skill

import (
	"fmt"
	"strconv"
	"strings"
)

// Version represents a semantic version (MAJOR.MINOR.PATCH).
type Version struct {
	Major int
	Minor int
	Patch int
}

// ParseVersion parses a semver string into a Version.
func ParseVersion(s string) (Version, error) {
	parts := strings.Split(s, ".")
	if len(parts) != 3 {
		return Version{}, fmt.Errorf("invalid version %q: must be MAJOR.MINOR.PATCH", s)
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil || major < 0 {
		return Version{}, fmt.Errorf("invalid version %q: major must be a non-negative integer", s)
	}

	minor, err := strconv.Atoi(parts[1])
	if err != nil || minor < 0 {
		return Version{}, fmt.Errorf("invalid version %q: minor must be a non-negative integer", s)
	}

	patch, err := strconv.Atoi(parts[2])
	if err != nil || patch < 0 {
		return Version{}, fmt.Errorf("invalid version %q: patch must be a non-negative integer", s)
	}

	return Version{Major: major, Minor: minor, Patch: patch}, nil
}

// String returns the semver string representation.
func (v Version) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

// BumpPatch returns a new Version with the patch number incremented.
func (v Version) BumpPatch() Version {
	return Version{Major: v.Major, Minor: v.Minor, Patch: v.Patch + 1}
}

// BumpMinor returns a new Version with the minor number incremented and patch reset to 0.
func (v Version) BumpMinor() Version {
	return Version{Major: v.Major, Minor: v.Minor + 1, Patch: 0}
}

// BumpMajor returns a new Version with the major number incremented and minor/patch reset to 0.
func (v Version) BumpMajor() Version {
	return Version{Major: v.Major + 1, Minor: 0, Patch: 0}
}

// CompareVersions compares two version strings and returns the upgrade kind.
// Returns ("", nil) if versions are equal.
// Returns ("patch"|"minor"|"major", nil) indicating the kind of upgrade from a to b.
// Returns a negative kind if b < a (downgrade): the kind still reflects the most significant difference.
func CompareVersions(a, b string) (kind string, newer bool, err error) {
	va, err := ParseVersion(a)
	if err != nil {
		return "", false, fmt.Errorf("parsing version a: %w", err)
	}

	vb, err := ParseVersion(b)
	if err != nil {
		return "", false, fmt.Errorf("parsing version b: %w", err)
	}

	if va.Major != vb.Major {
		return "major", vb.Major > va.Major, nil
	}
	if va.Minor != vb.Minor {
		return "minor", vb.Minor > va.Minor, nil
	}
	if va.Patch != vb.Patch {
		return "patch", vb.Patch > va.Patch, nil
	}

	return "", false, nil
}

// BumpVersion parses a version string, applies the given bump level, and returns the new version string.
func BumpVersion(version, level string) (string, error) {
	v, err := ParseVersion(version)
	if err != nil {
		return "", err
	}

	switch level {
	case "patch":
		return v.BumpPatch().String(), nil
	case "minor":
		return v.BumpMinor().String(), nil
	case "major":
		return v.BumpMajor().String(), nil
	default:
		return "", fmt.Errorf("invalid bump level %q: must be patch, minor, or major", level)
	}
}
