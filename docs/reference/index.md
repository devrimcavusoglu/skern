# CLI Reference

Skern provides a set of commands for managing agent skills across platforms. All commands support `--json` for machine-readable output and `--quiet` for silent operation.

## Command Groups

### [Commands](/reference/commands)

Full reference for all CLI commands including flags and usage examples.

### [Validation](/reference/validation)

Validation rules enforced by `skern skill validate` and during skill creation.

### [Overlap Detection](/reference/overlap-detection)

How fuzzy matching and similarity scoring prevent skill duplication.

## Global Flags

| Flag | Description |
|------|-------------|
| `--json` | Output in JSON format |
| `--quiet` | Suppress non-error output |
| `--scope user\|project` | Target user-level or project-level registry |

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Error |
| 2 | Validation failure |
