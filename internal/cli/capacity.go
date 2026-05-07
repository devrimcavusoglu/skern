package cli

import (
	"fmt"
	"strings"

	"github.com/devrimcavusoglu/skern/internal/output"
	"github.com/devrimcavusoglu/skern/internal/platform"
	"github.com/devrimcavusoglu/skern/internal/skill"
)

// platformNamesList returns a comma-separated list of supported platform names
// for use in flag help text and error messages. Pulled from the platform spec
// registry so adding a platform takes one line.
func platformNamesList() string {
	names := platform.SupportedNames()
	parts := make([]string, len(names))
	for i, n := range names {
		parts[i] = string(n)
	}
	return strings.Join(parts, ", ")
}

// buildCapacityReport queries a platform for its currently-installed skills
// at the given scope and returns a CapacityReport ready for inclusion in
// command output. Returns nil if the query fails (capacity is best-effort
// telemetry, not an operation prerequisite).
func buildCapacityReport(p platform.Platform, scope skill.Scope) *output.CapacityReport {
	names, err := p.InstalledSkills(scope)
	if err != nil {
		return nil
	}
	threshold := skill.PlatformThreshold(scope)
	installed := len(names)
	return &output.CapacityReport{
		Platform:   string(p.Name()),
		Scope:      string(scope),
		Installed:  installed,
		Threshold:  threshold,
		Headroom:   threshold - installed,
		OverBudget: installed >= threshold,
	}
}

// formatCapacityWarning produces a human-readable line describing capacity
// usage after an operation. Returns "" when there is nothing actionable to
// report (well under threshold).
func formatCapacityWarning(r *output.CapacityReport) string {
	if r == nil {
		return ""
	}
	if r.OverBudget {
		return fmt.Sprintf("Capacity: %s (%s) at %d/%d skills — over threshold.\n",
			r.Platform, r.Scope, r.Installed, r.Threshold)
	}
	if r.Headroom <= 5 {
		return fmt.Sprintf("Capacity: %s (%s) at %d/%d skills — %d slot(s) remaining.\n",
			r.Platform, r.Scope, r.Installed, r.Threshold, r.Headroom)
	}
	return ""
}
