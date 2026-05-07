# Cursor

[Cursor](https://www.cursor.com/) is an AI-first code editor. Skern installs skills into its skills directory.

## Skill Paths

| Scope | Path |
|-------|------|
| User-level | `~/.cursor/skills/<name>/SKILL.md` |
| Project-level | `.agents/skills/<name>/SKILL.md` |

The project-level path is shared with `codex-cli`, `gemini-cli`, and `github-copilot`. A skill installed at `.agents/skills/` is visible to every agent that reads from there — see [Platform Adapters › Shared project directory](/concepts/platform-adapters#shared-project-directory).

## Install a Skill

```sh
# User-level (default scope is user)
skern skill install code-review --platform cursor

# Project-level
skern skill install code-review --platform cursor --scope project
```

## Uninstall a Skill

```sh
skern skill uninstall code-review --platform cursor
```

## Detection

Skern detects Cursor by the presence of `~/.cursor/`. The shared `.agents/skills/` project directory does not by itself signal Cursor — detection keys on the user-level config dir.
