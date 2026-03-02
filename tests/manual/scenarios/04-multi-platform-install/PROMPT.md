# Scenario 4: Multi-Platform Install

## Pre-populated

- Skill: `deploy-helper` — Assists with deployment steps and checklists
- Platform dirs: `.claude/`, `.agents/`, `.opencode/` (all three detected)

## Prompt to give the agent

> Install the deploy-helper skill to all available platforms, then show me which platforms have it.

## What to observe

1. Does the agent check which platforms are available (`skern platform list`)?
2. Does it use `--platform all` or install one-by-one?
3. Does it verify installation with `skern platform status`?
4. Does the status output show `deploy-helper` installed on all three platforms?

## Follow-up prompt

> Uninstall deploy-helper from codex-cli only, then show the status again.

## What to observe for follow-up

1. Does the agent uninstall from only codex-cli?
2. Does the updated status show deploy-helper on claude-code and opencode but NOT codex-cli?
