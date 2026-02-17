# Expected Behavior — Scenario 5

## Pass criteria

- [ ] All scribe commands use `--json` flag
- [ ] Agent correctly parses JSON at each step
- [ ] Search/recommend returns no matches (empty registry)
- [ ] Skill is created with a valid name and description in project scope
- [ ] Validation returns valid=true (or only warnings, no errors)
- [ ] Install succeeds for claude-code
- [ ] Platform status shows the skill installed on claude-code
- [ ] Agent summarizes results from JSON output (doesn't just dump raw JSON)

## Verification commands

```sh
scribe skill list --scope project --json    # should show 1 skill
scribe platform status --scope project --json  # should show skill on claude-code
```
