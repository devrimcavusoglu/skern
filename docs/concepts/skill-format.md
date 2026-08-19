# Skill Format

Skills follow the [Agent Skills](https://agentskills.io) open standard. Each skill is a directory containing a `SKILL.md` file with YAML frontmatter and a markdown body, plus any optional companion files.

## Structure

```markdown
---
name: code-review
description: |
  Use when reviewing pull requests for style and correctness.
tags:
  - review
  - quality
allowed-tools:
  - Read
  - Grep
  - Glob
metadata:
  author:
    name: Jane Doe
    type: human            # human | agent
    platform: claude-code  # only when type=agent
  version: "1.0.0"
  modified-by:             # append-only provenance
    - name: codex-cli
      type: agent
      platform: codex-cli
      date: "2026-04-12T10:30:00Z"
---

## Overview

1-2 sentence core principle of the skill.

## When to Use

- Triggering conditions and symptoms

## Core Pattern

The main technique or pattern (before/after for techniques).
```

## Frontmatter Fields

| Field | Required | Description |
|-------|----------|-------------|
| `name` | Yes | Skill name matching `[a-z0-9]+([.-][a-z0-9]+)*`, 1-64 chars. Hyphens and dots are both valid separators (`code-review`, `myorg.bootstrap`). Must equal the directory name. |
| `description` | Yes | What the skill does — start with "Use when…". Max 1024 chars. |
| `tags` | No | List of classification tags. Lowercase alphanumeric segments joined by hyphens (`code-review`), optionally namespaced as `category:value` with a single colon (`lang:python`, `topic:ci-cd`). Filter matching is case-insensitive. |
| `allowed-tools` | No | Tools the skill may use. No empty entries. |
| `metadata.author.name` | No | Author name |
| `metadata.author.type` | No | `human` or `agent` |
| `metadata.author.platform` | No | Platform name (e.g. `claude-code`) — used when `type: agent` |
| `metadata.version` | No | Semantic version (e.g. `1.0.0`); defaults to `0.0.1` on `skill create` |
| `metadata.modified-by` | No | Append-only modification history (set via `skern skill edit --modified-by`) |
| `install.exclude` | No | Glob patterns (relative to the skill directory) for files and directories that stay in the registry but are **not** copied on `skern skill install`. See [Install-time exclusions](#install-time-exclusions). |

`name` and `description` are the only hard requirements. The rest help discovery, validation, and provenance.

### Unrecognized keys pass through

The table above is the set of keys skern *models*. Every other frontmatter key — top-level (`license`, `compatibility`, `handoffs`, …), under `metadata` (`metadata.phases`, a `metadata.tags` list distinct from top-level `tags`, …), or nested inside `metadata.author` / a `metadata.modified-by` entry (`author.email`, `reason`) — is preserved whenever skern rewrites a `SKILL.md`: `skill create --from-template`, `skill import`, `skill edit`, `skill version`. Extended skill contracts survive the round trip, and since platform installs copy the registry file byte-for-byte, the keys reach the agent too.

What is preserved is the *data as YAML 1.2 parses it*: keys and values. Formatting is normalized on write — skern's own keys come first in a fixed order, pass-through keys follow sorted by name (nested mapping keys are sorted too), comments and quoting style are dropped, and YAML 1.1-only scalars are canonicalized: an unquoted date-like scalar (`created: 2024-01-15`) is re-emitted in RFC 3339 form, and bare `yes`/`no`/`on`/`off` are re-emitted quoted so they stay strings. `skern skill diff` reports pass-through keys that differ between two skills under their own names (`compatibility`, `metadata.phases`, `author.email`); an explicit `key: null` is reported as `null`, distinct from an absent key. Pass-through keys are not shown by `skill show` yet.

## Body

The markdown body contains the skill's instructions. This is what the agent reads when the skill is activated. It must be non-empty.

The recommended body structure (also what `skern skill create` scaffolds) is:

1. **Overview** — 1-2 sentence core principle
2. **When to Use** — triggering conditions
3. **Core Pattern** — the main technique
4. **Quick Reference** — scannable summary
5. **Common Mistakes** — frequent errors and fixes

See [Writing Skills](/writing-skills) for details on what makes a discoverable, well-structured skill.

## Author Provenance

Skills track an author and an append-only `modified-by` history. `skern skill show` prints the full provenance chain when present, including editor name, type (human/agent), platform, and date. `skern skill edit --modified-by <name>` adds a new entry without overwriting earlier ones.

## Folder Structure

A skill is a directory. `SKILL.md` is the manifest; sibling files travel with it.

```
my-skill/
├── SKILL.md
├── references/
│   └── notes.md
├── scripts/
│   ├── convert.py
│   └── setup.sh
└── assets/
    └── template.json
```

When a skill is installed to a platform, the directory is copied — minus anything the skill's `install.exclude` rules out (below). The `scripts/` directory is language-agnostic — skills can include Python, shell, JavaScript, or any other scripts. The agent decides which language is appropriate.

`skern skill validate <name>` checks that files referenced in the body (via backticks or markdown links) actually exist in the skill directory. Missing references produce **warnings**, not errors. `skern skill show <name>` lists every file bundled with the skill.

### Install-time exclusions

A skill directory often carries assets that belong in the registry but not in a platform's skill directory — evaluation corpora, fixtures, development scratch. Because host agents load or index everything under an installed skill, that material is dead weight in the agent's context on every project. The skill author lists what to leave behind in frontmatter:

```yaml
install:
  exclude:
    - eval            # a directory: excludes eval/ and everything beneath it
    - fixtures/*      # direct children of fixtures/ (and their subtrees)
    - "*.draft.md"    # top-level files by suffix (globs do not cross "/")
```

Rules:

- Patterns are slash-separated paths relative to the skill directory, using Go `path.Match` syntax (`*`, `?`, `[…]`; `\` escapes a metacharacter). `*` does not cross `/`, and `**` is **not** supported (it is rejected, not silently treated as `*`); a pattern matches a path when it matches the full relative path **or any leading directory of it**, which is what makes a bare directory name exclude its whole subtree. A trailing `/` or leading `./` is ignored. `exclude: eval` (a bare string) is accepted as shorthand for a one-element list.
- `SKILL.md` is never excluded, whatever the patterns say — so `*` or `*.md` are legal and mean "everything except `SKILL.md`". Naming `SKILL.md` literally, absolute paths, `..` segments, `**`, and malformed globs are validation **errors**, and `skern skill install` refuses a skill whose patterns fail validation rather than silently copying what the author meant to leave out.
- The registry always keeps the full directory — `skill create --from-template` and `skill import` copy everything, and `skill show` lists every file. Only the copy made by `skern skill install` is trimmed. (A pattern such as `fixtures/*` prunes the directory's contents; the now-empty `fixtures/` itself is still created.)
- `skern skill validate` **warns** when a pattern matches no file in the skill directory, and when a file the body references (`` `references/guide.md` ``) is excluded — it would exist in the registry but be missing from every installed copy.
- `skern skill diff` compares manifests (frontmatter and body), not file trees, so excluded files never show up as drift — but a change to the `install.exclude` list itself *is* reported (`install.exclude`), since it changes what the next install copies.
- `install` is a skern-modeled key: other keys under it (`install.mode`, …) round-trip like any other unmodeled key, but a top-level `install:` that is not a mapping (e.g. `install: pip install foo`) is a parse error.

## Creating Skills

Use `skern skill create` to scaffold a new skill:

```sh
skern skill create code-review \
  --description "Use when reviewing PRs for style and correctness" \
  --author "Jane Doe" \
  --author-type human \
  --version 1.0.0 \
  --tags review,quality
```

Or seed from another skill. `--from-template <dir>` requires a **skill directory** — a directory containing `SKILL.md` and any optional companion files (`references/`, `templates/`, `VENDORED.md`, …). Skern parses the template's frontmatter and copies every sibling into the new skill:

```sh
skern skill create code-review \
  --from-template ~/.skern/skills/source-template
```

A bare file path (a `SKILL.md` or a body-only markdown file) is rejected with an error pointing at the parent directory. The CLI `<name>` argument always wins over the template's name; other flags (`--description`, `--tags`, `--author*`, `--version`) override template values when explicitly set, otherwise the template's values are preserved.

## Importing Skills

`skern skill import <url>` pulls an existing skill into the registry from a GitHub repository directory or gist:

```sh
skern skill import https://github.com/owner/repo/tree/main/skills/code-review
skern skill import https://gist.github.com/<id>
skern skill import <url> --name local-name --scope project
```

Overlap detection runs on import too — pass `--force` to override a near-duplicate block.
