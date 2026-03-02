# Expected Behavior — Scenario 7

## Pass criteria

- [ ] `json-fmt` created in project scope
- [ ] `api-docs` created in project scope
- [ ] Both have descriptions
- [ ] `skern skill list --scope project --json` returns both skills
- [ ] Agent shows disk paths from `skern skill show` output
- [ ] Both pass validation

## Verification commands

```sh
skern skill list --scope project --json    # count: 2
skern skill show json-fmt --scope project --json
skern skill show api-docs --scope project --json
skern skill validate json-fmt
skern skill validate api-docs
```
