# Scenario 7: Scoped Skill Management

## Pre-populated

- Empty scribe registry (just initialized)

## Prompt to give the agent

> Create two skills in project scope:
> 1. A utility skill called "json-fmt" for formatting JSON files
> 2. A skill called "api-docs" for generating API documentation
>
> Then list skills and show me details of each.

## What to observe

1. Does the agent use `--scope project` for both skills?
2. Does it provide `--description` for each?
3. Does the agent verify creation with:
   - `scribe skill list --scope project` (both skills)
   - `scribe skill show <name> --scope project`
4. Does the agent understand that skills are stored under `.scribe/skills/<name>/`?

## Follow-up prompt

> Validate both skills and show me where they are stored on disk.

The agent should use `scribe skill validate <name>` and read the `path` field from `scribe skill show`.
