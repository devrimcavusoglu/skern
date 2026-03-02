# Development

## Build

```sh
make build        # Build binary with version/commit/date injected
```

The binary is placed at the repository root as `skern`.

## Test

```sh
make test         # Run unit tests
make test-v       # Verbose test output
make test-cover   # Generate coverage report
make test-smoke   # Smoke & E2E tests against built binary
```

## Lint & Format

```sh
make lint         # Run golangci-lint
make fmt          # Format code with gofmt
```

## Clean

```sh
make clean        # Remove build artifacts
```

## Project Structure

```
cmd/skern/main.go            # Entry point
internal/
  cli/                       # Cobra commands (root, version, init, completion, skill_*, platform_*)
  skill/                     # Domain: Skill struct, manifest parse/write, validation, scaffolding
  overlap/                   # Fuzzy name matching (Levenshtein) + description similarity scoring
  registry/                  # Filesystem CRUD over ~/.skern/skills/ and .skern/skills/
  platform/                  # Adapters: Claude Code, Codex CLI, OpenCode
  output/                    # JSON/text output formatting (--json, --quiet)
```

## Conventions

- Tests are table-driven with `testify` and use `t.TempDir()` for temp dirs
- `cli/` package uses injectable `newRegistryFunc` / `newDetectorFunc` for test isolation
- Errors are wrapped with `fmt.Errorf("context: %w", err)`
- Dependencies: `cobra`, `yaml.v3`, `testify`
