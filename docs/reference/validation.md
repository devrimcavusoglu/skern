# Validation

`skern skill validate <name>` checks skills against the [Agent Skills](https://agentskills.io) specification. Validation also runs automatically during `skern skill create`, issuing warnings for any issues.

## Rules

### Name Format

Skill names must match the pattern `[a-z0-9]+(-[a-z0-9]+)*` and be between 1 and 64 characters.

Valid examples: `code-review`, `lint-fix`, `deploy`

Invalid examples: `Code_Review`, `my skill`, `a-very-long-skill-name-that-exceeds-the-sixty-four-character-maximum-limit`

### Description

- Required — must not be empty
- Maximum 1024 characters

### Body

The `SKILL.md` file must have non-empty body content after the YAML frontmatter.

### Allowed Tools

If `allowed-tools` is specified in the frontmatter, no entries may be empty strings.

### Metadata

- **Author type** — must be `human` or `agent`
- **Version** — should follow [semantic versioning](https://semver.org) (e.g., `1.0.0`)

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Skill is valid |
| 2 | Validation failure |
