# AGENTS.md — Skern Development Guide

Conventions, architecture, and design decisions for agents working in this repo.
Human contributor setup — the full build/test matrix, directory tree, and
dependency list — lives in [docs/contributing/development.md](./docs/contributing/development.md).
Release history lives in [CHANGELOG.md](./CHANGELOG.md); neither is restated here.

## Project Overview

Skern is a minimal, agent-first CLI for managing Agent Skills across agentic
development platforms. It follows the Agent Skills open standard
([agentskills.io](https://agentskills.io)) and uses `SKILL.md` files with YAML
frontmatter as the canonical format.

Written in **Go 1.25+**. Latest release **v0.4.0**; `main` carries unreleased
work (see the `[Unreleased]` section of the CHANGELOG).

### Scope

**The registry is the center of gravity.** Skern's value is being the single
cross-platform source of truth for skills: one `SKILL.md` per skill, stored
once, managed through skern regardless of which agent is running. Registry
commands (`create`, `import`, `edit`, `list`, `search`, `show`, `validate`,
`version`, `diff`, `recommend`, `remove`) are the primary surface, and
`skern init --instructions` exists so an agent can be told to reach for skern
directly instead of hunting through platform directories.

**Platform install/uninstall is first-class and stays.** Copying skills out to
native platform directories serves real cases the registry alone does not:
native progressive disclosure, sandboxed agents that cannot shell out to
skern, and one-shot migration from an existing `.claude/skills/`. Adapters are
maintained and extended, and adapter bugs are ordinary product bugs — they are
just not the headline drive. The `skern skill` command tree encodes this split
directly via two cobra groups: *Registry commands* and *Platform commands*
(`internal/cli/skill.go`).

When proposing features, weigh them against the registry thesis first. New
work that only makes sense as a per-platform copy needs a use case from the
list above.

**Deliberately out of scope.** Remote skill discovery — a hosted catalog,
community index, or remote search inside `skern skill search` — was considered
and set aside ([#19], [#50], closed April 2026) in favor of existing
ecosystems such as [skild](https://github.com/Peiiii/skild) and skills.sh.
Skern's job there is *import*, not hosting; widening import sources is
tracked as [#79]. An MCP server mode, sandboxed execution backends, and
usage-based eviction have likewise been floated and never tracked; treat them
as unplanned unless an issue appears.

## Package Responsibilities

| Package | Responsibility |
|---|---|
| `cli/` | Command wiring, flag parsing, output. **No business logic.** |
| `cli/instructions/` | Embedded snippets rendered by `skern init --instructions` |
| `skill/` | Domain types and operations: `Skill`, `Author`, `ModifiedByEntry`, manifest parse/serialize, validation, scaffolding, versioning, import |
| `registry/` | Filesystem CRUD over `~/.skern/skills/` and `.skern/skills/` |
| `platform/` | Declarative `Spec` table (`spec.go`) + one generic `Adapter` (`adapter.go`) implementing the `Platform` interface |
| `overlap/` | Similarity scoring (Levenshtein, keyword overlap); returns `float64` in [0, 1] |
| `output/` | `--json` / `--quiet` formatting. All commands print through this package. |

`cli/` uses injectable `NewRegistry` / `NewDetector` on `CommandContext` for
test isolation.

## Design Decisions

1. **`SKILL.md` is the canonical format.** Skern does not invent a
   `skill.yaml`. A skill is a directory containing a `SKILL.md` plus optional
   supporting files.

2. **The registry is a filesystem directory.** `~/.skern/skills/` for user
   scope, `.skern/skills/` for project scope. No database, no daemon, no lock
   files.

3. **Adapters are declarative, and may carry behavior beyond copying.**
   Every platform is one row in `Specs` (`internal/platform/spec.go`) driving a
   single generic `Adapter` — adding a platform means a `Spec` row plus a
   `Type` constant, never a per-platform Go file. Installing is a directory
   copy filtered by the skill's own `install.exclude` ([#103]) *today*, but
   pure-copy is **not** an invariant: a `Spec` may grow declarative
   post-install behavior (companion files, per-platform destination
   overrides) where a platform's loader demands it. Any such
   behavior belongs in the spec table as data, not as a hand-written
   per-platform code path, and `Uninstall` must reverse whatever `Install`
   produced. Open work depending on this: [#99], [#101], [#47].

4. **Platform detection is per-platform, not per-directory.** Skern checks
   each spec's home-relative `DetectHome` paths (`~/.claude/`, `~/.cursor/`,
   `~/.gemini/`, …). This matters because several platforms share
   `.agents/skills/` as their project directory, so the directory's existence
   cannot identify the agent. Consequence: a skill installed to a shared
   directory is visible to every platform reading it, and capacity counts
   protect the *directory*, not the logical agent.

5. **One platform per invocation.** `install`/`uninstall` target exactly one
   platform — there is no `--platform all`, and skern never broadcasts across
   platforms. Agents state the platform they run on. Multiple *skill names*
   per call are supported for batch operations.

6. **JSON output is first-class.** Every command supports `--json`. Default
   output is human-readable text. Exit codes are semantic: `0` success, `1`
   error, `2` validation failure.

## Tool-Forming Loop

Skern's core differentiator: agents don't just *use* skills, they *create*
them when a recurring need appears.

```
Agent identifies a recurring need
  --> skern skill search <query>
  --> no results (or low similarity)
  --> skern skill create <name>
  --> Agent implements the skill
  --> Skill becomes reusable
```

On later encounters the agent finds the existing skill via
`skern skill search` and reuses it.

### Guardrails

| Guardrail | Mechanism | Default |
|---|---|---|
| Overlap warning threshold | Similarity score 0.0–1.0 | Warn at >= 0.6 |
| Overlap block threshold | Similarity score | Block at >= 0.9, require `--force` |
| Skill count warning (project) | Count in `.skern/skills/` | Warn at >= 20 |
| Skill count warning (user) | Count in `~/.skern/skills/` | Warn at >= 50 |
| Per-platform capacity (project scope) | Count in the platform's project skills dir (e.g. `.claude/skills/`) | Warn at >= 20 |
| Per-platform capacity (user scope) | Count in the platform's user skills dir (e.g. `~/.claude/skills/`) | Warn at >= 50 |
| Capacity enforcement (install) | `--enforce-budget` | Off by default; refuses install when the threshold would be exceeded |
| Deduplication hints | On `skern skill list` | Flags potential duplicates |

Thresholds are defined in `internal/skill/capacity.go`; overlap thresholds in
`internal/overlap/detector.go`. Every `install`/`uninstall` JSON response
carries a `capacity` block (count, threshold, headroom, over-budget flag) so
agents can react to capacity pressure without a second query.

## SKILL.md Format

```markdown
---
name: skill-name
description: |
  Use when <triggering conditions for this skill>.
tags: [code-review, lang:python]   # optional
allowed-tools: []                  # optional; omitted when empty
metadata:
  author:
    name: author-name
    type: human           # human | agent
    platform: claude-code # only when type=agent
  version: "0.0.1"
  modified-by:            # append-only provenance list
    - name: codex-cli
      type: agent
      platform: codex-cli
      date: "2025-07-15T10:30:00Z"
install:                  # optional; shapes the platform copy only
  exclude: [eval, fixtures/*]   # kept in the registry, not installed
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

Required fields: `name`, `description`. The directory name must match `name`.
Optional fields are omitted from output when empty rather than emitted as
empty values.

### Validation Rules

**Names** must match `^[a-z0-9]+([.-][a-z0-9]+)*$`, 1–64 characters. Dots and
hyphens are both valid segment separators; dots enable namespace-style names
(e.g. `myorg.bootstrap`).

**Tags** are either flat (`code-review`) or categorical (`lang:python`):
lowercase alphanumeric segments joined by hyphens, with at most one colon
separating category from value. Uppercase is rejected on write so stored tags
have one canonical form; tag *filters* stay case-insensitive so legacy
hand-edited tags still match. `skill list`, `skill install`, and
`skill uninstall` filter on them via `--tag` (flat) and `--category`
(namespaced, repeatable, OR within a category, AND across categories) — one
`skillFilter` in `internal/cli/skill_helpers.go` defines the flags and match
semantics for all three; on install/uninstall the filter is mutually
exclusive with positional names.

**`install.exclude`** ([#103]) lists `path.Match` globs, relative to the skill
directory, that `skill install` leaves out of the platform copy (the registry
keeps everything). A pattern matches a path or any leading directory of it,
so a bare directory name excludes its subtree; `**` is rejected; `SKILL.md`
is never excluded (so `*` is legal). Matching lives in `skill.MatchExclude`
(`internal/skill/folder.go`) and is applied by `platform.copyDir` via
`InstallOptions.Exclude`; `skill validate` errors on malformed patterns and
warns on no-match / excluded-but-referenced files; `skill install` refuses a
skill whose patterns fail validation; `skill diff` reports a changed list
under `install.exclude`. Other keys under `install:` pass through via
`InstallConfig.Extra`.

**Unmodeled keys pass through.** The fields above are the keys skern models;
any other key — top-level, `metadata.*`, or nested in `metadata.author` /
a `modified-by` entry — lands in the matching `Extra` inline map
(`internal/skill/skill.go`) on parse and is written back by `WriteManifest`,
so `create --from-template`, `import`, `edit`, and `version` never drop a
consumer's extended contract ([#100]). Values are preserved as YAML 1.2 parses
them; formatting is normalized (modeled keys first, extras sorted, 1.1-only
scalars like bare dates and `yes`/`no` canonicalized). An `Extra` key that
shadows a modeled key is a `WriteManifest` error, never a silent overwrite —
yaml.v3 would otherwise panic; the modeled-key sets are derived from the
struct tags (`yamlKeys`), so a new modeled field is guarded automatically.

### Writing Skills Guidelines

Guidelines for authoring skill content, adapted from
[superpowers/writing-skills](https://github.com/obra/superpowers/tree/main/skills/writing-skills),
are enforced in two places — `internal/skill/scaffold.go` (the scaffolded
template) and `internal/skill/validator.go` (stylistic hints on
`skill validate`). Full details: [docs/writing-skills.md](./docs/writing-skills.md).

Key points:

- **Description**: start with "Use when..." — triggering conditions, not a
  workflow summary
- **Naming**: `kebab-case`, verb-first active voice (`creating-skills`, not
  `skill-creation`)
- **Body structure**: Overview → When to Use → Core Pattern → Quick Reference
  → Common Mistakes
- **Token budget**: getting-started < 150 words; frequently-loaded < 200
  words; others < 500 words
- **Examples**: one excellent example beats many mediocre ones

## Paths

### Registry

| Scope | Path |
|---|---|
| User | `~/.skern/skills/` |
| Project | `.skern/skills/` |

### Platforms

Generated from `internal/platform/spec.go` — append a row there to add a
platform. Paths follow
[vercel-labs/skills](https://github.com/vercel-labs/skills#supported-agents)
where possible.

| Adapter name | User-level | Project-level |
|---|---|---|
| `claude-code` | `~/.claude/skills/<name>/` | `.claude/skills/<name>/` |
| `codex-cli` | `~/.agents/skills/<name>/` | `.agents/skills/<name>/` |
| `opencode` | `~/.config/opencode/skills/<name>/` | `.opencode/skills/<name>/` |
| `cursor` | `~/.cursor/skills/<name>/` | `.agents/skills/<name>/` |
| `gemini-cli` | `~/.gemini/skills/<name>/` | `.agents/skills/<name>/` |
| `github-copilot` | `~/.copilot/skills/<name>/` | `.agents/skills/<name>/` |
| `windsurf` | `~/.codeium/windsurf/skills/<name>/` | `.windsurf/skills/<name>/` |
| `continue` | `~/.continue/skills/<name>/` | `.continue/skills/<name>/` |

`codex-cli` keeps `~/.agents/skills/` rather than vercel's `~/.codex/skills/`
so existing skern users see no disk-layout change. GitHub Copilot accepts
`.github/skills`, `.claude/skills`, **and** `.agents/skills` for project
scope, and `~/.copilot/skills` or `~/.agents/skills` for personal scope — the
values above are valid Copilot paths, not a skern-specific convention
([GitHub docs](https://docs.github.com/en/copilot/concepts/agents/about-agent-skills)).

## Code Conventions

### Go Style

- Standard Go idioms and `gofmt` formatting
- Exported names `CamelCase`; unexported `camelCase`
- Prefer stdlib over third-party unless there is a strong reason
- Keep packages small and single-responsibility
- Use `internal/` to prevent external imports of implementation details

### Error Handling

- Return `error` values; do not panic
- Wrap with `fmt.Errorf("context: %w", err)`
- Error messages include actionable suggestions
- Semantic exit codes: `0` success, `1` error, `2` validation failure

### CLI Output

- Every command supports `--json`
- Default output is human-friendly text
- `--quiet` suppresses non-essential output
- All printing goes through `internal/output`

### Testing

- Table-driven tests with descriptive subtest names
- `testify/assert` and `testify/require` for assertions
- Filesystem-touching tests must use `t.TempDir()`
- `*_test.go` in the same package

### Essential Commands

```sh
make build   # build with version/commit/date injected
make test    # go test ./...
make lint    # golangci-lint run
make fmt     # gofmt -w .
```

The full matrix — coverage, installer tests, smoke tests, and the manual
agent-driven harness — is documented in
[docs/contributing/development.md](./docs/contributing/development.md).

### Git Workflow

Work is tracked as GitHub issues via the `gh` CLI and merged to `main` by pull
request.

- **Branch naming**: `<type>/<slug>` where type is `feature`, `fix`, `chore`,
  `docs`, or `release` — e.g. `feature/category-tag-filter`,
  `fix/release-duplicate-trigger`, `release/v0.3.1`. (Milestone-numbered
  branches like `feature/m7-…` are historical; milestones M0–M7 are closed.)
- **Commits**: concise, imperative subject ("Add manifest parser", not
  "Added manifest parser"), referencing issues as `#<number>`
- **User-facing changes** land a `CHANGELOG.md` `[Unreleased]` entry in the
  same PR

## Roadmap

Everything planned is a tracked issue; this list is a map, not a commitment.

**Adapter model** — all three need the declarative-hook mechanism from design
decision 3; settle the mechanism once rather than special-casing each.

- [#99] — OpenCode's loader rejects dot-named skills; bridge via command shims
- [#101] — four platforms share `.agents/skills/`, blocking per-platform content
- [#47] — platform-specific skill variants

**Registry and workflow**

- [#48] — skill sync: bulk reconcile registry with platforms
- [#51] — dry run / preview mode for mutating commands
- [#45] — skill dependencies and composition
- [#79] — import sources beyond GitHub (GitLab, Bitbucket, skills.sh, skild)
- [#88] — ship `writing-skills` as a built-in skill
- [#75] — skill stats for context optimization

[#19]: https://github.com/devrimcavusoglu/skern/issues/19
[#45]: https://github.com/devrimcavusoglu/skern/issues/45
[#47]: https://github.com/devrimcavusoglu/skern/issues/47
[#48]: https://github.com/devrimcavusoglu/skern/issues/48
[#50]: https://github.com/devrimcavusoglu/skern/issues/50
[#51]: https://github.com/devrimcavusoglu/skern/issues/51
[#75]: https://github.com/devrimcavusoglu/skern/issues/75
[#79]: https://github.com/devrimcavusoglu/skern/issues/79
[#88]: https://github.com/devrimcavusoglu/skern/issues/88
[#99]: https://github.com/devrimcavusoglu/skern/issues/99
[#100]: https://github.com/devrimcavusoglu/skern/issues/100
[#101]: https://github.com/devrimcavusoglu/skern/issues/101
[#102]: https://github.com/devrimcavusoglu/skern/issues/102
[#103]: https://github.com/devrimcavusoglu/skern/issues/103
[#104]: https://github.com/devrimcavusoglu/skern/issues/104
