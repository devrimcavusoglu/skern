# scribe

A minimal, agent-first CLI for managing [Agent Skills](https://agentskills.io) across agentic development platforms.

Scribe provides a standardized lifecycle — create, validate, install, remove — for skills that work natively with **Claude Code**, **Codex CLI**, and **OpenCode**. Skills follow the Agent Skills open standard (`SKILL.md` with YAML frontmatter) and are immediately usable by any compatible platform without adapters or format conversion.

## Features

- **Unified skill management** — one CLI to create, validate, search, install, and remove skills across all supported platforms
- **Agent Skills spec compliance** — reads and writes `SKILL.md` files directly, no proprietary format
- **Platform adapters** — install skills to Claude Code, Codex CLI, or OpenCode with a single command
- **Tool-forming loop** — agents can search for existing skills and scaffold new ones, turning recurring needs into reusable capabilities
- **Overlap detection** — fuzzy name matching and description similarity prevent skill duplication
- **JSON output** — every command supports `--json` for machine-readable output, making scribe fully agent-operable

## Installation

### From source

Requires Go 1.23+.

```sh
go install github.com/devrimcavusoglu/scribe/cmd/scribe@latest
```

### Build from repository

```sh
git clone https://github.com/devrimcavusoglu/scribe.git
cd scribe
make build
```

## Quick Start

```sh
# Check installation
scribe version

# Initialize scribe in your project
scribe init

# Create a new skill
scribe skill create my-skill --description "Automates X for Y"

# List skills
scribe skill list

# Validate a skill against the Agent Skills spec
scribe skill validate my-skill

# Install a skill to a platform
scribe skill install my-skill --platform claude-code

# Search for existing skills
scribe skill search "code review"
```

## CLI Reference

```
scribe init                                    # Initialize .scribe/ in current project
scribe skill create <name>                     # Scaffold a new skill
scribe skill search <query>                    # Search skills by name/description
scribe skill list [--scope user|project|all]   # List skills in registry
scribe skill show <name>                       # Display skill details
scribe skill validate <name>                   # Validate against Agent Skills spec
scribe skill remove <name>                     # Remove skill from registry
scribe skill install <name> --platform <p>     # Install skill to platform
scribe skill uninstall <name> --platform <p>   # Remove skill from platform
scribe platform list                           # List detected platforms
scribe platform status                         # Skill x platform installation matrix
scribe version                                 # Print version info
```

**Global flags:** `--json`, `--quiet`, `--scope user|project`

## Supported Platforms

| Platform | User-level skills | Project-level skills |
|----------|-------------------|----------------------|
| Claude Code | `~/.claude/skills/<name>/` | `.claude/skills/<name>/` |
| Codex CLI | `~/.agents/skills/<name>/` | `.agents/skills/<name>/` |
| OpenCode | `~/.config/opencode/skills/<name>/` | `.opencode/skills/<name>/` |

Scribe auto-detects which platforms are installed. Use `--platform all` to install a skill to every detected platform at once.

## Development

```sh
# Run tests
make test

# Run tests with coverage
make test-cover

# Lint
make lint

# Format
make fmt

# Build
make build

# Clean build artifacts
make clean
```

## License

MIT
