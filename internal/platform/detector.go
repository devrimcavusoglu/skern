package platform

import (
	"fmt"
	"os"
	"strings"
)

// Detector discovers installed platforms and provides access to adapters.
type Detector struct {
	platforms []Platform
}

// NewDetector creates a Detector initialized with one Adapter per registered
// Spec. Adding a platform never requires touching this function.
func NewDetector() (*Detector, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("determining home directory: %w", err)
	}

	platforms := make([]Platform, 0, len(Specs))
	for _, s := range Specs {
		platforms = append(platforms, New(s.Name, home, "."))
	}

	return &Detector{platforms: platforms}, nil
}

// NewDetectorWithPlatforms creates a Detector with the given platform adapters.
// Useful for testing with mock platforms.
func NewDetectorWithPlatforms(platforms []Platform) *Detector {
	return &Detector{platforms: platforms}
}

// DetectAll returns only the platforms that are detected as installed.
func (d *Detector) DetectAll() []Platform {
	var detected []Platform
	for _, p := range d.platforms {
		if p.Detect() {
			detected = append(detected, p)
		}
	}
	return detected
}

// Get returns the platform with the given type, or nil if not found.
func (d *Detector) Get(name Type) Platform {
	for _, p := range d.platforms {
		if p.Name() == name {
			return p
		}
	}
	return nil
}

// All returns all registered platforms, whether detected or not.
func (d *Detector) All() []Platform {
	return d.platforms
}

// ParsePlatformType validates and returns a platform type from a string flag value.
// Each invocation of skern targets exactly one platform; the agent that is
// running skern is expected to specify its own platform explicitly.
func ParsePlatformType(s string) (Type, error) {
	normalized := strings.ToLower(strings.TrimSpace(s))
	if normalized == "all" {
		return "", fmt.Errorf("platform %q is no longer supported; specify a single platform (one of: %s) — agents should target the platform they are running on", s, supportedNamesList())
	}
	if SpecFor(Type(normalized)) != nil {
		return Type(normalized), nil
	}
	return "", fmt.Errorf("unknown platform %q: must be one of %s", s, supportedNamesList())
}

// supportedNamesList returns a comma-separated list of registered platform
// names for use in user-facing error messages.
func supportedNamesList() string {
	names := SupportedNames()
	parts := make([]string, len(names))
	for i, n := range names {
		parts[i] = string(n)
	}
	return strings.Join(parts, ", ")
}
