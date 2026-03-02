# Expected Behavior — Scenario 3

## Pass criteria

- [ ] Agent attempts to create `code-reviewer` in project scope
- [ ] Skern shows overlap warning with `code-review` (score >= 0.6)
- [ ] Agent recognizes the overlap warning from command output
- [ ] Agent makes an informed decision (reuse existing or proceed)
- [ ] Agent communicates the overlap to the user
- [ ] Follow-up: `python-linter` triggers overlap with `lint-python`

## Key note on overlap scoring

The overlap algorithm uses keyword matching (no stemming). For the warning to trigger
(threshold 0.6), the description keywords must overlap significantly with the existing
skill. The prompt wording guides the agent toward similar keywords.

## Verification commands

```sh
# This should show overlap warning:
skern skill create code-reviewer --description "Review code changes and provide code improvement suggestions" --scope project

# Check overlap interactively:
skern skill recommend "review code changes" --scope project --json
```
