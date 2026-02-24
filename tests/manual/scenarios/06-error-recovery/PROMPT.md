# Scenario 6: Error Recovery

## Pre-populated

- Skill: `test-runner` (project scope)
- Platform: `.claude/` detected

## Prompt to give the agent

> Do the following steps in order:
> 1. Install a skill called "nonexistent-skill" to claude-code
> 2. Create a skill with the name "INVALID_NAME"
> 3. Install test-runner to claude-code, then try installing it again
> 4. Remove test-runner, then try to show its details

## What to observe

For each step:
1. **Install nonexistent** — Does the agent see the "not found" error? Does it recover gracefully?
2. **Invalid name** — Does the agent get exit code 2 (validation error)? Does it understand why?
3. **Double install** — Does the agent understand "already installed" and not retry?
4. **Remove then show** — Does the agent understand the skill is gone?

Key question: Does the agent parse error messages from `--json` output and take corrective action, or does it get stuck?
