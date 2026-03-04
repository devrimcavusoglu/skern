# Registry

The skill registry is where skern stores skill definitions. Skills are organized as directories, each containing a `SKILL.md` file.

## Scopes

### Project Scope

Project-scoped skills live in `.skern/skills/` within your project directory. These are specific to the project and can be checked into version control.

```
.skern/
  skills/
    code-review/
      SKILL.md
    deploy/
      SKILL.md
      scripts/
        deploy.sh
      config/
        targets.json
```

Initialize the project registry with:

```sh
skern init
```

### User Scope

User-scoped skills live in `~/.skern/skills/` and are available across all projects on your machine.

```
~/.skern/
  skills/
    global-lint/
      SKILL.md
      scripts/
        lint-rules.js
```

### Scope Selection

Most commands accept a `--scope` flag:

| Value | Description |
|-------|-------------|
| `project` | Project-level registry only (default for most commands) |
| `user` | User-level registry only |
| `all` | Both registries |

## Skill Count Warnings

Registries have recommended size limits:

- **Project scope** — warns at > 20 skills
- **User scope** — warns at > 50 skills

These are soft limits — skern will continue to work, but warns to encourage organization.

## Registry vs Platform

The registry is where skills are **defined**. Platforms are where skills are **installed**. A skill must exist in the registry before it can be installed to a platform.

```
Registry (.skern/skills/) --install--> Platform (.claude/skills/)
```

See [Platform Adapters](/concepts/platform-adapters) for installation details.
