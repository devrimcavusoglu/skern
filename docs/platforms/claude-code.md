# Claude Code

[Claude Code](https://docs.anthropic.com/en/docs/claude-code) is Anthropic's AI coding assistant that works in the terminal. It reads skills from the `.claude/skills/` directory.

## Skill Paths

| Scope | Path |
|-------|------|
| User-level | `~/.claude/skills/<name>/SKILL.md` |
| Project-level | `.claude/skills/<name>/SKILL.md` |

## Install a Skill

```sh
# Project-level (default)
skern skill install code-review --platform claude-code

# User-level
skern skill install code-review --platform claude-code --scope user
```

## Uninstall a Skill

```sh
skern skill uninstall code-review --platform claude-code
```

## Detection

Skern detects Claude Code by checking for the presence of `.claude/` in the current project or `~/.claude/` at the user level.

## How Skills Work in Claude Code

When Claude Code starts a session, it reads all `SKILL.md` files from the skills directories. Skills become available as capabilities that Claude can use during the session. The skill's markdown body serves as instructions that Claude follows when the skill is activated.
