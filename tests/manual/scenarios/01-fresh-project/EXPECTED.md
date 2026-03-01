# Expected Behavior — Scenario 1

## Pass criteria

- [ ] Agent discovers skern via AGENTS.md
- [ ] Agent searches for existing skills before creating (recommend or search)
- [ ] Agent creates the skill with a valid name (e.g., `go-formatter`, `format-go`)
- [ ] Agent provides a `--description` flag
- [ ] Skill is created in project scope: `.skern/skills/<name>/SKILL.md`
- [ ] Agent optionally validates the skill with `skern skill validate <name>`

## Verification commands

```sh
skern skill list --scope project --json
skern skill show <name> --scope project --json
skern skill validate <name>
```
