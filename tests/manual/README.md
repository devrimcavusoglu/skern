# Manual Agent Test Harness

Tests how AI agents interact with scribe — discovery, command chaining, JSON parsing, error handling, and deduplication reasoning. Run before releases and attach the report to PRs.

## Prerequisites

- `scribe` built and in PATH (`make build && export PATH="$PWD:$PATH"`)
- An AI agent (Claude Code, Codex CLI, OpenCode, etc.)

## Workflow

```sh
# 1. Set up all 10 scenarios in /tmp
make test-manual-setup

# 2. Test each scenario with your agent
cd /tmp/scribe-manual-tests/01-fresh-project
cat PROMPT.md      # Read what to ask the agent
cat EXPECTED.md    # Read the pass criteria
# Open your AI agent and run the prompt

# 3. Generate report (interactive pass/fail checklist)
make test-manual-report

# 4. Clean up
make test-manual-teardown
```

## Scenarios

| # | Name | What it tests |
|---|------|---------------|
| 01 | fresh-project | Agent discovers scribe, searches before creating |
| 02 | existing-skills | Agent finds and reuses existing skill |
| 03 | overlap-detection | Agent handles overlap warnings |
| 04 | multi-platform-install | Install/uninstall across platforms |
| 05 | full-lifecycle-json | End-to-end with `--json` on every command |
| 06 | error-recovery | Agent handles errors without getting stuck |
| 07 | scoped-skill-management | Project scope create/list/show/validate |
| 08 | deduplication-advisory | Agent audits and advises on duplicate skills |
| 09 | template-skills | Skill creation with `--from-template` |
| 10 | platform-status-matrix | Fill gaps in skill-platform matrix |

## Directory layout

```
tests/manual/
  scenarios/          # Source-controlled scenario definitions
    01-fresh-project/
      AGENTS.md       # Agent instructions (copied into test dir)
      PROMPT.md       # What to ask the agent
      EXPECTED.md     # Pass/fail checklist (parsed by report.sh)
    ...
  setup.sh            # Creates /tmp/scribe-manual-tests/ with pre-populated state
  teardown.sh         # Removes temp dirs + platform markers
  report.sh           # Interactive checklist -> markdown report
  reports/            # Generated reports (gitignored)
```

## Notes

- Setup creates `~/.agents/` and `~/.config/opencode/` if missing (for 3-platform detection). Teardown removes them only if still empty.
- All pre-populated skills use project scope (`.scribe/skills/` inside each test dir).
- Reports are per-run artifacts saved to `tests/manual/reports/` and gitignored.
