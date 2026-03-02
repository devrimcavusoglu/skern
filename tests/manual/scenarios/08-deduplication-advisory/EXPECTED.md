# Expected Behavior — Scenario 8

## Pass criteria

- [ ] Agent lists skills and checks for duplicates
- [ ] Agent identifies `test-runner` / `run-tests` / `test-runner-v2` as overlapping
- [ ] Agent identifies `code-review` / `code-reviewer` as overlapping
- [ ] Agent suggests which to keep (e.g., keep `test-runner-v2` as most capable)
- [ ] Agent offers to remove the others (with user confirmation)
- [ ] Agent uses the `duplicates` array from `skern skill list --json` output

## Verification commands

```sh
skern skill list --json
# Expected: duplicates array with pairs like:
# {"skill_a": "test-runner", "skill_b": "run-tests", "score": ...}
# {"skill_a": "code-review", "skill_b": "code-reviewer", "score": ...}
```
