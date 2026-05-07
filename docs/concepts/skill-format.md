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
| `name` | Yes | Skill name matching `[a-z0-9]+(-[a-z0-9]+)*`, 1-64 chars. Must equal the directory name. |
| `description` | Yes | What the skill does — start with "Use when…". Max 1024 chars. |
| `tags` | No | List of classification tags |
| `allowed-tools` | No | Tools the skill may use. No empty entries. |
| `metadata.author.name` | No | Author name |
| `metadata.author.type` | No | `human` or `agent` |
| `metadata.author.platform` | No | Platform name (e.g. `claude-code`) — used when `type: agent` |
| `metadata.version` | No | Semantic version (e.g. `1.0.0`); defaults to `0.0.1` on `skill create` |
| `metadata.modified-by` | No | Append-only modification history (set via `skern skill edit --modified-by`) |

`name` and `description` are the only hard requirements. The rest help discovery, validation, and provenance.

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

When a skill is installed to a platform, the **entire directory** is copied. The `scripts/` directory is language-agnostic — skills can include Python, shell, JavaScript, or any other scripts. The agent decides which language is appropriate.

`skern skill validate <name>` checks that files referenced in the body (via backticks or markdown links) actually exist in the skill directory. Missing references produce **warnings**, not errors. `skern skill show <name>` lists every file bundled with the skill.

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
