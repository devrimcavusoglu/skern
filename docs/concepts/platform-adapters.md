# Platform Adapters

Platform adapters bridge the skill registry with agent runtimes. Each adapter knows the target platform's directory structure and copies skill files accordingly.

## How Adapters Work

When you run `skern skill install`, the adapter:

1. Reads the `SKILL.md` from the skern registry
2. Creates the platform-specific skill directory
3. Copies the `SKILL.md` into the target location
4. The agent runtime discovers the skill on its next invocation

## Supported Platforms

Adapter names match `vercel-labs/skills` agent identifiers wherever possible.

| Platform | Adapter name | Detection |
|----------|--------------|-----------|
| Claude Code | `claude-code` | `~/.claude/` |
| Codex CLI | `codex-cli` | `~/.codex/` or `~/.agents/` |
| OpenCode | `opencode` | `~/.config/opencode/` |
| Cursor | `cursor` | `~/.cursor/` |
| Gemini CLI | `gemini-cli` | `~/.gemini/` |
| GitHub Copilot | `github-copilot` | `~/.copilot/` |
| Windsurf | `windsurf` | `~/.codeium/windsurf/` or `~/.windsurf/` |
| Continue | `continue` | `~/.continue/` |

See [Supported Platforms](/platforms/) for the full path reference.

## Installation Paths

Each platform uses different directories for user-level and project-level skills. Several platforms share `.agents/skills/` as the project directory — see [Shared project directory](#shared-project-directory) below.

## Auto-Detection

Skern auto-detects which platforms are installed on your system by checking each platform's user-level config directory. Use `skern platform list` to see detected platforms:

```sh
skern platform list
```

## Shared project directory

`codex-cli`, `cursor`, `gemini-cli`, and `github-copilot` all use `.agents/skills/` as their project-level skills directory. A skill installed there is visible to every agent that reads from it — that is intentional and matches the conventions in `vercel-labs/skills`.

Two consequences:

- **Detection is per-platform**, not per-directory. The presence of `.agents/skills/` does not by itself indicate which agents are installed; skern looks at each platform's distinct user-level config dir (`~/.cursor`, `~/.gemini`, `~/.copilot`, `~/.codex`) to disambiguate.
- **Capacity is per-directory.** When two platforms share a directory, both adapters see the same installed-skills count. Capacity thresholds protect the directory, not the logical agent — installing 50 skills via `cursor` will register as full capacity for `gemini-cli` too, because the agent will load all of them.
- **One body per skill name.** Because the four adapters write to the same path, you cannot install different content for the same skill name to, say, `codex-cli` and `github-copilot` — the last install wins. Per-platform variants and a per-platform destination override are tracked in [#47](https://github.com/devrimcavusoglu/skern/issues/47) and [#101](https://github.com/devrimcavusoglu/skern/issues/101).

The shared directory is not a skern invention: GitHub Copilot accepts `.github/skills/`, `.claude/skills/`, **and** `.agents/skills/` for project scope ([GitHub docs](https://docs.github.com/en/copilot/concepts/agents/about-agent-skills)), and Codex CLI, Cursor, and Gemini CLI follow the [vercel-labs/skills](https://github.com/vercel-labs/skills#supported-agents) layout.

## One Platform per Invocation

Each `skern skill install` call targets exactly one platform. Agents are expected to specify the platform they are running on — there is no `all` value, and skern does not broadcast skills across platforms automatically.

This design supports the [dynamic skill loading](./registry) model: each agent maintains its own working set of installed skills, sized to its context budget, independent of other platforms.

To deploy a skill across several platforms, loop the call:

```sh
for p in claude-code codex-cli opencode cursor gemini-cli; do
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

## Adding a Platform

Adapters are declarative: every supported platform is one row in `Specs` in [`internal/platform/spec.go`](https://github.com/devrimcavusoglu/skern/blob/main/internal/platform/spec.go). Adding a new platform takes:

1. A new `Type` constant in `internal/platform/platform.go`.
2. A new `Spec` row giving its name, user-level skills dir, project-level skills dir, and one or more home-relative detection paths.

A single generic `Adapter` struct implements every platform's `Platform` interface from the spec, so no platform-specific Go code is required. The table-driven tests in `internal/platform/platform_test.go` cover every registered spec automatically.
