# OpenCode

[OpenCode](https://github.com/opencode-ai/opencode) is an open-source AI coding tool. It reads skills from the `.opencode/skills/` directory.

## Skill Paths

| Scope | Path |
|-------|------|
| User-level | `~/.config/opencode/skills/<name>/SKILL.md` |
| Project-level | `.opencode/skills/<name>/SKILL.md` |

OpenCode uses dedicated user and project directories — it does not share `.agents/skills/`.

## Install a Skill

```sh
# User-level (default scope is user)
skern skill install code-review --platform opencode

# Project-level
skern skill install code-review --platform opencode --scope project
```

## Uninstall a Skill

```sh
skern skill uninstall code-review --platform opencode
```

## Detection

Skern detects OpenCode by the presence of `~/.config/opencode/`.

## How Skills Work in OpenCode

OpenCode reads `SKILL.md` files from its skills directories. The skill's instructions become part of the agent's context, providing reusable capabilities across coding sessions.
