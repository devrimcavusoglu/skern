# Quick Start

This walkthrough creates a skill and installs it to a platform in under a minute.

## 1. Initialize the Skill Registry

```sh
skern init
```

This creates a `.skern/` directory in your project for project-scoped skills.

## 2. Create a Skill

```sh
skern skill create code-review --description "Review PRs for style and correctness"
```

This scaffolds a `SKILL.md` file in the registry with the given name and description.

## 3. Install to a Platform

Install to Claude Code:

```sh
skern skill install code-review --platform claude-code
```

Each invocation targets one platform. To install to multiple platforms, loop the call:

```sh
for p in claude-code codex-cli opencode; do
  skern skill install code-review --platform "$p"
done
```

You can also install several skills at once:

```sh
skern skill install code-review test-runner deploy-checker --platform claude-code
```

## 4. List Installed Skills

```sh
skern skill list
```

## 5. Search Before Creating

Avoid duplicates by searching first:

```sh
skern skill search "review"
```

## What's Next?

- [Agent Setup](/guide/agent-setup) — enable agents to manage skills automatically
- [CLI Reference](/reference/commands) — full command documentation
- [Skill Format](/concepts/skill-format) — understand the `SKILL.md` structure
