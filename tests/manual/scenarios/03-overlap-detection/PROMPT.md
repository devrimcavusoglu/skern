# Scenario 3: Overlap Detection Workflow

## Pre-populated skills (project scope)

- `code-review` — Review code changes and suggest code improvements
- `lint-python` — Lint Python source code and report lint errors

## Prompt to give the agent

> Create a new skill called "code-reviewer" that reviews code changes and provides code improvement suggestions. Use project scope.

## What to observe

1. Does the agent attempt `skern skill create code-reviewer --description "..." --scope project`?
2. Does skern warn about overlap with the existing `code-review` skill (score >= 0.6)?
3. Does the agent understand the warning and communicate it to the user?
4. Does the agent make an informed decision: reuse `code-review`, or proceed with creation?

## Follow-up prompt

> Now create a skill called "python-linter" for linting Python code and reporting lint errors. Use project scope.

This should trigger overlap with `lint-python`. Observe the same behavior.
