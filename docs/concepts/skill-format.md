# Skill Format

Skills follow the [Agent Skills](https://agentskills.io) open standard. Each skill is a single `SKILL.md` file with YAML frontmatter and a markdown body.

## Structure

```markdown
---
name: code-review
description: Review PRs for style and correctness
tags:
  - review
  - quality
version: 1.0.0
author:
  name: Jane Doe
  type: human
allowed-tools:
  - Read
  - Grep
  - Glob
---

## Instructions

Review pull requests for:
- Code style consistency
- Correctness of logic
- Test coverage
```

## Frontmatter Fields

| Field | Required | Description |
|-------|----------|-------------|
| `name` | Yes | Skill name matching `[a-z0-9]+(-[a-z0-9]+)*`, 1-64 chars |
| `description` | Yes | What the skill does, max 1024 chars |
| `version` | No | Semantic version (e.g., `1.0.0`) |
| `author.name` | No | Author name |
| `author.type` | No | `human` or `agent` |
| `author.platform` | No | Platform name (e.g., `claude-code`) |
| `tags` | No | List of classification tags |
| `allowed-tools` | No | List of tools the skill may use |
| `modified-by` | No | Modification history entries |

## Body

The markdown body contains the skill's instructions. This is what the agent reads when the skill is activated. It must be non-empty.

## Author Provenance

Skills track author metadata and an optional `modified-by` history. `skern skill show` displays the full provenance chain when present, including editor name, type (human/agent), platform, and date.

## Folder Structure

Skills can include additional files alongside `SKILL.md` — helper scripts, templates, configuration files, and other assets. When a skill is installed to a platform, the entire directory is copied.

```
my-skill/
├── SKILL.md
├── scripts/
│   ├── convert.py
│   └── setup.sh
└── assets/
    └── template.json
```

The `scripts/` directory is language-agnostic — skills can include Python, shell, JavaScript, or any other scripts. The agent decides which language is appropriate for the skill.

Use `skern skill show <name>` to see which files are bundled with a skill, and `skern skill validate <name>` to check that files referenced in the skill body actually exist.

## Creating Skills

Use `skern skill create` to scaffold a new skill:

```sh
skern skill create code-review \
  --description "Review PRs for style and correctness" \
  --author "Jane Doe" \
  --author-type human
```

Or seed from another skill. `--from-template <dir>` requires a **skill directory**
— a directory containing `SKILL.md` and any optional companion files
(`references/`, `templates/`, `VENDORED.md`, …). Skern parses the template's
frontmatter and copies every sibling into the new skill:

```sh
# Seeds the new skill from an existing skill directory. SKILL.md frontmatter
# (description, tags, metadata.author, metadata.version) is preserved; all
# sibling files and subdirectories are copied alongside the new SKILL.md.
skern skill create code-review --from-template ~/.skern/skills/source-template
```

A bare file path (e.g., a `SKILL.md` or a markdown body file) is rejected with
an error pointing you at the parent directory. The CLI `<name>` argument
always wins over the template's `name`; other flags
(`--description`, `--tags`, `--author*`, `--version`) override template values
when explicitly set, otherwise the template's values are preserved.
