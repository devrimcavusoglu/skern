# Supported Platforms

Skern ships eight adapters out of the box. Each adapter knows where its platform stores user-level and project-level skills, copies skills into place, and reports whether the platform is installed on the current host.

The set is driven by a declarative registry in [`internal/platform/spec.go`](https://github.com/devrimcavusoglu/skern/blob/main/internal/platform/spec.go) — adding a new platform is a one-line append (see [Platform Adapters › Adding a Platform](/concepts/platform-adapters#adding-a-platform)).

## Path Reference

| Adapter name | User-level skills | Project-level skills |
|--------------|-------------------|----------------------|
| [`claude-code`](/platforms/claude-code) | `~/.claude/skills/` | `.claude/skills/` |
| [`codex-cli`](/platforms/codex-cli) | `~/.agents/skills/` | `.agents/skills/` |
| [`opencode`](/platforms/opencode) | `~/.config/opencode/skills/` | `.opencode/skills/` |
| [`cursor`](/platforms/cursor) | `~/.cursor/skills/` | `.agents/skills/` |
| [`gemini-cli`](/platforms/gemini-cli) | `~/.gemini/skills/` | `.agents/skills/` |
| [`github-copilot`](/platforms/github-copilot) | `~/.copilot/skills/` | `.agents/skills/` |
| [`windsurf`](/platforms/windsurf) | `~/.codeium/windsurf/skills/` | `.windsurf/skills/` |
| [`continue`](/platforms/continue) | `~/.continue/skills/` | `.continue/skills/` |

Path conventions track [vercel-labs/skills](https://github.com/vercel-labs/skills#supported-agents), the closest thing to a community standard for Agent Skills directory layout.

`codex-cli`, `cursor`, `gemini-cli`, and `github-copilot` all use `.agents/skills/` as their project-level directory. A skill installed there is visible to every agent that reads from it — that's intentional and matches the upstream convention. See [Platform Adapters › Shared project directory](/concepts/platform-adapters#shared-project-directory) for the implications on capacity reporting and detection.

## Auto-Detection

`skern platform list` reports which adapters appear installed on the current host. Detection is per-platform: each adapter checks its own user-level config dir (`~/.claude`, `~/.cursor`, `~/.gemini`, `~/.copilot`, `~/.codex`, …) so platforms that share a project directory are still distinguished.

## Feature Comparison

| Feature | skern | Manual folder skills | AI tool built-in |
|---------|-------|----------------------|------------------|
| System-wide registry | Yes | No | No |
| Cross-platform install | Yes | No | No |
| Overlap detection | Yes | No | No |
| Per-platform capacity reporting | Yes | No | No |
| Validation (errors / warnings / hints) | Yes | No | No |
| Versioning (semver, bump, diff) | Yes | No | No |
| Skill import (URL / GitHub / gist) | Yes | No | No |
| CLI-first | Yes | Partial | No |
| Agent-agnostic | Yes | Partial | No |

## Quick Links

- [Claude Code](/platforms/claude-code) — Anthropic's terminal AI assistant
- [Codex CLI](/platforms/codex-cli) — OpenAI's terminal coding agent
- [OpenCode](/platforms/opencode) — open-source AI coding tool
- [Cursor](/platforms/cursor) — AI-first code editor
- [Gemini CLI](/platforms/gemini-cli) — Google's command-line agent
- [GitHub Copilot](/platforms/github-copilot) — GitHub's AI coding assistant
- [Windsurf](/platforms/windsurf) — Codeium's agentic IDE
- [Continue](/platforms/continue) — open-source AI code assistant
