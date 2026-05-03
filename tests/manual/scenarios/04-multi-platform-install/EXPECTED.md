# Expected Behavior — Scenario 4

## Pass criteria

- [ ] Agent discovers all 3 platforms via `skern platform list`
- [ ] Agent installs to each platform individually (one `--platform <name>` invocation per platform; `--platform all` was removed in v0.2.0)
- [ ] Agent verifies with `skern platform status`
- [ ] Status shows deploy-helper installed on claude-code, codex-cli, opencode
- [ ] Follow-up: agent uninstalls from codex-cli only
- [ ] Updated status shows deploy-helper on claude-code + opencode, not codex-cli

## Verification commands

```sh
skern platform list --json
skern platform status --json
skern skill install deploy-helper --platform claude-code
skern skill install deploy-helper --platform codex-cli
skern skill install deploy-helper --platform opencode
skern platform status --json
skern skill uninstall deploy-helper --platform codex-cli
skern platform status --json
```
