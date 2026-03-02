# Platform Adapters

Platform adapters bridge the skill registry with agent runtimes. Each adapter knows the target platform's directory structure and copies skill files accordingly.

## How Adapters Work

When you run `skern skill install`, the adapter:

1. Reads the `SKILL.md` from the skern registry
2. Creates the platform-specific skill directory
3. Copies the `SKILL.md` into the target location
4. The agent runtime discovers the skill on its next invocation

## Supported Platforms

| Platform | Adapter name | Detection |
|----------|-------------|-----------|
| Claude Code | `claude-code` | Looks for `.claude/` or `~/.claude/` |
| Codex CLI | `codex-cli` | Looks for `.agents/` or `~/.agents/` |
| OpenCode | `opencode` | Looks for `.opencode/` or `~/.config/opencode/` |

## Installation Paths

Each platform uses different directories for user-level and project-level skills:

| Platform | User-level | Project-level |
|----------|-----------|---------------|
| Claude Code | `~/.claude/skills/<name>/` | `.claude/skills/<name>/` |
| Codex CLI | `~/.agents/skills/<name>/` | `.agents/skills/<name>/` |
| OpenCode | `~/.config/opencode/skills/<name>/` | `.opencode/skills/<name>/` |

## Auto-Detection

Skern auto-detects which platforms are installed on your system. Use `skern platform list` to see detected platforms:

```sh
skern platform list
```

## Install to All Platforms

Use `--platform all` to install a skill to every detected platform at once:

```sh
skern skill install code-review --platform all
```

## Platform Status Matrix

View which skills are installed on which platforms:

```sh
skern platform status
```

This shows a matrix of skills and their installation status across all detected platforms.
