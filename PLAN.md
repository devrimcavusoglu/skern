# SCRIBE — Project Plan

## Context

AI development agents (Claude Code, Codex CLI, OpenCode) can execute commands and generate code, but lack a first-class concept of reusable, versionable skills/tools. Developers end up with ad-hoc scripts, hidden prompts, and zero governance. A critical finding from researching all three target platforms is that they are **converging on the Agent Skills open standard** ([agentskills.io](https://agentskills.io)) — all use `SKILL.md` files with YAML frontmatter. Scribe builds on this shared foundation rather than inventing a new format. Beyond managing skills, scribe enables a **tool-forming loop** — agents can search for existing skills, and when none match, scaffold and implement new ones, turning recurring needs into reusable capabilities.

## Project Description

**scribe** is a minimal, agent-first CLI tool for managing Agent Skills across agentic development platforms. It provides a standardized lifecycle (create, validate, install, remove) for skills that work natively with Claude Code, Codex CLI, and OpenCode. Skills created by scribe follow the Agent Skills open standard and are immediately usable by any compatible platform — no adapters or format conversion needed.

---

## Technical Stack

| Component | Choice | Rationale |
|-----------|--------|-----------|
| Language | **Go 1.23+** | Single binary, official MCP SDK, industry standard for CLIs |
| CLI Framework | **Cobra** | De facto standard (kubectl, helm, docker), subcommand hierarchy |
| YAML Parsing | **`gopkg.in/yaml.v3`** | Stdlib-quality, handles YAML frontmatter extraction |
| JSON Schema | **`santhosh-tekuri/jsonschema/v6`** | Validates skill manifests against Agent Skills spec |
| Filesystem | **stdlib `os`/`filepath`** | No abstraction layer needed for v1; `afero` if testing demands it |
| Testing | **stdlib `testing` + `testify`** | Table-driven tests, assertion helpers |
| Build/Release | **GoReleaser** | Cross-platform builds, Homebrew tap, GitHub releases |
| CI | **GitHub Actions** | Lint (`golangci-lint`), test, build matrix (macOS/Linux) |
| Issue Tracking | **beads-rust (`br`)** | Agent-first issue tracker, SQLite + JSONL, CLI-native |

---

## Architecture

```
scribe/
├── cmd/scribe/
│   └── main.go                    # Entry point
├── internal/
│   ├── cli/                       # Cobra command definitions
│   │   ├── root.go                # Root command, global flags (--json, --quiet)
│   │   ├── version.go             # Version command
│   │   ├── init.go                # scribe init
│   │   ├── completion.go          # scribe completion (bash, zsh, fish)
│   │   ├── skill.go               # skill subcommand group
│   │   ├── skill_create.go        # scribe skill create <name>
│   │   ├── skill_search.go        # scribe skill search <query>
│   │   ├── skill_list.go          # scribe skill list (with dedup hints)
│   │   ├── skill_show.go          # scribe skill show <name> (with modified-by)
│   │   ├── skill_validate.go      # scribe skill validate <name>
│   │   ├── skill_remove.go        # scribe skill remove <name>
│   │   ├── skill_install.go       # scribe skill install <name> --platform <p>
│   │   ├── skill_uninstall.go     # scribe skill uninstall <name> --platform <p>
│   │   ├── platform.go            # platform subcommand group
│   │   ├── platform_list.go       # scribe platform list
│   │   ├── platform_status.go     # scribe platform status
│   │   └── e2e_test.go            # End-to-end lifecycle integration tests
│   ├── skill/                     # Skill domain logic
│   │   ├── skill.go               # Skill struct, Author, ModifiedByEntry types
│   │   ├── manifest.go            # SKILL.md frontmatter parsing/serialization
│   │   ├── validator.go           # Validate against Agent Skills spec
│   │   └── scaffold.go            # Skill scaffolding templates
│   ├── overlap/                   # Skill overlap detection
│   │   ├── detector.go            # Fuzzy name matching, description similarity
│   │   └── scorer.go              # Overlap scoring heuristics
│   ├── registry/                  # Local skill registry
│   │   ├── registry.go            # CRUD over filesystem registry
│   │   └── discovery.go           # Discover skills across locations
│   ├── platform/                  # Platform adapters
│   │   ├── platform.go            # Platform interface definition
│   │   ├── claude.go              # Claude Code adapter
│   │   ├── codex.go               # Codex CLI adapter
│   │   ├── opencode.go            # OpenCode adapter
│   │   └── detector.go            # Auto-detect installed platforms
│   └── output/                    # Structured output
│       └── output.go              # JSON/text formatting, error types
├── go.mod
├── go.sum
├── Makefile
├── .goreleaser.yaml
├── .github/workflows/ci.yml
└── .golangci.yml
```

### Key Design Decisions

**1. SKILL.md as the canonical format** — Scribe does NOT invent its own `skill.yaml`. It reads and writes `SKILL.md` files directly, matching the Agent Skills spec. A skill is a directory containing a `SKILL.md` and optional supporting files.

**2. Scribe registry = filesystem directory** — `~/.scribe/skills/` stores user-level skills. `.scribe/skills/` stores project-level skills. No database, no daemon, no lock files.

**3. Platform adapters are copiers/linkers** — Installing a skill to a platform means copying (or symlinking) the skill directory to the platform's expected location. Each adapter knows its platform's directory convention:

| Platform | User-level path | Project-level path |
|----------|----------------|-------------------|
| Claude Code | `~/.claude/skills/<name>/` | `.claude/skills/<name>/` |
| Codex CLI | `~/.agents/skills/<name>/` | `.agents/skills/<name>/` |
| OpenCode | `~/.config/opencode/skills/<name>/` | `.opencode/skills/<name>/` |

**4. Platform auto-detection** — Scribe detects which platforms are installed by checking for their config directories/binaries (`~/.claude/`, `~/.codex/` or `~/.agents/`, `~/.config/opencode/`). `scribe skill install --platform all` installs to every detected platform.

**5. JSON output as first-class** — Every command supports `--json` for machine-readable output. Default is human-friendly text. Exit codes are semantic: 0=success, 1=error, 2=validation failure.

---

## Tool-Forming Loop

The core differentiator of scribe is enabling a **tool-forming loop** — agents don't just *use* skills, they *create* them when a recurring need arises. This turns one-off solutions into reusable, discoverable capabilities.

### The Loop

```
Agent identifies a recurring need
  --> scribe skill search <query>
  --> no results (or low similarity)
  --> scribe skill create <name>
  --> CLI scaffolds SKILL.md, tests, permissions, docs
  --> Agent implements
  --> Skill becomes reusable
```

On subsequent encounters, the agent finds the existing skill via `scribe skill search` and reuses it instead of reimplementing.

### Overlap Detection

Before creating a new skill, `scribe skill create` implicitly checks for overlap to prevent skill bloat and duplication:

- **Fuzzy name matching** — Levenshtein distance, prefix/suffix overlap
- **Description similarity** — keyword overlap scoring
- **`allowed-tools` overlap** — skills wrapping the same underlying tools

Three outcomes based on similarity score:

1. **Proceed** — no similar skills found (score < 0.6)
2. **Warn** — similar skills listed, agent/user can proceed or choose existing (score >= 0.6)
3. **Block** — near-duplicate detected, requires `--force` to override (score >= 0.9)

### Guardrails

| Guardrail | Mechanism | Default |
|---|---|---|
| Overlap warning threshold | Similarity score 0.0–1.0 | Warn at >= 0.6 |
| Overlap block threshold | Similarity score | Block at >= 0.9, require `--force` |
| Skill count warning (project) | Count in `.scribe/skills/` | Warn at > 20 |
| Skill count warning (user) | Count in `~/.scribe/skills/` | Warn at > 50 |
| Deduplication hints | On `scribe skill list` | Flag potential duplicates |

### Search Behavior

`scribe skill search <query>` searches across project and user scopes, returning ranked results with similarity scores. Supports `--json` for agent-programmatic decisions (e.g., an agent can parse the JSON output to decide whether to create a new skill or reuse an existing one).

---

## CLI Command Reference (v1)

```
scribe init                                    # Initialize .scribe/ in current project
scribe skill create <name>                     # Scaffold a new skill (SKILL.md + directory)
scribe skill search <query>                    # Search skills by name/description (ranked results)
scribe skill list [--scope user|project|all]   # List skills in registry
scribe skill show <name>                       # Display skill details (frontmatter + body)
scribe skill validate <name>                   # Validate SKILL.md against Agent Skills spec
scribe skill remove <name>                     # Remove skill from scribe registry
scribe skill install <name> --platform <p>     # Copy/link skill to platform directory
scribe skill uninstall <name> --platform <p>   # Remove skill from platform directory
scribe platform list                           # List detected platforms and their paths
scribe platform status                         # Show installation status per skill per platform
scribe completion [bash|zsh|fish]              # Generate shell completion scripts
scribe version                                 # Print version info
```

Global flags: `--json`, `--quiet`, `--scope user|project` (default: project if in a project, user otherwise)

`scribe skill create` flags:
- `--author <name>` — author name/identifier
- `--author-type human|agent` — defaults to `human`
- `--author-platform <platform>` — required when `--author-type agent`; one of `claude-code`, `codex-cli`, `opencode` (extensible)
- `--description <text>` — skill description
- `--from-template <template>` — scaffold from a template
- `--force` — bypass overlap detection block (score >= 0.9)

---

## Milestones

### M0 — Project Bootstrap [P0, Week 1] ✅

- [x] Brainstorming and PRD (done)
- [x] `br init` — initialize beads-rust issue tracker
- [x] Create `br` epics for M1–M2, seed initial issues
- [x] `go mod init github.com/devrimcavusoglu/scribe`
- [x] Cobra skeleton with root, version commands
- [x] Makefile (build, test, lint)
- [x] `.goreleaser.yaml` config
- [x] GitHub Actions CI (lint + test + build)
- [x] `.golangci.yml` linter config
- [x] `output` package — JSON/text formatter with `--json` flag support
- [x] Basic project README

**Exit criteria**: `scribe version` works, CI is green, `br` tracker initialized with epics

### M1 — Skill Manifest & Registry [P0, Week 2] ✅

- [x] `skill.Skill` struct matching Agent Skills spec frontmatter fields
- [x] `skill.Author` and `skill.ModifiedByEntry` types for structured author metadata
- [x] `manifest.Parse(path)` — extract YAML frontmatter from SKILL.md, including structured `author` object
- [x] `manifest.Write(skill, path)` — serialize skill to SKILL.md
- [x] `registry.Registry` — filesystem CRUD for `~/.scribe/skills/` and `.scribe/skills/`
- [x] `registry.Discovery` — walk and discover skills across scopes
- [x] `scribe skill create <name>` — scaffold `<name>/SKILL.md` with template
- [x] `--author`, `--author-type`, `--author-platform` flags on `scribe skill create`
- [x] `scribe skill search <query>` — basic name matching across project+user scopes
- [x] `scribe skill list` — list all skills with name, description, scope
- [x] `scribe skill show <name>` — display full skill contents
- [x] `scribe skill remove <name>` — delete skill directory
- [x] Unit tests for manifest parsing (including structured author) and registry operations

**Exit criteria**: Full CRUD cycle works — create, list, show, remove. Skill search returns basic name matches. JSON output for all commands.

### M2 — Skill Validation & Overlap Detection [P0, Week 3] ✅

- [x] `validator.Validate(skill)` — check required fields (name, description)
- [x] Name format validation (`^[a-z0-9]+(-[a-z0-9]+)*$`, 1-64 chars)
- [x] Description length validation (1-1024 chars)
- [x] Optional field type checking (allowed-tools, metadata, license)
- [x] Validate SKILL.md body is non-empty
- [x] `scribe skill validate <name>` — run all checks, report issues
- [x] Validate on create (warn on issues)
- [x] Structured validation error output (JSON array of issues)
- [x] `overlap.Detector` — fuzzy name matching (Levenshtein distance, prefix/suffix)
- [x] `overlap.Scorer` — description similarity scoring (keyword overlap)
- [x] Overlap check integrated into `scribe skill create` (warn/block flow based on thresholds)
- [x] `--force` flag to bypass overlap block (score >= 0.9)
- [x] Skill count threshold warnings (project > 20, user > 50)

**Exit criteria**: `scribe skill validate` catches malformed skills and reports clear, actionable errors. Overlap detection warns or blocks on similar skills during create.

### M3 — Platform Adapters [P0, Week 4] ✅

- [x] `platform.Platform` interface: `Name()`, `Detect()`, `UserSkillsDir()`, `ProjectSkillsDir()`, `Install(skill)`, `Uninstall(skill)`, `InstalledSkills()`
- [x] `platform.ClaudeCode` adapter — `.claude/skills/` conventions
- [x] `platform.CodexCLI` adapter — `.agents/skills/` conventions
- [x] `platform.OpenCode` adapter — `.opencode/skills/` conventions
- [x] `platform.Detector` — auto-detect installed platforms
- [x] `scribe skill install <name> --platform <p>` — copy skill directory to platform path
- [x] `scribe skill install <name> --platform all` — install to all detected platforms
- [x] `scribe skill uninstall <name> --platform <p>`
- [x] `scribe platform list` — show detected platforms with paths
- [x] `scribe platform status` — matrix view: skill x platform installation status
- [x] Integration tests with temp directories simulating platform layouts

**Exit criteria**: A skill created by scribe can be installed to all three platforms and is immediately discoverable by each tool.

### M4 — Agent Experience & Polish [P1, Week 5] ✅

- [x] `scribe init` — create `.scribe/` directory in project, idempotent initialization
- [x] Shell completions (bash, zsh, fish) via Cobra (`scribe completion`)
- [x] Semantic exit codes consistently applied (audit of all RunE paths)
- [x] `--quiet` flag suppresses non-essential output (already done in M0)
- [x] Error messages include actionable suggestions (e.g., "run 'scribe skill list'", "valid platforms: ...")
- [x] `scribe skill create` supports `--description` and `--from-template` flags
- [x] Deduplication hints in `scribe skill list` — pairwise overlap detection flags potential duplicates (score >= 0.6)
- [x] Author provenance display in `scribe skill show` — shows `modified-by` history with name, type, platform, date
- [x] GoReleaser cross-compilation verified (darwin/amd64, darwin/arm64, linux/amd64, linux/arm64)
- [x] Homebrew formula generation (`brews` section in `.goreleaser.yaml`)

**Exit criteria**: scribe can be installed via `brew install` or binary download. Agent-friendly in non-interactive mode. Deduplication hints visible on list.

### M5 — Release v0.0.1 [P1, Week 6]

- [x] End-to-end test: create skill -> validate -> install to all 3 platforms -> verify -> uninstall
- [ ] Manual testing with actual Claude Code, Codex CLI, OpenCode installations
- [x] CLAUDE.md / AGENTS.md for the scribe project itself (dogfooding)
- [x] OS-agnostic install script (`scripts/install.sh`)
- [x] GitHub Actions release workflow (`.github/workflows/release.yml`)
- [ ] Tag v0.0.1
- [ ] GitHub Release with GoReleaser artifacts

**Exit criteria**: v0.0.1 released, usable by all three target platforms.

### M6 — Post-v1 Roadmap [P2, Future]

- [ ] MCP server mode (`scribe serve`) — expose skills as MCP tools
- [ ] Skill import from URL / git repo
- [ ] Skill versioning (semver in frontmatter, upgrade detection)
- [ ] Community skill catalog integration
- [ ] Refine `scribe skill search` to also cover remote catalogs
- [ ] `scribe skill update` with `--author` flag appending to `modified-by` list
- [ ] Skill dependency resolution
- [ ] WASI/Docker execution backends

---

## Issue Tracking with `br`

Development of scribe is tracked using **beads-rust (`br`)**, an agent-first issue tracker that stores data in SQLite + JSONL and is fully operable via CLI.

### Key Commands

```
br init                        # Initialize tracker in project root
br create "Set up CI" --epic <epic-id>  # Create issue under epic
br list                        # List open issues
br list --all                  # List all issues (including closed)
br show <id>                   # Show issue details
br update <id> --status done   # Update issue status
br close <id>                  # Close an issue
br epic status                 # Show epic progress
br stats                       # Show project statistics
br sync --flush-only           # Flush local changes to JSONL
```

### Conventions

- Milestones map to `br` epics (M1: `bd-oaj`, M2: `bd-1vg`, M4: `bd-171`, M5: `bd-1ma`)
- Individual tasks within milestones are tracked as `br` issues
- Issue IDs use the `bd-<epic>.<n>` format (e.g., `bd-oaj.1`, `bd-1vg.5`)
- Agents can query `br list --json` to discover open tasks programmatically

### Current Status

- **35 total issues** (31 closed, 4 open)
- **4 epics tracked**: M1 (bd-oaj, 9 issues), M2 (bd-1vg, 10 issues), M4 (bd-171, 8 issues), M5 (bd-1ma, 4 issues)
- M0 and M3 were completed without `br` issue tracking

---

## Agent Skills Spec Alignment

Scribe's SKILL.md template follows the Agent Skills spec:

```markdown
---
name: <skill-name>
description: |
  <what this skill does and when to use it>
allowed-tools: []
metadata:
  author:
    name: <author-name>       # e.g. "Jane Doe" or "claude-code"
    type: human                # human | agent
    platform: claude-code      # only when type=agent; one of: claude-code, codex-cli, opencode
  version: "0.1.0"
  modified-by:                 # append-only list (like git blame)
    - name: "codex-cli"
      type: agent
      platform: codex-cli
      date: "2025-07-15T10:30:00Z"
---

## Instructions

<step-by-step instructions for the agent>
```

Required fields: `name`, `description`
Optional fields: `license`, `compatibility`, `metadata`, `allowed-tools`
Convention: directory name MUST match the `name` field

### Author Metadata

The `metadata.author` field is a structured object (not a plain string) to support provenance tracking:

- **`name`** — human name or agent identifier
- **`type`** — `human` or `agent`
- **`platform`** — required when `type` is `agent`; specific platform identifier (`claude-code`, `codex-cli`, `opencode`) rather than generic `agent`, enabling auditing of which platform created what

The `metadata.modified-by` list is append-only, recording each modification with the author info and a timestamp. This provides a lightweight provenance trail without requiring full version control integration.

**Architecture**: `Author` and `ModifiedByEntry` types live in `internal/skill/skill.go` alongside the existing `Skill` struct. No new files needed.

---

## Verification Plan

1. **Unit tests**: Each package has table-driven tests (manifest parsing, validation rules, registry CRUD, platform path resolution)
2. **Integration tests**: Temp directory-based tests simulating full workflows (create -> validate -> install -> list -> uninstall -> remove)
3. **Platform smoke tests**: Install a scribe-created skill into actual Claude Code/Codex CLI/OpenCode directories and verify the platform discovers it
4. **Overlap detection tests**: Verify fuzzy matching accuracy (Levenshtein, prefix/suffix), threshold behavior (warn at >= 0.6, block at >= 0.9), `--force` bypass, skill count warnings
5. **CI**: `go test ./...` on every push, `golangci-lint` for style, build matrix for macOS + Linux
6. **Dogfooding**: Use scribe to manage skills for the scribe project itself
