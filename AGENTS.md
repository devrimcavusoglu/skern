# AGENTS.md — Skern Development Guide

## Project Overview

Skern is a minimal, agent-first CLI tool for managing Agent Skills across agentic development platforms (Claude Code, Codex CLI, OpenCode, Cursor, Gemini CLI, GitHub Copilot, Windsurf, Continue, and more). It follows the Agent Skills open standard (agentskills.io) and uses `SKILL.md` files with YAML frontmatter as the canonical format.

The project is written in **Go 1.25+** and the current release is **v0.3.0**.

## Repository Layout

```
cmd/skern/main.go           # Entry point
internal/
  cli/                        # Cobra command definitions (root, version, init, completion, skill_*, platform_*)
  skill/                      # Domain logic: Skill struct, manifest parsing, validation, scaffolding
  overlap/                    # Fuzzy name matching and description similarity scoring
  registry/                   # Filesystem CRUD over ~/.skern/skills/ and .skern/skills/
  platform/                   # Platform adapters — declarative spec table + generic Adapter
  output/                     # JSON/text structured output formatting
scripts/
  install.sh                  # Installer script
  install_test.sh             # Installer tests
  smoke_test.sh               # Smoke tests for built binary
tests/manual/                 # Manual (agent-driven) test scenarios
docs/                         # Documentation site (VitePress)
go.mod, go.sum
Makefile
.goreleaser.yaml
.golangci.yml
.github/
  workflows/ci.yml
  workflows/release.yml
  workflows/docs-deploy.yml
  workflows/docs-pr-check.yml
  CODEOWNERS
```

## Build, Test & Lint

```sh
make build                  # Build binary with version/commit/date injected
make test                   # go test ./...
make test-v                 # Verbose test output
make test-cover             # Generate coverage report (coverage.out + coverage.html)
make test-install           # Run installer script tests
make test-smoke             # Build binary then run smoke tests
make test-manual-setup      # Set up manual test scenarios
make test-manual-report     # Report manual test results
make test-manual-teardown   # Clean up manual test environment
make lint                   # golangci-lint run
make fmt                    # gofmt -w .
make clean                  # Remove binary and coverage files
```

Tests use stdlib `testing` + `testify`. Follow table-driven test patterns. Integration tests should use temporary directories to simulate filesystem layouts.

Linter configuration lives in `.golangci.yml`.

## Issue Tracking Workflow

Development is tracked using **GitHub Issues** via the `gh` CLI.

```sh
gh issue list                             # List open issues
gh issue create --title "Title" --body "" # Create a new issue
gh issue view <number>                    # View issue details
gh issue close <number>                   # Close an issue
gh issue edit <number> --add-label "bug"  # Add labels
```

Reference issues in commit messages as `#<number>` (e.g. `Fix validation edge case (#42)`).

## Branching Strategy

All work is organized by milestone using feature branches:

- **Branch naming**: `feature/m<N>-<slug>` (e.g., `feature/m1-skill-registry`)
- **Created from**: `main`
- **Merged back via**: Pull request to `main`

Each milestone gets its own feature branch. All commits for that milestone go on the branch, then a PR merges everything back to `main` upon completion.

## Code Conventions

### Go Style

- Follow standard Go idioms and `gofmt` formatting
- Exported names use `CamelCase`; unexported use `camelCase`
- Prefer stdlib over third-party packages unless there is a strong reason
- Keep packages small and focused on a single responsibility
- Use `internal/` to prevent external imports of implementation details

### Package Responsibilities

- **`cli/`** — Only command wiring, flag parsing, and output. No business logic.
- **`skill/`** — Domain types and operations. The `Skill` struct, `Author`, `ModifiedByEntry` types, manifest parsing/serialization, validation rules, and scaffolding templates.
- **`registry/`** — Filesystem operations for skill storage. CRUD and discovery across user/project scopes.
- **`platform/`** — Declarative `Spec` table (`spec.go`) plus a single generic `Adapter` (`adapter.go`) that implements the `Platform` interface (`Name()`, `Detect()`, `UserSkillsDir()`, `ProjectSkillsDir()`, `Install()`, `Uninstall()`, `InstalledSkills()`) from any spec row. Adding a platform = one row in `Specs` plus one `Type` constant; no per-platform Go file.
- **`overlap/`** — Similarity scoring (Levenshtein distance, keyword overlap). Returns a float64 score in [0, 1].
- **`output/`** — Handles `--json` and `--quiet` flags. All commands go through this package for consistent formatting.

