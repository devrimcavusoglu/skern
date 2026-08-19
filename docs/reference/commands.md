# Commands

## Overview

```
skern init                                    # Initialize .skern/ in current project
skern version                                 # Print version, commit, build date
skern completion [bash|zsh|fish]              # Generate shell completions

# Registry commands (manage skills in skern)
skern skill create <name>                     # Scaffold a new skill
skern skill import <url>                      # Import a skill from URL/git/gist
skern skill edit <name>                       # Edit skill metadata or body
skern skill remove <name>                     # Remove skill from registry
skern skill list [--scope user|project|all]   # List skills in registry
skern skill show <name>                       # Display skill details
skern skill search <query>                    # Search skills by name/description
skern skill validate <name>                   # Validate against Agent Skills spec
skern skill version <name> [--bump LEVEL]     # Show or bump a skill's version
skern skill diff <name> [name-b]              # Compare two skills (or registry vs platform)
skern skill recommend <query>                 # Reuse / extend / create suggestion

# Platform commands (deploy skills to platforms)
skern skill install <name>... --platform <p>  # Install one or more skills to platform
skern skill uninstall <name>... --platform <p># Remove one or more skills from platform
skern platform list                           # List detected platforms
skern platform status                         # Skill x platform installation matrix
```

`skern skill --help` groups subcommands into **Registry commands** (skills in skern) and **Platform commands** (skills deployed to a platform). The same split is reflected above.

## Global Flags

| Flag | Description |
|------|-------------|
| `--json` | Output in JSON format. Available on every command. |
| `--quiet` | Suppress non-essential output. |

Most commands also accept `--scope user|project|all`. Exit codes: `0` success, `1` error, `2` validation failure.

## `skern init`

Initialize the `.skern/` directory in the current project (creates the project-scoped skill registry). Idempotent — safe to re-run.

Optionally writes a skern usage snippet into agent instruction files (`AGENTS.md`, `CLAUDE.md`, `.claude/CLAUDE.md`) so the agent uses skern for all skill-related tasks instead of reading platform-native skill directories.

```sh
skern init                                    # creates .skern/ only (default)
skern init --instructions                     # also writes the snippet to discovered agent config files
skern init --instructions --tool-forming-loop # adds the search-before-create workflow section
skern init --target ./MY_AGENT.md             # write to a specific file (skips auto-discovery; repeatable)
skern init --print-instructions               # print the snippet to stdout instead of writing files
skern init --no-instructions                  # explicit opt-out: never writes, never prompts (installers, CI)
```

**Flags:**

| Flag | Description |
|------|-------------|
| `--instructions` | Write the skern usage snippet to discovered agent config files. Default: off. |
| `--no-instructions` | Explicit opt-out: do not write or offer the snippet, and never prompt. Mutually exclusive with `--instructions`, `--print-instructions`, `--target`, and `--tool-forming-loop` (exit 2 if combined, and nothing — not even `.skern/` — is created). |
| `--tool-forming-loop` | Include the tool-forming-loop section (search-before-create workflow). Default: off. |
| `--target <path>` | Explicit instruction file path. Repeatable. Disables auto-discovery when set. |
| `--print-instructions` | Print the rendered snippet to stdout instead of writing files. |

The instruction snippet is wrapped in `<!-- skern:instructions:start -->` / `<!-- skern:instructions:end -->` markers so re-running `skern init --instructions` updates the block in place rather than appending a duplicate.

**Interactivity contract.** `skern init` asks its two questions (write instructions? include tool-forming loop?) only when **all** of these hold: no instruction flag was given, stdin is a terminal, and `--json` is not set. Both default to **No**.

- Any instruction flag — `--instructions`, `--no-instructions`, `--print-instructions`, `--target`, `--tool-forming-loop` — disables **both** prompts; an unasked question keeps its default (so `--instructions` alone writes the snippet without the tool-forming section, and never waits on the second question).
- When stdin is **not** a terminal (installers, CI, piped or redirected input, `/dev/null`) or `--json` is set, skern never prompts and never blocks on input — both answers resolve to **No** and only flag values are honored. Terminal detection is a real isatty check, not a character-device test, so `< /dev/null` counts as non-interactive.

