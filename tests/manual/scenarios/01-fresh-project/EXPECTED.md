# Expected Behavior — Scenario 1

## Pass criteria

- [ ] Agent discovers scribe via AGENTS.md
- [ ] Agent searches for existing skills before creating (recommend or search)
- [ ] Agent creates the skill with a valid name (e.g., `go-formatter`, `format-go`)
- [ ] Agent provides a `--description` flag
- [ ] Skill is created in project scope: `.scribe/skills/<name>/SKILL.md`
- [ ] Agent optionally validates the skill with `scribe skill validate <name>`

## Verification commands

```sh
scribe skill list --scope project --json
scribe skill show <name> --scope project --json
scribe skill validate <name>
```
