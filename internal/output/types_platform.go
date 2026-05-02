package output

// SkillActionEntry records the result of installing or uninstalling one skill
// during a batch operation against a single platform.
type SkillActionEntry struct {
	Skill   string `json:"skill"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// CapacityReport summarizes how many skills are installed on a platform after
// an install/uninstall operation, alongside the configured threshold. Agents
// can use this to make eviction decisions before installing more skills.
//
// Per #52 (D4): tracked per (platform, scope) pair; project and user scopes
// are reported as separate snapshots when both are relevant.
type CapacityReport struct {
	Platform   string `json:"platform"`
	Scope      string `json:"scope"`
	Installed  int    `json:"installed"`
	Threshold  int    `json:"threshold"`
	Headroom   int    `json:"headroom"`
	OverBudget bool   `json:"over_budget"`
}

// SkillInstallResult is the JSON envelope for skill install output.
//
// Each invocation targets exactly one platform (per #52 D6); the Skills slice
// contains one entry per skill name passed on the command line.
type SkillInstallResult struct {
	Platform string             `json:"platform"`
	Scope    string             `json:"scope"`
	Skills   []SkillActionEntry `json:"skills"`
	Capacity *CapacityReport    `json:"capacity,omitempty"`
}

// SkillUninstallResult is the JSON envelope for skill uninstall output.
type SkillUninstallResult struct {
	Platform string             `json:"platform"`
	Scope    string             `json:"scope"`
	Skills   []SkillActionEntry `json:"skills"`
	Capacity *CapacityReport    `json:"capacity,omitempty"`
}

// PlatformResult represents a single detected platform.
type PlatformResult struct {
	Name        string `json:"name"`
	Detected    bool   `json:"detected"`
	UserPath    string `json:"user_path"`
	ProjectPath string `json:"project_path"`
}

// PlatformListResult is the JSON envelope for platform list output.
type PlatformListResult struct {
	Platforms []PlatformResult `json:"platforms"`
	Count     int              `json:"count"`
}

// PlatformInstallStatus shows whether a skill is installed on a platform.
type PlatformInstallStatus struct {
	Platform  string `json:"platform"`
	Installed bool   `json:"installed"`
}

// PlatformStatusEntry shows install status for one skill across platforms.
type PlatformStatusEntry struct {
	Skill     string                  `json:"skill"`
	Platforms []PlatformInstallStatus `json:"platforms"`
}

// PlatformStatusResult is the JSON envelope for platform status output.
type PlatformStatusResult struct {
	Scope  string                `json:"scope"`
	Status []PlatformStatusEntry `json:"status"`
}
