# Development

This page covers contributor setup: building, testing, linting, and where files
live. Coding conventions, architecture, and design decisions are documented in
[AGENTS.md](https://github.com/devrimcavusoglu/skern/blob/main/AGENTS.md) in the
repo root — that file is canonical for both agents and humans, and is not
restated here.

## Build

```sh
make build        # Build binary with version/commit/date injected via ldflags
```

The binary is placed at the repository root as `skern`.

## Test

```sh
make test                  # Run unit tests
make test-v                # Verbose test output
make test-cover            # Generate coverage report (coverage.out + coverage.html)
make test-install          # Run installer script tests
make test-smoke            # Build binary and run smoke tests against it
make test-manual-setup     # Set up manual test scenarios under /tmp
make test-manual-report    # Generate the manual test pass/fail checklist
make test-manual-teardown  # Clean up manual test environment
```

Tests use stdlib `testing` plus `testify`. Integration tests that touch the filesystem must use `t.TempDir()`. Table-driven test patterns are preferred.

## Lint & Format

```sh
make lint   # golangci-lint run
make fmt    # gofmt -w .
```

Linter config lives in `.golangci.yml`.

## Clean

```sh
make clean  # Remove binary and coverage files
```

## Project Structure

```
cmd/skern/main.go              # Entry point
internal/
  cli/                         # Cobra commands (root, version, init, completion, skill_*, platform_*)
    instructions/              # Embedded snippets used by `skern init --instructions`
  skill/                       # Domain: Skill struct, manifest parse/write, validation, scaffolding, versioning, importer
  overlap/                     # Fuzzy name matching (Levenshtein) + description similarity scoring
  registry/                    # Filesystem CRUD over ~/.skern/skills/ and .skern/skills/
  platform/                    # Declarative Spec table (spec.go) + a single generic Adapter (adapter.go)
  output/                      # JSON/text output formatting (--json, --quiet)
scripts/
  install.sh                   # Unix installer
  install.ps1                  # Windows PowerShell installer
  install_test.sh              # Installer tests
  smoke_test.sh                # Smoke tests for built binary
tests/manual/                  # Agent-driven test scenarios (11 scenarios)
docs/                          # VitePress documentation site
```

## Conventions

Coding conventions — package responsibilities, Go style, error handling, CLI
output, and testing patterns — are documented in
[AGENTS.md](https://github.com/devrimcavusoglu/skern/blob/main/AGENTS.md#code-conventions).
Two that shape where you'll be editing:

- `cli/` is wiring, flag parsing, and output. Business logic lives in `skill/`, `registry/`, `platform/`, `overlap/`.
- Adding a new platform = one row in `internal/platform/spec.go` plus one `Type` constant. No per-platform Go file.

## Dependencies

| Dependency | Purpose |
|------------|---------|
| `github.com/spf13/cobra` | CLI framework |
| `gopkg.in/yaml.v3` | YAML frontmatter parsing |
| `github.com/stretchr/testify` | Test assertions |

## Issue Tracking & Branching

Work is issue-driven: branch from `main`, merge back by pull request. Branch
naming, commit style, and the CHANGELOG rule are defined once in
[AGENTS.md › Git Workflow](https://github.com/devrimcavusoglu/skern/blob/main/AGENTS.md#git-workflow).

```sh
gh issue list                              # List open issues
gh issue create --title "Title" --body ""  # Create a new issue
gh issue view <number>                     # View issue details
```
