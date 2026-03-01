# Scenario 11: Autonomous Skill Creation

## Pre-populated

- Go project with 3 packages under `pkg/`:
  - `pkg/auth/auth.go` — Clean formatting, no doc comments
  - `pkg/handler/handler.go` — Bad formatting: ungrouped imports, missing spaces, inconsistent indentation
  - `pkg/store/store.go` — Well-formatted, partial doc comments (some functions missing them)
- Empty scribe registry (initialized but no skills)
- `.claude/` directory present (platform detection)

## Prompt to give the agent

> We have a Go project with 3 packages (auth, handler, store) and new contributors are joining the team. I want to make sure our Go code stays clean and consistent going forward — proper formatting, doc comments on exported functions, grouped imports, etc. Set things up so that any agent working in this repo can enforce these standards.

## What to observe

1. Does the agent read the Go source files to understand what inconsistencies exist?
2. Does the agent discover `scribe` by reading AGENTS.md?
3. Does the agent **autonomously reason** that a reusable skill is the right approach (not just fix the files directly)?
4. Does it search/recommend before creating to avoid duplicates?
5. Does it create at least one skill with a valid name, description, and meaningful body?
6. Does it validate the skill after creation?
7. Does the agent also fix the existing issues in the Go files, or just create the skill?
