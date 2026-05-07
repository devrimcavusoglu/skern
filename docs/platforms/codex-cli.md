# Codex CLI

[Codex CLI](https://github.com/openai/codex) is OpenAI's terminal-based coding agent. It reads skills from the `.agents/skills/` directory.

## Skill Paths

| Scope | Path |
|-------|------|
| User-level | `~/.agents/skills/<name>/SKILL.md` |
| Project-level | `.agents/skills/<name>/SKILL.md` |

The project-level path is shared with `cursor`, `gemini-cli`, and `github-copilot`. A skill installed there is visible to every agent that reads from `.agents/skills/` — see [Platform Adapters › Shared project directory](/concepts/platform-adapters#shared-project-directory).

::: tip
The user-level path stays at `~/.agents/skills/` for compatibility with skern's original layout. Detection covers both `~/.codex/` and `~/.agents/`.
:::

## Install a Skill

```sh
# User-level (default scope is user)
skern skill install code-review --platform codex-cli

# Project-level
skern skill install code-review --platform codex-cli --scope project
```

## Uninstall a Skill

```sh
skern skill uninstall code-review --platform codex-cli
```

## Detection

Skern detects Codex CLI by the presence of `~/.codex/` (preferred) or `~/.agents/`.

## How Skills Work in Codex CLI

Codex CLI reads `SKILL.md` files from its skills directories at startup. The skill's instructions are incorporated into the agent's context, making the capabilities available during the coding session.
