# Expected Behavior — Scenario 9

## Pass criteria

- [ ] Agent uses `--from-template templates/review.md` for code-review
- [ ] Agent uses `--from-template templates/test-helper.md` for test-helper
- [ ] Both skills provide a `--description` flag
- [ ] Both skills pass `scribe skill validate` with no errors
- [ ] Skill body content matches the template files
- [ ] Agent verifies with `scribe skill show <name>`

## Verification commands

```sh
scribe skill create code-review --description "Reviews code changes" --from-template templates/review.md
scribe skill create test-helper --description "Helps write tests" --from-template templates/test-helper.md
scribe skill validate code-review
scribe skill validate test-helper
scribe skill show code-review
scribe skill show test-helper
```
