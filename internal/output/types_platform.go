package output

// PlatformActionEntry records the result of installing or uninstalling a skill on one platform.
type PlatformActionEntry struct {
	Platform string `json:"platform"`
	Success  bool   `json:"success"`
	Error    string `json:"error,omitempty"`
}

// SkillInstallResult is the JSON envelope for skill install output.
type SkillInstallResult struct {
	Skill     string                `json:"skill"`
	Scope     string                `json:"scope"`
	Platforms []PlatformActionEntry `json:"platforms"`
}

// SkillUninstallResult is the JSON envelope for skill uninstall output.
type SkillUninstallResult struct {
	Skill     string                `json:"skill"`
	Scope     string                `json:"scope"`
	Platforms []PlatformActionEntry `json:"platforms"`
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
