# Expected Behavior — Scenario 7

## Pass criteria

- [ ] `json-fmt` created in project scope
- [ ] `api-docs` created in project scope
- [ ] Both have descriptions
- [ ] `scribe skill list --scope project --json` returns both skills
- [ ] Agent shows disk paths from `scribe skill show` output
- [ ] Both pass validation

## Verification commands

```sh
scribe skill list --scope project --json    # count: 2
scribe skill show json-fmt --scope project --json
scribe skill show api-docs --scope project --json
scribe skill validate json-fmt
scribe skill validate api-docs
```
