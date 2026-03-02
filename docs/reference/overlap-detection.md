# Overlap Detection

When creating a skill, skern checks existing skills for similarity to prevent duplication. This runs automatically during `skern skill create` and during `skern skill list`.

## Scoring Methods

### Fuzzy Name Matching

Uses Levenshtein distance with prefix and suffix bonuses to compare skill names. This catches near-identical names like `code-review` and `code-reveiw`.

### Description Similarity

Keyword overlap scoring using Jaccard similarity between skill descriptions. This identifies skills that serve the same purpose even if named differently.

### Tools Overlap

Shared `allowed-tools` entries between skills contribute to the overall similarity score.

## Thresholds

| Score | Behavior |
|-------|----------|
| < 0.6 | Proceed normally |
| >= 0.6 | Warn — show similar skills, continue |
| >= 0.9 | Block — require `--force` to override |

## Skill Count Warnings

Warnings are triggered when registries grow beyond recommended sizes:

- **Project scope** — warns at > 20 skills
- **User scope** — warns at > 50 skills

## During `skill list`

`skern skill list` runs pairwise overlap detection across all listed skills and appends a "Potential duplicates" section when matches are found (score >= 0.6). In `--json` mode, these appear in the `duplicates` array.

## Overriding

Use the `--force` flag to bypass overlap blocking during skill creation:

```sh
skern skill create my-skill --description "..." --force
```