This is a documented guarantee, not an accident of the prompt's default. Automated callers should still pass `--no-instructions` (or `--instructions`) so their intent is explicit rather than inferred from stdin.

## `skern skill create`

Scaffold a new skill in the registry.

```sh
skern skill create <name> [flags]
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--description` | — | Skill description (start with "Use when…") |
| `--author` | — | Author name |
| `--author-type` | `human` | `human` or `agent` |
| `--author-platform` | — | Platform name (e.g. `claude-code`) — used with `agent` author type |
| `--tags` | — | Comma-separated list of tags. Each tag is lowercase alphanumeric segments joined by hyphens, optionally namespaced as `category:value` (single colon); anything else exits with code 2 |
| `--version` | `0.0.1` | Initial semver version |
| `--scope` | `user` | `user` or `project` |
| `--force` | `false` | Bypass overlap block |
| `--from-template <dir>` | — | Seed from a skill directory containing `SKILL.md` plus optional companion files |

Overlap detection runs automatically. Validation runs as warnings (does not block). Skill count thresholds emit warnings at >= 20 (project) / >= 50 (user).

### `--from-template`

`--from-template <dir>` accepts a **skill directory** — a directory containing a `SKILL.md` plus any optional companion files. Skern parses the template's frontmatter and recursively copies every sibling file and subdirectory (`references/`, `templates/`, `VENDORED.md`, …) into the new skill.

Anything else is rejected:

- `--from-template <file>` → "must point to a skill directory containing a SKILL.md file … pass the parent directory instead"
- `--from-template <dir-with-no-SKILL.md>` → "directory has no SKILL.md; a skill template must be a directory containing a SKILL.md file"

The new skill's `name` is always taken from the CLI argument. Other CLI flags (`--description`, `--tags`, `--author*`, `--version`) override template values when explicitly set; otherwise the template's values are preserved.