### Testing Conventions

- Use table-driven tests with descriptive subtest names
- Use `testify/assert` and `testify/require` for assertions
- Integration tests that touch the filesystem must use `t.TempDir()`
- Name test files as `*_test.go` in the same package

### Error Handling

- Return `error` values; do not panic
- Wrap errors with `fmt.Errorf("context: %w", err)` for stack tracing
- Use semantic exit codes: 0 = success, 1 = error, 2 = validation failure

### CLI Output

- Every command must support `--json` for machine-readable output
- Default output is human-friendly text
- Use `--quiet` to suppress non-essential output
- Error messages should include actionable suggestions

### Commit Messages

- Keep the subject line concise and imperative ("Add manifest parser", not "Added manifest parser")
- Reference GitHub issues when applicable: `#<number>`

## Architecture Notes

### Design Decisions

1. **SKILL.md as the canonical format** — Skern does NOT invent its own `skill.yaml`. It reads and writes `SKILL.md` files directly, matching the Agent Skills spec. A skill is a directory containing a `SKILL.md` and optional supporting files.

2. **Skern registry = filesystem directory** — `~/.skern/skills/` stores user-level skills. `.skern/skills/` stores project-level skills. No database, no daemon, no lock files.

3. **Platform adapters are copiers** — Installing a skill to a platform means copying the skill directory to the platform's expected location. Each adapter knows its platform's directory convention.

4. **Platform auto-detection** — Skern detects which platforms are installed by checking each adapter's home-relative `DetectHome` paths (e.g. `~/.claude/`, `~/.cursor/`, `~/.gemini/`). Detection is per-platform even when several adapters share `.agents/skills/` as their project directory. Each `install`/`uninstall` invocation targets exactly one platform (per #52 D6); `--platform all` is not accepted. Agents specify the platform they are running on, and `skill install`/`skill uninstall` accept multiple skill names per call for batch operations.

5. **JSON output as first-class** — Every command supports `--json` for machine-readable output. Default is human-friendly text. Exit codes are semantic: 0=success, 1=error, 2=validation failure.

### Tool-Forming Loop

The core differentiator of skern is enabling a **tool-forming loop** — agents don't just *use* skills, they *create* them when a recurring need arises:

```
Agent identifies a recurring need
  --> skern skill search <query>
  --> no results (or low similarity)
  --> skern skill create <name>
  --> Agent implements the skill
  --> Skill becomes reusable
```

On subsequent encounters, the agent finds the existing skill via `skern skill search` and reuses it.

**Guardrails:**

| Guardrail | Mechanism | Default |
|---|---|---|
| Overlap warning threshold | Similarity score 0.0–1.0 | Warn at >= 0.6 |
| Overlap block threshold | Similarity score | Block at >= 0.9, require `--force` |
| Skill count warning (project) | Count in `.skern/skills/` | Warn at >= 20 |
| Skill count warning (user) | Count in `~/.skern/skills/` | Warn at >= 50 |
| Per-platform capacity (project scope) | Count installed at `<platform>/.skern/skills/` (project) | Warn at >= 20 |
| Per-platform capacity (user scope) | Count installed at user-level platform skills dir | Warn at >= 50 |
| Capacity enforcement (install) | `--enforce-budget` | Off by default; refuses install when threshold would be exceeded |
| Deduplication hints | On `skern skill list` | Flag potential duplicates |

Capacity thresholds are defined in `internal/skill/capacity.go`. Every `install`/`uninstall` JSON response carries a `capacity` block (count, threshold, headroom, over-budget flag) so agents can react to capacity pressure without an extra query.

### SKILL.md Format (Agent Skills Spec)

```markdown
---
name: skill-name
description: |
  Use when <triggering conditions for this skill>.
allowed-tools: []
metadata:
  author:
    name: author-name
    type: human           # human | agent
    platform: claude-code  # only when type=agent
  version: "0.0.1"
  modified-by:            # append-only provenance list
    - name: codex-cli
      type: agent
      platform: codex-cli
      date: "2025-07-15T10:30:00Z"
---

## Overview

1-2 sentence core principle of the skill.

## When to Use

- Triggering conditions and symptoms

## Core Pattern

The main technique or pattern (before/after for techniques).

## Quick Reference

- Scannable summary for fast lookup

## Common Mistakes

- Frequent errors and fixes
```

