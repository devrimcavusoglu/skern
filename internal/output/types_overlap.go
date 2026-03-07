package output

// OverlapResult represents a single overlap match during skill creation.
type OverlapResult struct {
	Name  string  `json:"name"`
	Score float64 `json:"score"`
	Scope string  `json:"scope"`
}

// OverlapCheckResult is the JSON envelope for overlap detection output.
type OverlapCheckResult struct {
	Blocked  bool            `json:"blocked"`
	Matches  []OverlapResult `json:"matches"`
	MaxScore float64         `json:"max_score"`
}

// RecommendAction is the recommended action for the agent.
type RecommendAction string

const (
	// RecommendReuse indicates an existing skill closely matches and should be reused.
	RecommendReuse RecommendAction = "reuse"
	// RecommendExtend indicates a partial match exists that could be extended.
	RecommendExtend RecommendAction = "extend"
	// RecommendCreate indicates no close match exists and a new skill should be created.
	RecommendCreate RecommendAction = "create"
)

// ScoredSkillResult is a skill with its relevance score.
type ScoredSkillResult struct {
	SkillResult
	Score float64 `json:"score"`
}

// SkillRecommendResult is the JSON envelope for skill recommend output.
type SkillRecommendResult struct {
	Query         string              `json:"query"`
	Action        RecommendAction     `json:"action"`
	Reason        string              `json:"reason"`
	SuggestedName string              `json:"suggested_name,omitempty"`
	Matches       []ScoredSkillResult `json:"matches"`
	Count         int                 `json:"count"`
}