Frontmatter keys skern does not model (`compatibility`, `handoffs`, `metadata.phases`, …) are preserved verbatim in the new skill — see [Skill Format › Unrecognized keys pass through](/concepts/skill-format#unrecognized-keys-pass-through).

## `skern skill import`

Import a skill from a remote source into the registry.

```sh
skern skill import <url>
```

Supports GitHub repository directories and gists:

```sh
skern skill import https://github.com/owner/repo/tree/main/skills/code-review
skern skill import https://gist.github.com/<id>
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--scope` | `user` | `user` or `project` |
| `--name` | — | Override the skill name from the imported manifest |
| `--force` | `false` | Overwrite if the skill already exists and bypass overlap block |

The skill's directory contents (`SKILL.md` plus siblings discoverable via the source) are fetched and written into the registry. Overlap detection runs against existing skills.

## `skern skill edit`

Edit a skill's metadata fields, or open the body in `$EDITOR` (defaults to `vi`).

```sh
skern skill edit <name> [flags]
```

When called with field flags, the specified fields are updated directly. When called without field flags, the skill body is opened in your editor.

**Flags:**

| Flag | Description |
|------|-------------|
| `--scope` | `user` or `project` |
| `--description` | New description |
| `--author` | New author name |
| `--author-type` | `human` or `agent` |
| `--author-platform` | Platform name |
| `--version` | New version string |
| `--modified-by` | Name of modifier (appends to the `modified-by` list) |
| `--modified-by-type` | `human` or `agent` |
| `--modified-by-platform` | Platform name for the modifier |

`--modified-by` is **append-only** — every edit can record provenance without overwriting earlier entries.

## `skern skill remove`

Remove a skill from the registry. Does not affect platform installs.

```sh
skern skill remove <name> [--scope user|project]
```

To remove a skill from a platform without removing it from the registry, use `skern skill uninstall`.

## `skern skill list`

List skills in the registry.

```sh
skern skill list [--scope user|project|all] [flags]
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--scope` | `all` | `user`, `project`, or `all` |
| `--tag` | — | Filter results to skills with this flat tag (exact, case-insensitive) |
| `--category` | — | Filter by a namespaced `category:value` tag (repeatable) |
| `--include-untagged` | `false` | Treat a skill with no tag in a requested category as matching that category |
| `--with-platforms` | `false` | Include `installed_on` per skill (the detected platforms where the skill is installed at the same scope) |

### Categorical-tag filtering (`--category`)

`--category` narrows the list by structured `category:value` tags (e.g. `lang:python`, `topic:testing`). It is fully category-agnostic — the namespace is whatever precedes the first `:`; skern never enumerates known categories. Flat tags with no colon are not categorical and are matched by `--tag` instead.

- **Repeatable, with comma-lists:** `--category lang:python --category lang:go` and `--category lang:python,go` are equivalent.
- **OR within a category, AND across categories:** `--category lang:python,go --category topic:testing` matches skills tagged (`lang:python` **or** `lang:go`) **and** `topic:testing`.
- **Strict by default:** a skill that carries no tag in a requested category is excluded. Pass `--include-untagged` to treat a category-absent skill as applying to all values of that category. A category the skill *does* declare must still match a requested value even with `--include-untagged`.
- Matching is case-insensitive. `--tag` and `--category` compose with AND. Malformed input (`--category value` with no colon, an empty or comma-containing category name, or an empty value) exits with code 2.

Also runs pairwise overlap detection across all listed skills and appends a "Potential duplicates" section when matches are found (score >= 0.6). In `--json` mode they appear in the `duplicates` array.

Skills that cannot be parsed are reported as parse warnings rather than silently skipped — text mode prints `WARNING:` lines, `--json` mode populates the `parse_warnings` array.

When `--with-platforms` is set the JSON output contains an `installed_on` array on every skill — empty for skills not installed on any detected platform. Without the flag the field is omitted entirely so consumers can distinguish "queried, none" from "not queried".

## `skern skill show`

Display full details for a skill, including author provenance, modification history, and bundled files.

```sh
skern skill show <name> [--scope user|project]
```

## `skern skill search`

Search skills by name or description.

```sh
skern skill search <query> [--tag <tag>]
```

## `skern skill recommend`

Get an explicit recommendation: **REUSE** an existing skill, **EXTEND** one, or **CREATE** a new one.

```sh
skern skill recommend <query> [flags]
```

Thresholds: score ≥ 0.8 → REUSE; score ≥ 0.5 → EXTEND; below → CREATE.

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--name` | — | Agent-suggested skill name (used in CREATE recommendation) |
| `--threshold` | `0.3` | Minimum relevance score to include in results |
| `--scope` | `all` | `user`, `project`, or `all` |

## `skern skill validate`

Validate a skill against the Agent Skills spec.

```sh
skern skill validate <name> [--scope user|project]
```

Reports errors, warnings, and stylistic hints. Exit code is `2` when errors are present. See [Validation](/reference/validation) for the full list of checks.

## `skern skill version`

Show or bump a skill's semver version.

```sh
skern skill version <name>                       # show current version
skern skill version <name> --bump patch          # 0.0.1 -> 0.0.2
skern skill version <name> --bump minor          # 0.0.2 -> 0.1.0
skern skill version <name> --bump major          # 0.1.0 -> 1.0.0
```

**Flags:**

| Flag | Description |
|------|-------------|
| `--scope` | `user` or `project` |
| `--bump` | `patch`, `minor`, or `major` |

Without `--bump`, prints the current version. The JSON envelope returns `previous_version` and `bumped: true` after a bump.

## `skern skill diff`

Compare two skills side by side. With one argument, compares a registry skill against its installed copy on a platform. With two arguments, compares two registry skills.

```sh
# Registry vs installed
skern skill diff code-review --platform claude-code

# Registry vs registry
skern skill diff code-review code-review-strict --scope user
```

**Flags:**

| Flag | Description |
|------|-------------|
| `--scope` | `user` or `project` (omit to search both) |
| `--platform` | Platform to compare against — required when using one argument |

The output reports per-field diffs (description, version, author, …, plus any unmodeled frontmatter key that differs, reported under its own name) and a body diff flag. In `--json` mode, both bodies are returned in full.

## `skern skill install`

Install one or more skills to a single platform.

```sh
skern skill install <name>... --platform <platform>
skern skill install --tag <tag> --platform <platform>                 # install a tagged group
skern skill install --category lang:python --platform <platform>     # or a namespaced-tag group
```

Each invocation targets exactly one platform — `--platform all` is not accepted. Multiple skill names can be passed in one call. Each skill's outcome is reported in the `skills[]` array; a failure on one skill does not abort the batch — the command exits non-zero only when *every* install fails.

### Group installs (`--tag`, `--category`)

Instead of names, select a group with the same filters [`skill list`](#skern-skill-list) accepts: `--tag <tag>` (flat tag) and/or `--category <category:value>` (repeatable, comma-lists values, OR within a category, AND across categories; `--include-untagged` applies as in `list`). The filter resolves against the registry at `--scope`, so `--tag workflow --scope project` installs the project-registry skills tagged `workflow`. Resolved names are installed in sorted order and reported per-skill exactly as a name batch would be.

- Names and filters are **mutually exclusive** — passing both is a validation error (exit 2), as is passing neither.
- A filter that matches no registered skill is an **error** (exit 1: `no registered skills match --tag workflow in user scope`), never a silent no-op. Registry parse warnings (a skill directory whose `SKILL.md` could not be read — which may be exactly why a tag matched nothing) are included in that error, and printed to stderr when the filter does match.
- `--enforce-budget` counts the resolved group.

The response includes a top-level `capacity` block reporting the platform's installed-skill count after the operation, the threshold for that scope, and remaining headroom.

The whole skill directory is copied except paths the skill's frontmatter lists under `install.exclude` — see [Skill Format › Install-time exclusions](/concepts/skill-format#install-time-exclusions). The registry copy is never trimmed.

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--platform` | — *(required)* | One of: `claude-code`, `codex-cli`, `opencode`, `cursor`, `gemini-cli`, `github-copilot`, `windsurf`, `continue` |
| `--scope` | `user` | `user` or `project` |
| `--force` | `false` | Overwrite existing installation |
| `--enforce-budget` | `false` | Refuse the operation if it would push the platform's installed-skill count past the per-scope threshold |
| `--tag` | — | Select registry skills carrying this tag instead of naming them. Mutually exclusive with names. |
| `--category` | — | Select by namespaced tag `category:value`; repeatable, comma-lists values. Mutually exclusive with names. |
| `--include-untagged` | `false` | With `--category`: treat a skill with no tag in a requested category as matching it. |

## `skern skill uninstall`

Remove one or more skills from a platform. Mirrors `install` semantics: one platform per call, multiple skills allowed, partial failures reported per-skill, post-op `capacity` block in the response. The registry copy is **not** affected — use `skern skill remove` for that.

```sh
skern skill uninstall <name>... --platform <platform>
skern skill uninstall --tag <tag> --platform <platform>     # evict a tagged group
```

`--tag` / `--category` select a group the same way as on `install`: the filter resolves against the registry at `--scope`, then is narrowed to the skills actually installed on the platform. Tagged-but-not-installed skills are skipped, not reported as failures. If nothing in the group is installed the command errors (`no installed skills match --tag workflow on claude-code (user scope)`); if the tag matches nothing in the registry, the error says so instead. Names and filters are mutually exclusive.

Because the registry defines the group, a skill that was already removed from the registry (`skern skill remove`) is no longer reachable by tag — uninstall it by name. Retire a group by uninstalling it from platforms *before* removing it from the registry.

**Flags:**

| Flag | Description |
|------|-------------|
| `--platform` | Required. Same enumeration as `install`. |
| `--scope` | `user` or `project` |
| `--tag` / `--category` / `--include-untagged` | Group selection, as on `install`. |

## `skern platform list`

List all detected platforms, the directories they install to, and whether each was detected on the current host.

```sh
skern platform list
```

Detection is per-platform: each adapter checks its own user-level config dir (`~/.claude`, `~/.cursor`, `~/.gemini`, `~/.copilot`, `~/.codex`, …) so platforms that share `.agents/skills/` as a project directory are still distinguished correctly.

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

Print version, commit, and build date. Falls back to `runtime/debug.ReadBuildInfo` when ldflags aren't injected (e.g. binaries built via `go install`).

```sh
skern version
```