Required fields: `name`, `description`. Directory name must match the `name` field.

### Writing Skills Guidelines

When creating or editing skills (via `skern skill create` or manually), follow these guidelines adapted from [superpowers/writing-skills](https://github.com/obra/superpowers/tree/main/skills/writing-skills). For full details, see [docs/writing-skills.md](docs/writing-skills.md).

Key points:

- **Description**: Start with "Use when..." — describe triggering conditions, not a workflow summary
- **Naming**: `kebab-case`, verb-first active voice (`creating-skills` not `skill-creation`)
- **Body structure**: Overview → When to Use → Core Pattern → Quick Reference → Common Mistakes
- **Token budget**: Getting-started < 150 words; frequently-loaded < 200 words; others < 500 words
- **Examples**: One excellent example beats many mediocre ones

### Skill Name Validation

Names must match `^[a-z0-9]+(-[a-z0-9]+)*$` and be 1-64 characters.

### Registry Paths

| Scope   | Path                    |
|---------|-------------------------|
| User    | `~/.skern/skills/`     |
| Project | `.skern/skills/`       |

### Platform Paths

| Adapter name     | User-level                          | Project-level              |
|------------------|--------------------------------------|----------------------------|
| `claude-code`    | `~/.claude/skills/<name>/`          | `.claude/skills/<name>/`   |
| `codex-cli`      | `~/.agents/skills/<name>/`          | `.agents/skills/<name>/`   |
| `opencode`       | `~/.config/opencode/skills/<name>/` | `.opencode/skills/<name>/` |
| `cursor`         | `~/.cursor/skills/<name>/`          | `.agents/skills/<name>/`   |
| `gemini-cli`     | `~/.gemini/skills/<name>/`          | `.agents/skills/<name>/`   |
| `github-copilot` | `~/.copilot/skills/<name>/`         | `.agents/skills/<name>/`   |
| `windsurf`       | `~/.codeium/windsurf/skills/<name>/`| `.windsurf/skills/<name>/` |
| `continue`       | `~/.continue/skills/<name>/`        | `.continue/skills/<name>/` |

The full list is generated from `internal/platform/spec.go` — append a row there to add a platform.

### Overlap Detection Thresholds

- Score < 0.6 — proceed normally
- Score >= 0.6 — warn, show similar skills
- Score >= 0.9 — block creation, require `--force`

### Dependencies

| Dependency | Purpose |
|------------|---------|
| `github.com/spf13/cobra` | CLI framework |
| `gopkg.in/yaml.v3` | YAML frontmatter parsing |
| `santhosh-tekuri/jsonschema/v6` | Agent Skills spec validation (planned) |
| `github.com/stretchr/testify` | Test assertions |

## Current Status

Milestones M0–M7 are complete (v0.3.0).

- **M6** (v0.2.0) — dynamic skill loading per #52: batch install/uninstall, capacity reporting in install/uninstall output, `--enforce-budget` opt-in, `--with-platforms` flag on `skill list`, removal of `--platform all`. The v0.1.x → v0.2.0 transition introduced a breaking change to the JSON shape of install/uninstall results (`skills[]` + top-level `platform`/`capacity` instead of `platforms[]`).
- **M7** (v0.3.0) — declarative platform registry: per-platform Go files collapsed into a single `Spec` table (`internal/platform/spec.go`) plus a generic `Adapter`. Five new adapters (`cursor`, `gemini-cli`, `github-copilot`, `windsurf`, `continue`) shipped alongside the rewrite. Also in v0.3.0: `skern init --instructions` writes the skern usage snippet into agent instruction files, `--from-template` requires a skill directory (small breaking change), and Windows joins the CI matrix.

### Future Roadmap

These items are tracked as GitHub issues:

- MCP server mode (`skern serve`) — expose skills as MCP tools
- Community skill catalog integration
- Remote catalog search in `skern skill search`
- Skill dependencies and composition (#45)
- Platform-specific skill variants (#47)
- Skill sync: bulk reconcile registry with platforms (#48)
- Dry run / preview mode for mutating commands (#51)
- WASI/Docker execution backends
- LRU usage tracking for dynamic loading (#52 Phase 3) — state file at `~/.skern/state/usage.json`, `skern skill touch`/`skern skill evict` commands; deferred until the agent-side "skill use" signal is settled
- Skill stats for context optimization (#75) — byte size, cross-platform presence, future token estimation

