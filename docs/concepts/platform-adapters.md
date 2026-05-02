# Platform Adapters

Platform adapters bridge the skill registry with agent runtimes. Each adapter knows the target platform's directory structure and copies skill files accordingly.

## How Adapters Work

When you run `skern skill install`, the adapter:

1. Reads the `SKILL.md` from the skern registry
2. Creates the platform-specific skill directory
3. Copies the `SKILL.md` into the target location
4. The agent runtime discovers the skill on its next invocation

## Supported Platforms

| Platform | Adapter name | Detection |
|----------|-------------|-----------|
| Claude Code | `claude-code` | Looks for `.claude/` or `~/.claude/` |
| Codex CLI | `codex-cli` | Looks for `.agents/` or `~/.agents/` |
| OpenCode | `opencode` | Looks for `.opencode/` or `~/.config/opencode/` |

## Installation Paths

Each platform uses different directories for user-level and project-level skills:

| Platform | User-level | Project-level |
|----------|-----------|---------------|
| Claude Code | `~/.claude/skills/<name>/` | `.claude/skills/<name>/` |
| Codex CLI | `~/.agents/skills/<name>/` | `.agents/skills/<name>/` |
| OpenCode | `~/.config/opencode/skills/<name>/` | `.opencode/skills/<name>/` |

## Auto-Detection

Skern auto-detects which platforms are installed on your system. Use `skern platform list` to see detected platforms:

```sh
skern platform list
```

## One Platform per Invocation

Each `skern skill install` call targets exactly one platform. Agents are expected to specify the platform they are running on — there is no `all` value, and skern does not broadcast skills across platforms automatically.

This design supports the [dynamic skill loading](./registry) model: each agent maintains its own working set of installed skills, sized to its context budget, independent of other platforms.

To deploy a skill across several platforms, loop the call:

```sh
for p in claude-code codex-cli opencode; do
  skern skill install code-review --platform "$p"
done
```

## Batch Install/Uninstall

Multiple skills can be installed (or uninstalled) in one call:

```sh
skern skill install code-review test-runner deploy-checker --platform claude-code
skern skill uninstall stale-a stale-b --platform claude-code
```

Each skill's outcome is reported separately in the JSON output's `skills` array, and a `capacity` block reports the platform's installed-skill count after the batch.

## Platform Status Matrix

View which skills are installed on which platforms:

```sh
skern platform status
```

This shows a matrix of skills and their installation status across all detected platforms.
