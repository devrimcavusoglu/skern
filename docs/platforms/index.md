# Supported Platforms

Skern supports three agentic development platforms. Each platform has a dedicated adapter that handles skill installation and uninstallation.

## Platform Comparison

| Feature | Claude Code | Codex CLI | OpenCode |
|---------|-------------|-----------|----------|
| User-level skills | `~/.claude/skills/` | `~/.agents/skills/` | `~/.config/opencode/skills/` |
| Project-level skills | `.claude/skills/` | `.agents/skills/` | `.opencode/skills/` |
| Auto-detection | Yes | Yes | Yes |
| Install with `--platform all` | Yes | Yes | Yes |

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
