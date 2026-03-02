# Agent Setup

After installing skern, you can enable the **tool-forming loop** — a pattern where agents search for existing skills before creating new ones, keeping your skill set deduplicated and organized.

## Enable the Loop

Add a line to your project's `AGENTS.md` (or `CLAUDE.md`) so that agents know to use skern for skill management:

```sh
echo 'Use skern to manage skills. Run `skern --help` for usage, `skern skill search <query>` to find existing skills before creating new ones.' >> AGENTS.md
```

## How It Works

When an agent encounters a recurring task:

1. The agent reads the instruction from `AGENTS.md`
2. It runs `skern skill search <query>` to check for existing skills
3. If a matching skill exists, the agent reuses it
4. If no match is found, the agent creates a new skill with `skern skill create`
5. The new skill is immediately available for future sessions

## JSON Mode for Agents

Every skern command supports `--json` for machine-readable output:

```sh
skern skill list --json
skern skill search "review" --json
skern skill show code-review --json
```

This makes skern fully agent-operable — agents can parse responses and make decisions programmatically.

## Recommended Agent Instructions

For more comprehensive agent integration, add these to your `AGENTS.md`:

```markdown
## Skill Management

- Use `skern` to manage reusable skills
- Before creating a new skill, search with `skern skill search <query>`
- Use `skern skill recommend <query>` to get suggestions before creating
- Always validate skills after creation: `skern skill validate <name>`
- Install to the current platform: `skern skill install <name> --platform all`
```
