# Expected Behavior — Scenario 2

## Pass criteria

- [ ] Agent searches/recommends before attempting to create
- [ ] Agent finds `go-formatter` via search or recommend
- [ ] Agent inspects the existing skill to understand what it does
- [ ] Agent does NOT create a new skill
- [ ] Agent tells the user the existing skill matches their need

## Verification commands

```sh
skern skill search "formatter" --json
# Expected: results include go-formatter

skern skill show go-formatter --scope project --json
# Expected: full skill details

skern skill list --scope project --json
# Should still show exactly 3 skills (no new ones created)
```
