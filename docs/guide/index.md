# Getting Started

Skern is a minimal, agent-first CLI for managing [Agent Skills](https://agentskills.io) across agentic development platforms. It provides a standardized lifecycle — create, validate, install, remove — for skills that work natively with **Claude Code**, **Codex CLI**, and **OpenCode**.

Skills follow the Agent Skills open standard (`SKILL.md` with YAML frontmatter) and are immediately usable by any compatible platform without adapters or format conversion.

## Why skern?

Modern AI coding tools (Claude Code, Codex CLI, OpenCode) lack a standardized skill management layer. Each defines skills in its own directory structure, with no shared tooling for creation, validation, deduplication, or cross-platform installation.

Skern provides:

- **Reusable skill definitions** — one `SKILL.md` per skill, portable across platforms
- **Project-scoped or system-scoped registration** — local skills for a repo, global skills for your machine
- **Overlap detection** — fuzzy matching prevents skill duplication before it happens
- **Cross-platform installation** — install to any supported platform with a single command
- **Agent-operable interface** — every command supports `--json`, enabling agents to manage their own skills

## Design Principles

- **CLI-first** — terminal is the primary interface
- **File-system native** — skills are files, registries are directories
- **Agent-agnostic** — works with any platform that reads `SKILL.md`
- **Deterministic outputs** — same input, same result
- **Minimal dependencies** — small binary, fast startup
- **No cloud lock-in** — everything is local, everything is yours

## Philosophy

Skills should not live inside models. They should live in code. Versioned. Composable. Auditable. Portable.

## Next Steps

- [Installation](/guide/installation) — install skern on your machine
- [Quick Start](/guide/quick-start) — create and install your first skill
- [Agent Setup](/guide/agent-setup) — enable the tool-forming loop
