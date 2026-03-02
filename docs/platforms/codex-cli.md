# Codex CLI

[Codex CLI](https://github.com/openai/codex) is OpenAI's terminal-based coding agent. It reads skills from the `.agents/skills/` directory.

## Skill Paths

| Scope | Path |
|-------|------|
| User-level | `~/.agents/skills/<name>/SKILL.md` |
| Project-level | `.agents/skills/<name>/SKILL.md` |

## Install a Skill

```sh
# Project-level (default)
skern skill install code-review --platform codex-cli

# User-level
skern skill install code-review --platform codex-cli --scope user
```

## Uninstall a Skill

```sh
skern skill uninstall code-review --platform codex-cli
```

## Detection

Skern detects Codex CLI by checking for the presence of `.agents/` in the current project or `~/.agents/` at the user level.

## How Skills Work in Codex CLI

Codex CLI reads `SKILL.md` files from its skills directories at startup. The skill's instructions are incorporated into the agent's context, making the capabilities available during the coding session.
