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

# Create a new skill (overlap detection warns on similar existing skills)
scribe skill create my-skill --description "Automates X for Y"

# Create with author provenance
scribe skill create my-skill --author "alice" --author-type human --description "Automates X"

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

**`skill create` flags:** `--author`, `--author-type human|agent`, `--author-platform`, `--description`, `--force` (bypass overlap block)

### Validation

`scribe skill validate <name>` checks skills against the Agent Skills spec:

- **Name format** — must match `[a-z0-9]+(-[a-z0-9]+)*`, 1-64 characters
- **Description** — required, max 1024 characters
- **Body** — SKILL.md must have non-empty body content
- **Allowed-tools** — no empty entries
- **Metadata** — author type must be `human` or `agent`, version should follow semver

Validation also runs automatically during `scribe skill create`, issuing warnings for any issues.

### Overlap Detection

When creating a skill, scribe checks existing skills for similarity using:

- **Fuzzy name matching** — Levenshtein distance with prefix/suffix bonuses
- **Description similarity** — keyword overlap scoring (Jaccard similarity)
- **Tools overlap** — shared `allowed-tools` entries

| Score | Behavior |
|-------|----------|
| < 0.6 | Proceed normally |
| >= 0.6 | Warn — show similar skills, continue |
| >= 0.9 | Block — require `--force` to override |

Skill count warnings trigger at > 20 skills (project scope) or > 50 skills (user scope).

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

Apache 2.0 — see [LICENSE](LICENSE) for details.
