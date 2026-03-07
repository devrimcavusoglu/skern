# AGENTS.md — Skern Development Guide

## Project Overview

Skern is a minimal, agent-first CLI tool for managing Agent Skills across agentic development platforms (Claude Code, Codex CLI, OpenCode). It follows the Agent Skills open standard (agentskills.io) and uses `SKILL.md` files with YAML frontmatter as the canonical format.

The project is written in **Go 1.25+** and the current release is **v0.1.0**.

## Repository Layout

```
cmd/skern/main.go           # Entry point
internal/
  cli/                        # Cobra command definitions (root, version, init, completion, skill_*, platform_*)
  skill/                      # Domain logic: Skill struct, manifest parsing, validation, scaffolding
  overlap/                    # Fuzzy name matching and description similarity scoring
  registry/                   # Filesystem CRUD over ~/.skern/skills/ and .skern/skills/
  platform/                   # Platform adapters (Claude Code, Codex CLI, OpenCode)
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
- **`platform/`** — Each adapter implements the `Platform` interface: `Name()`, `Detect()`, `UserSkillsDir()`, `ProjectSkillsDir()`, `Install()`, `Uninstall()`, `InstalledSkills()`.
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

4. **Platform auto-detection** — Skern detects which platforms are installed by checking for their config directories/binaries (`~/.claude/`, `~/.codex/` or `~/.agents/`, `~/.config/opencode/`). `--platform all` installs to every detected platform.

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
| Skill count warning (project) | Count in `.skern/skills/` | Warn at > 20 |
| Skill count warning (user) | Count in `~/.skern/skills/` | Warn at > 50 |
| Deduplication hints | On `skern skill list` | Flag potential duplicates |

### SKILL.md Format (Agent Skills Spec)

```markdown
---
name: skill-name
description: |
  What this skill does and when to use it
allowed-tools: []
metadata:
  author:
    name: author-name
    type: human           # human | agent
    platform: claude-code  # only when type=agent
  version: "0.1.0"
  modified-by:            # append-only provenance list
    - name: codex-cli
      type: agent
      platform: codex-cli
      date: "2025-07-15T10:30:00Z"
---

## Instructions

Step-by-step instructions for the agent.
```

Required fields: `name`, `description`. Directory name must match the `name` field.

### Skill Name Validation

Names must match `^[a-z0-9]+(-[a-z0-9]+)*$` and be 1-64 characters.

### Registry Paths

| Scope   | Path                    |
|---------|-------------------------|
| User    | `~/.skern/skills/`     |
| Project | `.skern/skills/`       |

### Platform Paths

| Platform    | User-level                         | Project-level            |
|-------------|-------------------------------------|--------------------------|
| Claude Code | `~/.claude/skills/<name>/`         | `.claude/skills/<name>/` |
| Codex CLI   | `~/.agents/skills/<name>/`         | `.agents/skills/<name>/` |
| OpenCode    | `~/.config/opencode/skills/<name>/`| `.opencode/skills/<name>/`|

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

All milestones (M0–M5) are complete. The current release is **v0.1.0**, which added `skill edit`, skill tags/categories, `--force` install, parse warnings in `skill list`, stylistic lint hints, and internal refactors (CommandContext, unified scoring, output split).

### Future Roadmap

These items are tracked as GitHub issues:

- MCP server mode (`skern serve`) — expose skills as MCP tools
- Skill import from URL / git repo
- Skill versioning (semver in frontmatter, upgrade detection)
- Community skill catalog integration
- Remote catalog search in `skern skill search`
- Skill dependency resolution
- WASI/Docker execution backends

