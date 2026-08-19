# Validation

`skern skill validate <name>` checks skills against the [Agent Skills](https://agentskills.io) specification. Validation also runs automatically during `skern skill create` (warnings only — never blocks creation).

## Severity

Validation reports three severities:

| Severity | Prefix (text) | JSON `severity` | Effect |
|----------|---------------|-----------------|--------|
| **Error** | `✗` | `"error"` | `valid: false`, exit code `2` |
| **Warning** | `!` | `"warning"` | Reported, does not affect validity |
| **Hint** | `~` | `"hint"` | Stylistic suggestion only |

The summary line counts each separately, e.g. `Skill "x" has 1 error(s), 0 warning(s), 2 hint(s)`.

## Errors

### Name Format

Skill names must match `[a-z0-9]+([.-][a-z0-9]+)*` and be 1–64 characters. Lowercase alphanumeric segments are joined by hyphens or dots; dots enable namespace-style names used by skill installers.

Valid: `code-review`, `lint-fix`, `deploy`, `myorg.bootstrap`, `codebase-intelligence.scan`. Invalid: `Code_Review`, `my skill`, `.bootstrap`, `myorg..bootstrap`.

### Description

Required, non-empty, max 1024 characters.

### Body

The `SKILL.md` body (everything after the YAML frontmatter) must not be empty.

### Allowed Tools

If `allowed-tools` is set, no entry may be empty.

### Author Type

`metadata.author.type` must be `human` or `agent` if set.

### Version

`metadata.version` must be a valid semver string (`MAJOR.MINOR.PATCH`).

### Install Exclusions

Every `install.exclude` entry must be a non-empty, relative, well-formed glob (`path.Match` syntax) with no `..` segment, and must not match `SKILL.md`. See [Skill Format › Install-time exclusions](/concepts/skill-format#install-time-exclusions).

## Warnings

### Folder Integrity

When the body references files (backtick-enclosed paths like `` `scripts/run.py` `` or markdown links like `[script](scripts/run.py)`), validation checks that those files exist in the skill directory. Missing references produce **warnings**, not errors — references may be aspirational or provided at runtime.

Inline code spans and URLs are excluded from this check (no false positives from `` `--flag` `` or `https://…`).

### Install Exclusions

Two warnings relate to `install.exclude`: a pattern that matches no file in the skill directory (usually a typo, or stale after a rename), and a body-referenced file that an exclude pattern would drop — it exists in the registry but would be missing from every installed copy.

## Hints

Stylistic suggestions adapted from [writing-skills](/writing-skills). They never block creation or validation.

| Hint | Trigger |
|------|---------|
| Body too short | Body has fewer than 20 words |
| Description too vague | Description has fewer than 3 words |
| Missing trigger prefix | Description doesn't start with "Use when", "Use for", "Use to", "Trigger when", or "Apply when" |
| No step markers | Body lacks bullet points, numbered lists, or "step" markers |
| Missing recommended sections | Body is missing one or more of `When to Use`, `Core Pattern`, `Quick Reference`, `Common Mistakes` |

## Parse Warnings

When `skern skill list` encounters a skill directory that cannot be parsed (e.g. malformed YAML frontmatter), it reports a **parse warning** instead of silently skipping the entry. In text mode these appear as `WARNING:` lines; in `--json` mode they populate the `parse_warnings` array.

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Skill is valid (warnings/hints OK) |
| 2 | Validation failure — at least one error |
