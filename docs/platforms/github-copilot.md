# GitHub Copilot

[GitHub Copilot](https://github.com/features/copilot) is GitHub's AI coding assistant. Skern installs skills into its skills directory.

## Skill Paths

| Scope | Path |
|-------|------|
| User-level | `~/.copilot/skills/<name>/SKILL.md` |
| Project-level | `.agents/skills/<name>/SKILL.md` |

The project-level path is shared with `codex-cli`, `cursor`, and `gemini-cli`. See [Platform Adapters › Shared project directory](/concepts/platform-adapters#shared-project-directory).

### Why `.agents/skills/` and not `.github/skills/`?

Copilot discovers project skills from **any** of `.github/skills/`, `.claude/skills/`, and `.agents/skills/`, and personal skills from `~/.copilot/skills/` or `~/.agents/skills/` ([GitHub docs: About agent skills](https://docs.github.com/en/copilot/concepts/agents/about-agent-skills)). All three project locations are valid; skern uses `.agents/skills/` because it is the cross-agent convention shared with Codex CLI, Cursor, and Gemini CLI, so a project-scoped install reaches every agent that reads it. If you need a Copilot-only location, install manually to `.github/skills/` — a per-platform destination override is tracked in [#101](https://github.com/devrimcavusoglu/skern/issues/101).

## Install a Skill

```sh
skern skill install code-review --platform github-copilot
skern skill install code-review --platform github-copilot --scope project
```

## Uninstall a Skill

```sh
skern skill uninstall code-review --platform github-copilot
```

## Detection

Skern detects GitHub Copilot by the presence of `~/.copilot/`.
