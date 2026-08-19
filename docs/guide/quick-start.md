# Quick Start

This walkthrough creates a skill and installs it to a platform in under a minute.

## 1. Initialize the Skill Registry

```sh
skern init
```

This creates a `.skern/skills/` directory in your project for project-scoped skills. It is idempotent — safe to re-run.

To also drop a skern usage snippet into your agent's instruction file (`AGENTS.md`, `CLAUDE.md`, or `.claude/CLAUDE.md`) so the agent uses skern for skill management:

```sh
skern init --instructions
```

Add `--tool-forming-loop` to include the search-before-create workflow section. The snippet is wrapped in `<!-- skern:instructions:start -->` / `<!-- skern:instructions:end -->` markers, so re-running updates the block in place. See [`skern init`](/reference/commands#skern-init) for all flags.

## 2. Create a Skill

```sh
skern skill create code-review --description "Use when reviewing pull requests for style and correctness"
```

This scaffolds a `SKILL.md` file in the registry with the given name and description. Overlap detection runs automatically — see [Overlap Detection](/reference/overlap-detection).

## 3. Install to a Platform

Install to Claude Code:

```sh
skern skill install code-review --platform claude-code
```

Each invocation targets exactly one platform — `--platform all` is not accepted. To deploy across multiple platforms, loop the call:

```sh
for p in claude-code codex-cli opencode cursor gemini-cli; do
  skern skill install code-review --platform "$p"
done
```

Multiple skill names can be installed in a single call, or a whole tagged group at once:

```sh
skern skill install code-review test-runner deploy-checker --platform claude-code
skern skill install --tag workflow --platform claude-code     # every registry skill tagged "workflow"
```

The response includes a `capacity` block reporting the installed-skill count, threshold, and remaining headroom for the platform. Pass `--enforce-budget` to refuse the install when the count would exceed the threshold.

## 4. List Installed Skills

```sh
skern skill list
```

Add `--with-platforms` to surface per-skill installation state across detected platforms:

```sh
skern skill list --with-platforms
```

## 5. Search Before Creating

Avoid duplicates by searching first:

```sh
skern skill search "review"
skern skill recommend "review pull requests"   # decides reuse vs. extend vs. create
```

## 6. Diff Before Editing

Compare a registry skill against its installed copy on a platform, or two registry skills:

```sh
skern skill diff code-review --platform claude-code
skern skill diff code-review code-review-strict
```

## What's Next?

- [Agent Setup](/guide/agent-setup) — enable agents to manage skills automatically
- [CLI Reference](/reference/commands) — full command documentation
- [Skill Format](/concepts/skill-format) — understand the `SKILL.md` structure
- [Writing Skills](/writing-skills) — write skills agents can find and reuse
