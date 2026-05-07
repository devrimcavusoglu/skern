# GitHub Copilot

[GitHub Copilot](https://github.com/features/copilot) is GitHub's AI coding assistant. Skern installs skills into its skills directory.

## Skill Paths

| Scope | Path |
|-------|------|
| User-level | `~/.copilot/skills/<name>/SKILL.md` |
| Project-level | `.agents/skills/<name>/SKILL.md` |

The project-level path is shared with `codex-cli`, `cursor`, and `gemini-cli`. See [Platform Adapters › Shared project directory](/concepts/platform-adapters#shared-project-directory).

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
