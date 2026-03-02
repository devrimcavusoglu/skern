# Scenario 10: Platform Status Matrix

## Pre-populated

- Skills: `go-formatter`, `db-migrate`, `api-docs` (all project scope)
- Platforms: `.claude/`, `.agents/`, `.opencode/` (all detected)
- Partial installs:
  - `go-formatter` -> claude-code, codex-cli (NOT opencode)
  - `db-migrate` -> claude-code (NOT codex-cli, NOT opencode)
  - `api-docs` -> NOT installed anywhere

## Prompt to give the agent

> Show me which skills are installed on which platforms. Identify any gaps where skills are missing from platforms, and install the missing ones so every skill is available everywhere.

## What to observe

1. Does the agent run `skern platform status --json` to get the matrix?
2. Does it correctly identify the gaps:
   - `go-formatter` missing from opencode
   - `db-migrate` missing from codex-cli and opencode
   - `api-docs` missing from all platforms
3. Does it install the missing combinations?
4. Does it verify the final state with `skern platform status` again?
5. Does the final matrix show all 3 skills on all 3 platforms?
