# Scenario 5: Full Lifecycle with JSON Mode

## Pre-populated

- Empty scribe registry, `.claude/` dir present for platform detection

## Prompt to give the agent

> I need a skill for running database migrations. Do the full setup: check if one exists, create it in project scope if not, validate it, install it to claude-code, and show me the final platform status. Use --json for all commands and show me the structured output at each step.

## What to observe

1. Does the agent use `--json` on every scribe command?
2. Does it chain commands logically based on previous JSON output?
3. Full expected sequence:
   - `scribe skill search "migrate" --json` or `scribe skill recommend "..." --json` -> no match
   - `scribe skill create db-migrate --description "..." --scope project --json`
   - `scribe skill validate db-migrate --json` -> valid: true
   - `scribe skill install db-migrate --platform claude-code --scope project --json`
   - `scribe platform status --scope project --json` -> shows db-migrate on claude-code
4. Does the agent parse JSON correctly at each step (not just dump it)?
