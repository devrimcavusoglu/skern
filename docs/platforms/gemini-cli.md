# Gemini CLI

[Gemini CLI](https://github.com/google-gemini/gemini-cli) is Google's command-line AI agent. Skern installs skills into its skills directory.

## Skill Paths

| Scope | Path |
|-------|------|
| User-level | `~/.gemini/skills/<name>/SKILL.md` |
| Project-level | `.agents/skills/<name>/SKILL.md` |

The project-level path is shared with `codex-cli`, `cursor`, and `github-copilot`. See [Platform Adapters › Shared project directory](/concepts/platform-adapters#shared-project-directory).

## Install a Skill

```sh
skern skill install code-review --platform gemini-cli
skern skill install code-review --platform gemini-cli --scope project
```

## Uninstall a Skill

```sh
skern skill uninstall code-review --platform gemini-cli
```

## Detection

Skern detects Gemini CLI by the presence of `~/.gemini/`.
