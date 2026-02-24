# Scenario 1: Agent Discovery & First Skill Creation

## Prompt to give the agent

> Create a skill that formats Go source files using gofmt. Put it in project scope.

## What to observe

1. Does the agent read AGENTS.md and discover scribe?
2. Does it search/recommend first before creating?
3. Does it run `scribe skill search "go"` or `scribe skill recommend "format Go source files"`?
4. Since the registry is empty, does it proceed to create?
5. Does it use `scribe skill create` with a reasonable name, description, and `--scope project`?
6. Does it inspect the created SKILL.md and optionally write meaningful body content?
