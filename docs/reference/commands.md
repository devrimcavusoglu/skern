# Commands

## Overview

```
skern init                                    # Initialize .skern/ in current project
skern skill create <name>                     # Scaffold a new skill
skern skill search <query>                    # Search skills by name/description
skern skill recommend <query>                 # Recommend: reuse, extend, or create
skern skill list [--scope user|project|all]   # List skills in registry
skern skill show <name>                       # Display skill details
skern skill validate <name>                   # Validate against Agent Skills spec
skern skill remove <name>                     # Remove skill from registry
skern skill install <name> --platform <p>     # Install skill to platform
skern skill uninstall <name> --platform <p>   # Remove skill from platform
skern platform list                           # List detected platforms
skern platform status                         # Skill x platform installation matrix
skern completion [bash|zsh|fish]              # Generate shell completions
skern version                                 # Print version info
```

## `skern init`

Initialize the `.skern/` directory in the current project. This creates the project-scoped skill registry.

```sh
skern init
```

## `skern skill create`

Scaffold a new `SKILL.md` file in the registry.

```sh
skern skill create <name> [flags]
```

**Flags:**

| Flag | Description |
|------|-------------|
| `--author` | Author name |
| `--author-type` | `human` or `agent` |
| `--author-platform` | Platform name (e.g., `claude-code`) |
| `--description` | Skill description |
| `--force` | Bypass overlap block |
| `--from-template <path>` | Use file as skill body |

Overlap detection runs automatically during creation. See [Overlap Detection](/reference/overlap-detection) for details.

## `skern skill search`

Search skills by name or description.

```sh
skern skill search <query>
```

## `skern skill recommend`

Get recommendations on whether to reuse, extend, or create a skill.

```sh
skern skill recommend <query> [flags]
```

**Flags:**

| Flag | Description |
|------|-------------|
| `--name` | Agent-suggested skill name |
| `--threshold` | Minimum relevance score (default: 0.3) |
| `--scope` | `user`, `project`, or `all` |

## `skern skill list`

List all skills in the registry.

```sh
skern skill list [--scope user|project|all]
```

Also runs pairwise overlap detection across all listed skills and appends a "Potential duplicates" section when matches are found (score >= 0.6). In `--json` mode, these appear in the `duplicates` array.

## `skern skill show`

Display full details for a skill, including author provenance and modification history.

```sh
skern skill show <name>
```

## `skern skill validate`

Validate a skill against the Agent Skills spec.

```sh
skern skill validate <name>
```

See [Validation](/reference/validation) for the full list of checks.

## `skern skill remove`

Remove a skill from the registry.

```sh
skern skill remove <name>
```

## `skern skill install`

Install a skill to one or more platforms.

```sh
skern skill install <name> --platform <platform>
```

**Flags:**

| Flag | Description |
|------|-------------|
| `--platform` | `claude-code`, `codex-cli`, `opencode`, or `all` (required) |
| `--scope` | `user` or `project` |

## `skern skill uninstall`

Remove a skill from a platform.

```sh
skern skill uninstall <name> --platform <platform>
```

**Flags:**

| Flag | Description |
|------|-------------|
| `--platform` | `claude-code`, `codex-cli`, `opencode`, or `all` (required) |
| `--scope` | `user` or `project` |

## `skern platform list`

List all detected platforms.

```sh
skern platform list
```

## `skern platform status`

Show a matrix of skills and their installation status across platforms.

```sh
skern platform status [--scope user|project]
```

## `skern completion`

Generate shell completion scripts.

```sh
skern completion bash
skern completion zsh
skern completion fish
```

## `skern version`

Print version, commit, and build date.

```sh
skern version
```
