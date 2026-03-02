# Scenario 2: Agent Reuses Existing Skill

## Pre-populated skills (project scope)

- `go-formatter` — Formats Go source files using gofmt and goimports
- `python-linter` — Lints Python code with ruff and reports issues
- `markdown-toc` — Generates table of contents for markdown files

## Prompt to give the agent

> I need a skill that can format my Go code. Check what we already have before creating anything new.

## What to observe

1. Does the agent discover skern via AGENTS.md?
2. Does it search for existing skills first?
   - `skern skill search "go"` or `skern skill search "formatter"` (substring match)
   - or `skern skill recommend "format go source files" --scope project`
3. Does it find the existing `go-formatter` skill?
4. Does it inspect the skill (`skern skill show go-formatter --scope project`)?
5. Does it recommend reusing the existing skill instead of creating a new one?
6. Does it avoid creating a duplicate?
