# Architecture

Skern separates concerns into four layers:

```
Skill Author --> skern --> Registry --> Agent Runtime
                  |
          +-------+-------+
          |       |       |
        Claude  Codex   OpenCode
         Code    CLI
```

## Layers

### Skill Definition

Metadata and behavior live in a single `SKILL.md` file. The file uses YAML frontmatter for structured fields (name, description, author, version, allowed-tools) and markdown body for the skill's instructions.

See [Skill Format](/concepts/skill-format) for the full specification.

### Skill Registry

Skills are stored as directories containing a `SKILL.md` file. Two scopes are available:

- **Project scope** — `.skern/skills/<name>/` in the current project
- **User scope** — `~/.skern/skills/<name>/` for system-wide skills

See [Registry](/concepts/registry) for details.

### Validation

Before a skill enters the registry, skern validates it against the [Agent Skills](https://agentskills.io) specification. This includes name format checks, description requirements, and overlap detection against existing skills.

See [Validation](/reference/validation) and [Overlap Detection](/reference/overlap-detection) for rules and thresholds.

### Platform Adapters

Each supported platform has an adapter that knows where to install skills. Adapters copy the `SKILL.md` to the platform-specific directory, making the skill immediately available to the agent runtime.

See [Platform Adapters](/concepts/platform-adapters) for how adapters work.
