package output

// InitResult is the JSON envelope for init output.
type InitResult struct {
	Path    string `json:"path"`
	Created bool   `json:"created"`
}

// VersionResult is the JSON envelope for version output.
type VersionResult struct {
	Version string `json:"version"`
	Commit  string `json:"commit,omitempty"`
	Date    string `json:"date,omitempty"`
}

// AuthorResult is the JSON representation of a skill author.
type AuthorResult struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Platform string `json:"platform,omitempty"`
}

// ModifiedByResult is the JSON representation of a modified-by entry.
type ModifiedByResult struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Platform string `json:"platform,omitempty"`
	Date     string `json:"date"`
}

// SkillResult is the JSON representation of a skill.
type SkillResult struct {
	Name         string             `json:"name"`
	Description  string             `json:"description"`
	Version      string             `json:"version"`
	Author       AuthorResult       `json:"author"`
	Tags         []string           `json:"tags,omitempty"`
	Scope        string             `json:"scope,omitempty"`
	Path         string             `json:"path,omitempty"`
	AllowedTools []string           `json:"allowed_tools,omitempty"`
	Files        []string           `json:"files,omitempty"`
	ModifiedBy   []ModifiedByResult `json:"modified_by,omitempty"`
}

// SkillCreateResult is the JSON envelope for skill create output.
type SkillCreateResult struct {
	Name  string `json:"name"`
	Scope string `json:"scope"`
	Path  string `json:"path"`
}

// DuplicateHint represents a pair of skills flagged as potential duplicates.
type DuplicateHint struct {
	SkillA string  `json:"skill_a"`
	SkillB string  `json:"skill_b"`
	Score  float64 `json:"score"`
}

// ParseWarningResult records a skill directory that could not be parsed.
type ParseWarningResult struct {
	Name  string `json:"name"`
	Error string `json:"error"`
}

// SkillListResult is the JSON envelope for skill list output.
type SkillListResult struct {
	Skills        []SkillResult        `json:"skills"`
	Count         int                  `json:"count"`
	Duplicates    []DuplicateHint      `json:"duplicates,omitempty"`
	ParseWarnings []ParseWarningResult `json:"parse_warnings,omitempty"`
}

// SkillSearchResult is the JSON envelope for skill search output.
type SkillSearchResult struct {
	Query   string        `json:"query"`
	Results []SkillResult `json:"results"`
	Count   int           `json:"count"`
}

// SkillEditResult is the JSON envelope for skill edit output.
type SkillEditResult struct {
	Name    string   `json:"name"`
	Scope   string   `json:"scope"`
	Updated []string `json:"updated"`
}

// SkillRemoveResult is the JSON envelope for skill remove output.
type SkillRemoveResult struct {
	Name  string `json:"name"`
	Scope string `json:"scope"`
}

// ValidationIssueResult is the JSON representation of a single validation issue.
type ValidationIssueResult struct {
	Field    string `json:"field"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
}

// SkillValidateResult is the JSON envelope for skill validate output.
type SkillValidateResult struct {
	Name   string                  `json:"name"`
	Valid  bool                    `json:"valid"`
	Issues []ValidationIssueResult `json:"issues"`
	Errors int                     `json:"errors"`
	Warns  int                     `json:"warnings"`
	Hints  int                     `json:"hints"`
}

// SkillVersionResult is the JSON envelope for skill version output.
type SkillVersionResult struct {
	Name            string `json:"name"`
	Version         string `json:"version"`
	Scope           string `json:"scope"`
	PreviousVersion string `json:"previous_version,omitempty"`
	Bumped          bool   `json:"bumped"`
}

// VersionCompareResult is the JSON envelope for version comparison output.
type VersionCompareResult struct {
	Installed string `json:"installed"`
	Available string `json:"available"`
	Kind      string `json:"kind,omitempty"`
	Upgrade   bool   `json:"upgrade"`
}
