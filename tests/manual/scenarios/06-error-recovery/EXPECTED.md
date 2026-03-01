# Expected Behavior — Scenario 6

## Pass criteria

- [ ] Step 1: Agent sees "not found" error, reports it, moves on
- [ ] Step 2: Agent sees validation error (exit code 2), understands invalid name format
- [ ] Step 3: First install succeeds, second returns "already installed" error
- [ ] Step 4: Remove succeeds, show returns "not found"
- [ ] Agent does not get stuck in retry loops on any error
- [ ] Agent correctly interprets error JSON: `{"error": "..."}`
- [ ] Agent reports each error clearly to the user

## Error outputs to expect

```sh
skern skill install nonexistent-skill --platform claude-code --json
# {"error": "skill \"nonexistent-skill\" not found..."}

skern skill create INVALID_NAME --json
# {"error": "..."} with exit code 2

skern skill install test-runner --platform claude-code --json  # first: ok
skern skill install test-runner --platform claude-code --json  # second: already installed

skern skill remove test-runner --json   # ok
skern skill show test-runner --json     # not found
```
