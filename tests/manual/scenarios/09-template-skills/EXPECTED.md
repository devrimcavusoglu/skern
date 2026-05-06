# Expected Behavior — Scenario 9

## Pass criteria

- [ ] Agent uses `--from-template templates/code-review` (the directory) for code-review
- [ ] Agent uses `--from-template templates/test-helper` (the directory) for test-helper
- [ ] Agent does NOT pass a `SKILL.md` file path to `--from-template` (the CLI rejects that)
- [ ] Both skills pass `skern skill validate` with no errors
- [ ] Skill body content matches the template SKILL.md body sections
- [ ] Skill `description` and `metadata.version` come from the template's frontmatter (not placeholders)
- [ ] Agent verifies with `skern skill show <name>`

## Verification commands

```sh
skern skill create code-review --from-template templates/code-review
skern skill create test-helper --from-template templates/test-helper
skern skill validate code-review
skern skill validate test-helper
skern skill show code-review
skern skill show test-helper
```
