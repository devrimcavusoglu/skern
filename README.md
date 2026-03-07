<p align="center">
  <img src="logo.png" alt="skern" width="360" />
</p>

<p align="center">
  <strong>System-wide skill registry for AI agents.</strong><br/>
  Forge, manage, and compose agent capabilities from the terminal.
</p>

<p align="center">
  <a href="https://github.com/devrimcavusoglu/skern/releases"><img src="https://img.shields.io/github/v/release/devrimcavusoglu/skern" alt="Release"></a>
  <a href="https://github.com/devrimcavusoglu/skern/blob/main/LICENSE"><img src="https://img.shields.io/github/license/devrimcavusoglu/skern" alt="License"></a>
  <img src="https://img.shields.io/badge/type-CLI-black" alt="CLI">
  <a href="https://agentskills.io"><img src="https://img.shields.io/badge/spec-Agent%20Skills-blue" alt="Agent Skills"></a>
  <a href="https://skern.dev"><img src="https://img.shields.io/badge/docs-skern.dev-blue" alt="Docs"></a>
</p>

---

Skern is a minimal, agent-first CLI for managing [Agent Skills](https://agentskills.io) across **Claude Code**, **Codex CLI**, and **OpenCode**. One `SKILL.md` per skill — portable, validated, and instantly installable to any supported platform.

## Quick Example

```bash
skern init
skern skill create code-review --description "Review PRs for style and correctness"
skern skill install code-review --platform all
```

## Features

- **Unified skill lifecycle** — create, validate, search, install, and remove across platforms
- **Agent Skills spec** — reads and writes `SKILL.md` directly, no proprietary format
- **Cross-platform** — install to Claude Code, Codex CLI, or OpenCode in one command
- **Tool-forming loop** — agents scaffold and reuse skills automatically
- **Overlap detection** — fuzzy matching prevents duplication
- **JSON output** — every command supports `--json` for agent-operable workflows

## Install

```sh
curl -fsSL https://raw.githubusercontent.com/devrimcavusoglu/skern/main/scripts/install.sh | bash
```

Or with Go 1.25+:

```sh
go install github.com/devrimcavusoglu/skern/cmd/skern@latest
```

## Documentation

Full documentation is available at **[skern.dev](https://skern.dev)**:

- [Getting Started](https://skern.dev/guide/) — why skern, design principles
- [Installation](https://skern.dev/guide/installation) — all installation methods
- [Quick Start](https://skern.dev/guide/quick-start) — first workflow walkthrough
- [CLI Reference](https://skern.dev/reference/commands) — full command reference
- [Architecture](https://skern.dev/concepts/) — how the layers fit together
- [Platforms](https://skern.dev/platforms/) — platform-specific details
- [Contributing](https://skern.dev/contributing/) — development setup and testing

## License

Apache 2.0 — see [LICENSE](LICENSE) for details.
