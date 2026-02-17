# Expected Behavior — Scenario 4

## Pass criteria

- [ ] Agent discovers all 3 platforms via `scribe platform list`
- [ ] Agent installs to all platforms (via `--platform all` or individually)
- [ ] Agent verifies with `scribe platform status`
- [ ] Status shows deploy-helper installed on claude-code, codex-cli, opencode
- [ ] Follow-up: agent uninstalls from codex-cli only
- [ ] Updated status shows deploy-helper on claude-code + opencode, not codex-cli

## Verification commands

```sh
scribe platform list --json
scribe platform status --json
scribe skill install deploy-helper --platform all
scribe platform status --json
scribe skill uninstall deploy-helper --platform codex-cli
scribe platform status --json
```
