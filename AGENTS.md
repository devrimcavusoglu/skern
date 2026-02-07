# AGENTS.md — Scribe Development Guide

## Project Overview

Scribe is a minimal, agent-first CLI tool for managing Agent Skills across agentic development platforms (Claude Code, Codex CLI, OpenCode). It follows the Agent Skills open standard (agentskills.io) and uses `SKILL.md` files with YAML frontmatter as the canonical format.

The project is written in **Go 1.23+** and is currently in its bootstrap phase.

## Repository Layout

```
cmd/scribe/main.go           # Entry point
internal/
  cli/                        # Cobra command definitions (root, version, skill_*, platform_*)
  skill/                      # Domain logic: Skill struct, manifest parsing, validation, scaffolding
  overlap/                    # Fuzzy name matching and description similarity scoring
  registry/                   # Filesystem CRUD over ~/.scribe/skills/ and .scribe/skills/
  platform/                   # Platform adapters (Claude Code, Codex CLI, OpenCode)
  output/                     # JSON/text structured output formatting
go.mod, go.sum
Makefile
.goreleaser.yaml
.golangci.yml
.github/workflows/ci.yml
PLAN.md                       # Full project plan with milestones and architecture
```

## Build & Run

```sh
# Build
make build
# or directly:
go build -o scribe ./cmd/scribe

# Run
./scribe version
```

## Testing

```sh
# Run all tests
go test ./...

# Run tests for a specific package
go test ./internal/skill/...

# Run tests with verbose output
go test -v ./...

# Run tests with coverage
go test -coverprofile=coverage.out ./...
```

Tests use stdlib `testing` + `testify`. Follow table-driven test patterns. Integration tests should use temporary directories to simulate filesystem layouts.

## Linting & Formatting

```sh
# Format code
gofmt -w .

# Run linter
golangci-lint run

# Lint a specific package
golangci-lint run ./internal/skill/...
```

Configuration lives in `.golangci.yml`.

## Issue Tracking Workflow

Development is tracked using **beads-rust (`br`)**, an agent-first CLI issue tracker.

```sh
br issue list                     # List all open issues
br issue list --epic M0           # List issues for a milestone
br issue create "Title"           # Create a new issue
br issue update <id> --status done
br sync --flush-only              # Flush local changes
```

Each milestone (M0-M6) maps to a `br` epic. Reference issues in commit messages as `br#<id>`.

### Mandatory Workflow

1. **Before starting work**: Create a `br` epic for the milestone and individual issues for each task
2. **During development**: Update issue status (`in-progress`, `done`) as work progresses
3. **In commit messages**: Reference `br#<id>` to link commits to issues

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
- Reference `br` issues when applicable: `br#<id>`

## Architecture Notes

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
| User    | `~/.scribe/skills/`     |
| Project | `.scribe/skills/`       |

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
| `santhosh-tekuri/jsonschema/v6` | Agent Skills spec validation |
| `github.com/stretchr/testify` | Test assertions |

## Current Status

The project is in **M1 — Skill Manifest & Registry**. See `PLAN.md` for the full roadmap with milestones M0 through M6.

<!-- bv-agent-instructions-v1 -->

---

## Beads Workflow Integration

This project uses [beads_viewer](https://github.com/Dicklesworthstone/beads_viewer) for issue tracking. Issues are stored in `.beads/` and tracked in git.

### Essential Commands

```bash
# View issues (launches TUI - avoid in automated sessions)
bv

# CLI commands for agents (use these instead)
bd ready              # Show issues ready to work (no blockers)
bd list --status=open # All open issues
bd show <id>          # Full issue details with dependencies
bd create --title="..." --type=task --priority=2
bd update <id> --status=in_progress
bd close <id> --reason="Completed"
bd close <id1> <id2>  # Close multiple issues at once
bd sync               # Commit and push changes
```

### Workflow Pattern

1. **Start**: Run `bd ready` to find actionable work
2. **Claim**: Use `bd update <id> --status=in_progress`
3. **Work**: Implement the task
4. **Complete**: Use `bd close <id>`
5. **Sync**: Always run `bd sync` at session end

### Key Concepts

- **Dependencies**: Issues can block other issues. `bd ready` shows only unblocked work.
- **Priority**: P0=critical, P1=high, P2=medium, P3=low, P4=backlog (use numbers, not words)
- **Types**: task, bug, feature, epic, question, docs
- **Blocking**: `bd dep add <issue> <depends-on>` to add dependencies

### Session Protocol

**Before ending any session, run this checklist:**

```bash
git status              # Check what changed
git add <files>         # Stage code changes
bd sync                 # Commit beads changes
git commit -m "..."     # Commit code
bd sync                 # Commit any new beads changes
git push                # Push to remote
```

### Best Practices

- Check `bd ready` at session start to find available work
- Update status as you work (in_progress → closed)
- Create new issues with `bd create` when you discover tasks
- Use descriptive titles and set appropriate priority/type
- Always `bd sync` before ending session

<!-- end-bv-agent-instructions -->
