# OpenCode

[OpenCode](https://github.com/opencode-ai/opencode) is an open-source AI coding tool. It reads skills from the `.opencode/skills/` directory.

## Skill Paths

| Scope | Path |
|-------|------|
| User-level | `~/.config/opencode/skills/<name>/SKILL.md` |
| Project-level | `.opencode/skills/<name>/SKILL.md` |

## Install a Skill

```sh
# Project-level (default)
skern skill install code-review --platform opencode

# User-level
skern skill install code-review --platform opencode --scope user
```

## Uninstall a Skill

```sh
skern skill uninstall code-review --platform opencode
```

## Detection

Skern detects OpenCode by checking for the presence of `.opencode/` in the current project or `~/.config/opencode/` at the user level.

## How Skills Work in OpenCode

OpenCode reads `SKILL.md` files from its skills directories. The skill's instructions become part of the agent's context, providing reusable capabilities across coding sessions.
