# Agent Setup

After installing skern, you can enable the **tool-forming loop** — a pattern where agents search for existing skills before creating new ones, keeping your skill set deduplicated and organized.

## Recommended: `skern init --instructions`

The fastest way to make your agent use skern correctly is to let `skern init` write its own usage snippet into your agent's instruction file:

```sh
skern init --instructions
```

This auto-discovers `AGENTS.md`, `CLAUDE.md`, and `.claude/CLAUDE.md` under the current directory and writes a marker-wrapped block:

```markdown
<!-- skern:instructions:start -->
## Skern (skill management)

Use `skern` for ALL skill-related tasks — discovery, reading, creation, edits, removal. Run `skern --help` for the full command surface.
<!-- skern:instructions:end -->
```

Re-running `skern init --instructions` **updates the block in place** rather than appending a duplicate.

### Add the search-before-create workflow

```sh
skern init --instructions --tool-forming-loop
```

This appends a section that tells the agent to:

1. Run `skern skill search <topic>` first.
2. If a match scores ≥ 0.6, read it via `skern skill show <name>` and follow its body.
3. Only create a new skill (`skern skill create <name>`) when nothing matches.

### Other ways to drive `init`

| Need | Flag |
|------|------|
| Write to a specific file (skips auto-discovery) | `--target ./MY_AGENT.md` (repeatable) |
| Print the snippet to stdout instead of writing | `--print-instructions` |
| Run non-interactively (CI, scripts) | Pass `--instructions` or `--no-instructions` — any instruction flag disables the prompts, and they never fire without a TTY anyway |
| Skip the snippet entirely, no prompt | `--no-instructions` |

When run on a TTY with no instruction flag at all, `skern init` asks whether to write the snippet and whether to include the tool-forming loop. Both default to **No**. Any instruction flag silences both questions (so `skern init --instructions` in a terminal writes the snippet and returns — it does not stop to ask about the tool-forming loop). When stdin is not a TTY or `--json` is set, skern never prompts — both answers resolve to No. Installers and CI that don't want the snippet should say so explicitly with `skern init --no-instructions`, which writes nothing and never prompts regardless of TTY state.

## How the Loop Works

When an agent encounters a recurring task:

1. The agent reads the snippet from `AGENTS.md`/`CLAUDE.md`.
2. It runs `skern skill search <query>` to check for existing skills.
3. If a matching skill exists, the agent reuses it.
4. If no match, the agent creates a new skill with `skern skill create`.
5. The new skill is immediately available for future sessions.

## JSON Mode for Agents

Every skern command supports `--json` for machine-readable output:

```sh
skern skill list --json
skern skill list --with-platforms --json   # adds installed_on per skill
skern skill search "review" --json
skern skill show code-review --json
skern skill install code-review --platform claude-code --json
skern skill install --tag workflow --platform claude-code --json   # group install by tag
```

The `install`/`uninstall` JSON envelope carries a `skills[]` array (one entry per skill) and a top-level `capacity` block (`installed`, `threshold`, `headroom`, `over_budget`). Agents can react to capacity pressure without an extra query.

## Manual Setup (Without `init`)

If you'd rather hand-author the agent instructions, the bare minimum is:

```markdown
## Skill Management

- Use `skern` to manage reusable skills.
- Before creating a new skill, run `skern skill search <query>`.
- Use `skern skill recommend <query>` to get an explicit reuse-vs-extend-vs-create suggestion.
- Validate after creation: `skern skill validate <name>`.
- Install to your current platform, e.g. `skern skill install <name> --platform claude-code`. Each call targets one platform.
- Multiple skills install in one call: `skern skill install <a> <b> <c> --platform <p>`.
- Watch the `capacity` block in install/uninstall JSON output. Uninstall stale skills before adding new ones when capacity is tight.
```

`skern init --instructions --print-instructions` will print exactly the snippet skern uses if you want to copy/customize it manually.
