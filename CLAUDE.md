# CLAUDE.md — Scribe

Scribe is a minimal, agent-first CLI for managing Agent Skills across Claude Code, Codex CLI, and OpenCode. It follows the Agent Skills open standard (`SKILL.md` with YAML frontmatter).

## Build & Test

```sh
make build          # Build binary with version/commit/date injected
make test           # go test ./...
make test-v         # Verbose test output
make test-cover     # Generate coverage report
make lint           # golangci-lint run
make fmt            # gofmt -w .
```

Requires Go 1.23+. Dependencies: `cobra`, `yaml.v3`, `testify`.

## Project Structure

```
cmd/scribe/main.go            # Entry point
internal/
  cli/                         # Cobra commands (root, version, init, completion, skill_*, platform_*)
  skill/                       # Domain: Skill struct, manifest parse/write, validation, scaffolding
  overlap/                     # Fuzzy name matching (Levenshtein) + description similarity scoring
  registry/                    # Filesystem CRUD over ~/.scribe/skills/ and .scribe/skills/
  platform/                    # Adapters: Claude Code, Codex CLI, OpenCode
  output/                      # JSON/text output formatting (--json, --quiet)
```

## Key Commands

```
scribe init                                    # Initialize .scribe/ in project
scribe skill create <name> [flags]             # Scaffold SKILL.md
scribe skill list [--scope user|project|all]   # List skills (with dedup hints)
scribe skill show <name>                       # Show skill details
scribe skill search <query>                    # Search by name
scribe skill validate <name>                   # Validate against spec
scribe skill remove <name>                     # Remove from registry
scribe skill install <name> --platform <p>     # Install to platform
scribe skill uninstall <name> --platform <p>   # Uninstall from platform
scribe platform list                           # Show detected platforms
scribe platform status                         # Skill x platform matrix
scribe completion [bash|zsh|fish]              # Shell completions
```

## Conventions

- All commands support `--json` for machine-readable output
- Exit codes: 0=success, 1=error, 2=validation failure
- Skill names: `^[a-z0-9]+(-[a-z0-9]+)*$`, 1-64 chars
- Tests: table-driven with `testify`, temp dirs via `t.TempDir()`
- `cli/` package uses injectable `newRegistryFunc` / `newDetectorFunc` for test isolation
- Errors wrapped with `fmt.Errorf("context: %w", err)`
- Overlap thresholds: <0.6 proceed, >=0.6 warn, >=0.9 block (override with `--force`)

## Platform Paths

| Platform    | User-level                          | Project-level             |
|-------------|-------------------------------------|---------------------------|
| Claude Code | `~/.claude/skills/<name>/`          | `.claude/skills/<name>/`  |
| Codex CLI   | `~/.agents/skills/<name>/`          | `.agents/skills/<name>/`  |
| OpenCode    | `~/.config/opencode/skills/<name>/` | `.opencode/skills/<name>/`|

## Issue Tracking

Uses `br` (beads-rust). Milestones map to epics. Reference issues in commits as `br#<id>`.

```sh
br list               # Open issues
br create "Title"     # New issue
br update <id> --status in_progress
br close <id>
br sync --flush-only  # Export JSONL (then git add .beads/)
```
