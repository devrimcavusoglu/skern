# Supported Platforms

Skern ships adapters for the most common agentic development platforms. Each adapter knows where its platform stores user-level and project-level skills, copies skills into place, and reports which platforms are installed on the host.

The set is driven by a declarative registry in `internal/platform/spec.go` — adding a new platform is a one-line append (see [Platform Adapters](/concepts/platform-adapters#adding-a-platform)).

## Path Reference

| Adapter name | User-level skills | Project-level skills |
|--------------|--------------------|----------------------|
| `claude-code` | `~/.claude/skills/` | `.claude/skills/` |
| `codex-cli` | `~/.agents/skills/` | `.agents/skills/` |
| `opencode` | `~/.config/opencode/skills/` | `.opencode/skills/` |
| `cursor` | `~/.cursor/skills/` | `.agents/skills/` |
| `gemini-cli` | `~/.gemini/skills/` | `.agents/skills/` |
| `github-copilot` | `~/.copilot/skills/` | `.agents/skills/` |
| `windsurf` | `~/.codeium/windsurf/skills/` | `.windsurf/skills/` |
| `continue` | `~/.continue/skills/` | `.continue/skills/` |

Several platforms share `.agents/skills/` as their project directory — a skill installed there is visible to every agent that reads from it. See [Platform Adapters](/concepts/platform-adapters#shared-project-directory) for the implications on capacity reporting and detection.

## Auto-detection

`skern platform list` reports which adapters appear installed on the current host. Detection is per-platform: each adapter checks its own user-level config dir (e.g. `~/.cursor`, `~/.gemini`, `~/.copilot`) so platforms that share a project directory are still distinguished.

## Feature Comparison

| Feature | skern | Manual Folder Skills | AI Tool Built-in |
|---------|-------|----------------------|------------------|
| System-wide registry | Yes | No | No |
| Cross-platform install | Yes | No | No |
| Overlap detection | Yes | No | No |
| CLI-first | Yes | Partial | No |
| Agent-agnostic | Yes | Partial | No |
| Validation | Yes | No | No |
| Versioning | Planned | No | No |

## Quick Links

- [Claude Code](/platforms/claude-code) — Anthropic's AI coding assistant
- [Codex CLI](/platforms/codex-cli) — OpenAI's terminal-based coding agent
- [OpenCode](/platforms/opencode) — Open-source AI coding tool
