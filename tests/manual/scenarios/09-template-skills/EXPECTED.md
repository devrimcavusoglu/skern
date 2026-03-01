# Expected Behavior — Scenario 9

## Pass criteria

- [ ] Agent uses `--from-template templates/review.md` for code-review
- [ ] Agent uses `--from-template templates/test-helper.md` for test-helper
- [ ] Both skills provide a `--description` flag
- [ ] Both skills pass `skern skill validate` with no errors
- [ ] Skill body content matches the template files
- [ ] Agent verifies with `skern skill show <name>`

## Verification commands

```sh
skern skill create code-review --description "Reviews code changes" --from-template templates/review.md
skern skill create test-helper --description "Helps write tests" --from-template templates/test-helper.md
skern skill validate code-review
skern skill validate test-helper
skern skill show code-review
skern skill show test-helper
```
