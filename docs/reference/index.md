# CLI Reference

Skern provides a set of commands for managing agent skills across platforms. Every command supports `--json` for machine-readable output and `--quiet` for silent operation.

## Pages

- [**Commands**](/reference/commands) — full reference for every command including flags and usage examples.
- [**Validation**](/reference/validation) — rules enforced by `skern skill validate` and during skill creation.
- [**Overlap Detection**](/reference/overlap-detection) — fuzzy matching, similarity scoring, and capacity thresholds.

## Command Groups

`skern skill --help` splits subcommands into two groups:

- **Registry commands** — manage skills in skern itself (`create`, `import`, `edit`, `remove`, `list`, `show`, `search`, `validate`, `version`, `diff`, `recommend`).
- **Platform commands** — deploy registered skills to a platform (`install`, `uninstall`).

The registry holds the canonical skill. Install copies the skill into a platform's directory; uninstall removes the platform copy without touching the registry; remove deletes the skill from the registry entirely.

## Global Flags

| Flag | Description |
|------|-------------|
| `--json` | Output in JSON format |
| `--quiet` | Suppress non-essential output |
| `--scope user\|project\|all` | Target user-level or project-level registry (most commands) |

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Error |
| 2 | Validation failure (`skern skill validate` errors, name validation failures, etc.) |
