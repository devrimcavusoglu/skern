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

// NewDetector creates a Detector initialized with all known platform adapters using real paths.
func NewDetector() (*Detector, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("determining home directory: %w", err)
	}

	platforms := []Platform{
		NewClaudeCode(home, "."),
		NewCodexCLI(home, "."),
		NewOpenCode(home, "."),
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
// It accepts "all" as a special value.
func ParsePlatformType(s string) (Type, error) {
	normalized := strings.ToLower(strings.TrimSpace(s))
	switch Type(normalized) {
	case TypeClaudeCode, TypeCodexCLI, TypeOpenCode, TypeAll:
		return Type(normalized), nil
	}
	return "", fmt.Errorf("unknown platform %q: must be one of claude-code, codex-cli, opencode, or all", s)
}
