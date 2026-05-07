# Overlap Detection & Capacity

When creating a skill, skern checks existing skills for similarity to prevent duplication. The same scoring runs during `skern skill list` (pairwise) and `skern skill import`. Capacity thresholds are reported on every install/uninstall to help agents size their working set.

## Scoring Methods

### Fuzzy Name Matching

Levenshtein distance with prefix and suffix bonuses. Catches near-identical names like `code-review` and `code-reveiw`.

### Description Similarity

Keyword overlap (Jaccard similarity) between descriptions. Identifies skills that serve the same purpose even when named differently.

### Tools Overlap

Shared `allowed-tools` entries between two skills contribute to the overall score.

## Thresholds

| Score | Behavior |
|-------|----------|
| < 0.6 | Proceed normally |
| ≥ 0.6 | **Warn** — show similar skills, continue |
| ≥ 0.9 | **Block** — require `--force` to override |

## Skill Count Warnings

The registry warns when it grows past recommended sizes:

- **Project scope** (`.skern/skills/`) — warns at >= 20 skills
- **User scope** (`~/.skern/skills/`) — warns at >= 50 skills

These are soft limits — skern keeps working, but encourages cleanup.

## Per-Platform Capacity

Every `skern skill install` and `skern skill uninstall` response carries a `capacity` block:

```json
{
  "capacity": {
    "platform": "claude-code",
    "scope": "user",
    "installed": 18,
    "threshold": 50,
    "headroom": 32,
    "over_budget": false
  }
}
```

Thresholds match the registry: 20 for project scope, 50 for user scope. They are computed against the **platform's** installed-skills directory — when several adapters share `.agents/skills/` as a project directory, both adapters see the same count.

`--enforce-budget` on `skill install` makes the operation refuse to proceed when the resulting count would exceed the threshold. Off by default, so agents must explicitly opt in.

## During `skill list`

`skern skill list` runs pairwise overlap detection across all listed skills and appends a "Potential duplicates" section when matches are found (score ≥ 0.6). In `--json` mode they appear in the `duplicates` array. Use `--with-platforms` to also see where each skill is currently installed.

## Overriding

Use `--force` to bypass the overlap block during creation or import:

```sh
skern skill create my-skill --description "Use when …" --force
skern skill import https://github.com/owner/repo/tree/main/skill --force
```
