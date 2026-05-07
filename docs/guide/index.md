# Getting Started

Skern is a minimal, agent-first CLI for managing [Agent Skills](https://agentskills.io) across agentic development platforms. It provides a standardized lifecycle — create, validate, install, remove — for skills that work natively with **Claude Code**, **Codex CLI**, **OpenCode**, **Cursor**, **Gemini CLI**, **GitHub Copilot**, **Windsurf**, and **Continue**.

Skills follow the Agent Skills open standard (`SKILL.md` with YAML frontmatter) and are immediately usable by any compatible platform without adapters or format conversion.

## Why skern?

Modern AI coding tools each ship their own skill directory layout — `.claude/skills/`, `.agents/skills/`, `.opencode/skills/`, `.cursor/skills/`, and so on — with no shared tooling for creation, validation, deduplication, or cross-platform installation.

Skern provides:

- **Reusable skill definitions** — one `SKILL.md` per skill, portable across platforms
- **Project-scoped or user-scoped registration** — local skills for a repo, global skills for your machine
- **Overlap detection** — fuzzy matching prevents skill duplication before it happens
- **Cross-platform installation** — install to any supported platform with a single command
- **Dynamic skill loading** — batch install/uninstall plus per-platform capacity reporting so agents can size their working set to a context budget
- **Agent-operable interface** — every command supports `--json`, enabling agents to manage their own skills

## Design Principles

- **CLI-first** — terminal is the primary interface
- **File-system native** — skills are files, registries are directories
- **Declarative platform registry** — adding a platform is a one-line append to `internal/platform/spec.go`
- **Agent-agnostic** — works with any platform that reads `SKILL.md`
- **Deterministic outputs** — same input, same result
- **Minimal dependencies** — small binary, fast startup
- **No cloud lock-in** — everything is local, everything is yours

## Philosophy

Skills should not live inside models. They should live in code. Versioned. Composable. Auditable. Portable.

## What's Shipped

The current release is **v0.2.1**. Major capabilities already in place:

- Eight platform adapters (Claude Code, Codex CLI, OpenCode, Cursor, Gemini CLI, GitHub Copilot, Windsurf, Continue)
- Cross-platform binaries for macOS, Linux, and Windows
- Skill versioning with semver (`skern skill version --bump major|minor|patch`)
- `skill diff` and `skill import` (URL / GitHub repo / gist)
- Stylistic lint hints in `skill validate`
- Batch install/uninstall and per-platform capacity reporting

See the [CHANGELOG](https://github.com/devrimcavusoglu/skern/blob/main/CHANGELOG.md) for full release notes and the project's [GitHub issues](https://github.com/devrimcavusoglu/skern/issues) for the live roadmap (MCP server mode, remote catalog search, skill dependencies, etc.).

## Next Steps

- [Installation](/guide/installation) — install skern on your machine
- [Quick Start](/guide/quick-start) — create and install your first skill
- [Agent Setup](/guide/agent-setup) — enable the tool-forming loop
- [Writing Skills](/writing-skills) — author skills agents can actually find and use
