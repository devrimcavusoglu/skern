package skill

// DefaultBody returns the default body content for a new skill.
func DefaultBody() string {
	return "## Instructions\n\nTODO: Add step-by-step instructions for the agent.\n"
}

// NewSkill creates a new Skill with sensible defaults.
func NewSkill(name, description, authorName, authorType, authorPlatform string) *Skill {
	return NewSkillWithBody(name, description, authorName, authorType, authorPlatform, "")
}

// NewSkillWithBody creates a new Skill with a custom body. If body is empty, DefaultBody() is used.
func NewSkillWithBody(name, description, authorName, authorType, authorPlatform, body string) *Skill {
	if description == "" {
		description = "TODO: Describe what this skill does and when to use it.\n"
	}

	if body == "" {
		body = DefaultBody()
	}

	author := Author{
		Name:     authorName,
		Type:     authorType,
		Platform: authorPlatform,
	}

	return &Skill{
		Name:        name,
		Description: description,
		Metadata: Metadata{
			Author:  author,
			Version: "0.1.0",
		},
		Body: body,
	}
}
