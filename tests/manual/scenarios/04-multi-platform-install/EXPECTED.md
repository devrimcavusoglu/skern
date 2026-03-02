# Expected Behavior — Scenario 4

## Pass criteria

- [ ] Agent discovers all 3 platforms via `skern platform list`
- [ ] Agent installs to all platforms (via `--platform all` or individually)
- [ ] Agent verifies with `skern platform status`
- [ ] Status shows deploy-helper installed on claude-code, codex-cli, opencode
- [ ] Follow-up: agent uninstalls from codex-cli only
- [ ] Updated status shows deploy-helper on claude-code + opencode, not codex-cli

## Verification commands

```sh
skern platform list --json
skern platform status --json
skern skill install deploy-helper --platform all
skern platform status --json
skern skill uninstall deploy-helper --platform codex-cli
skern platform status --json
```
