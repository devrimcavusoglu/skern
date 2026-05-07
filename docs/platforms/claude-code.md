# Claude Code

[Claude Code](https://docs.anthropic.com/en/docs/claude-code) is Anthropic's terminal-based AI coding assistant. It reads skills from the `.claude/skills/` directory.

## Skill Paths

| Scope | Path |
|-------|------|
| User-level | `~/.claude/skills/<name>/SKILL.md` |
| Project-level | `.claude/skills/<name>/SKILL.md` |

Claude Code uses dedicated user and project directories — it does not share `.agents/skills/`.

## Install a Skill

```sh
# User-level (default scope is user)
skern skill install code-review --platform claude-code

# Project-level
skern skill install code-review --platform claude-code --scope project
```

## Uninstall a Skill

```sh
skern skill uninstall code-review --platform claude-code
```

## Detection

Skern detects Claude Code by the presence of `~/.claude/`.

## How Skills Work in Claude Code

When Claude Code starts a session, it reads `SKILL.md` files from the skills directory. Skills become available as capabilities Claude can use during the session — the skill's markdown body serves as instructions Claude follows when the skill is activated.
