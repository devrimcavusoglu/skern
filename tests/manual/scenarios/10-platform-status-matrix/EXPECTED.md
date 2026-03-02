# Expected Behavior — Scenario 10

## Pass criteria

- [ ] Agent reads platform status matrix
- [ ] Agent identifies all gaps (5 missing installations)
- [ ] Agent installs missing skill/platform combinations
- [ ] Agent does NOT re-install already-installed combinations
- [ ] Final status shows all 9 cells as installed (3 skills x 3 platforms)
- [ ] Agent summarizes what was done

## Expected gaps to fill

| Skill | claude-code | codex-cli | opencode |
|-------|-------------|-----------|----------|
| go-formatter | installed | installed | **MISSING** |
| db-migrate | installed | **MISSING** | **MISSING** |
| api-docs | **MISSING** | **MISSING** | **MISSING** |

## Verification commands

```sh
# Before:
skern platform status --json

# After all installs:
skern platform status --json
# All skills should show installed=true on all platforms
```
