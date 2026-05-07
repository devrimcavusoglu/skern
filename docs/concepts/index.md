# Architecture

Skern separates concerns into four layers:

```
Skill Author --> skern --> Registry --> Agent Runtime
                  |
        +---------+---------+---------+---------+
        |     |     |     |      |       |     |
      Claude Codex Open Cursor Gemini Copilot Continue / Windsurf / ...
       Code  CLI  Code        CLI
```

## Layers

### Skill Definition

Metadata and behavior live in a single `SKILL.md` file. The file uses YAML frontmatter for structured fields (name, description, author, version, tags, allowed-tools) and markdown body for the skill's instructions.

See [Skill Format](/concepts/skill-format) for the full specification.

### Skill Registry

Skills are stored as directories containing a `SKILL.md` file. Two scopes are available:

- **Project scope** — `.skern/skills/<name>/` in the current project
- **User scope** — `~/.skern/skills/<name>/` for system-wide skills

See [Registry](/concepts/registry) for details.

### Validation

Before a skill enters the registry, skern validates it against the [Agent Skills](https://agentskills.io) specification. This includes name format checks, description requirements, folder integrity (referenced files exist), and overlap detection against existing skills. Validation also surfaces stylistic **hints** (short body, vague description, missing trigger prefix) that don't affect validity.

See [Validation](/reference/validation) and [Overlap Detection](/reference/overlap-detection) for rules and thresholds.

### Platform Adapters

Each supported platform has an adapter that knows where to install skills. Adapters copy the skill directory (`SKILL.md` plus any sibling files and subdirectories) to the platform-specific location, making the skill immediately available to the agent runtime.

Adapters are **declarative**: every supported platform is one row in `Specs` in [`internal/platform/spec.go`](https://github.com/devrimcavusoglu/skern/blob/main/internal/platform/spec.go). A single generic `Adapter` struct implements every platform's interface from the spec — no per-platform Go file is required.

See [Platform Adapters](/concepts/platform-adapters) for how adapters work and the full path reference.

## Dynamic Skill Loading

Skern is not just a one-shot installer. The `install` and `uninstall` commands return a `capacity` block (installed count, per-scope threshold, headroom, over-budget flag) so agents can size their working set to a context budget. Pass `--enforce-budget` to refuse installs that would exceed the threshold.

See the [Quick Start](/guide/quick-start#3-install-to-a-platform) for the workflow.
