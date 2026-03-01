# Scenario 8: Deduplication Advisory

## Pre-populated skills (project scope)

- `test-runner` — Run test suites and report test results for the project
- `run-tests` — Run test suites and report results across the project
- `test-runner-v2` — Run test suites and report test results with coverage
- `code-review` — Review code changes and suggest code improvements
- `code-reviewer` — Review code changes and provide code improvement suggestions

## Prompt to give the agent

> Audit my skills for duplicates. List all skills and tell me which ones overlap. Suggest which to keep and which to remove.

## What to observe

1. Does the agent run `skern skill list --scope project --json`?
2. Does it inspect the `duplicates` array in the response?
3. Does it identify the overlapping pairs:
   - `test-runner` <-> `test-runner-v2` (score ~0.71)
   - `code-review` <-> `code-reviewer` (score ~0.61)
4. Does it notice that `run-tests` and `test-runner` are similar but below the threshold?
5. Does it provide actionable advice (which to keep, which to remove)?
6. Does it offer to run `skern skill remove` for the duplicates?
