# Manual Agent Test Harness

The `tests/manual/` directory contains 10 scenarios that test how AI agents interact with skern — discovery, command chaining, JSON parsing, error handling, and dedup reasoning.

Run these before releases to verify agent compatibility.

## Setup

```sh
# 1. Set up isolated test environments in /tmp
make test-manual-setup
```

## Running Scenarios

```sh
# 2. Test each scenario with your AI agent
cd /tmp/skern-manual-tests/01-fresh-project
cat PROMPT.md      # Read what to ask the agent
cat EXPECTED.md    # Read the pass criteria
# Open your AI agent and run the prompt
# Repeat for each scenario (01 through 10)
```

Each scenario directory contains:

| File | Purpose |
|------|---------|
| `AGENTS.md` | Agent instructions |
| `PROMPT.md` | What to ask the agent |
| `EXPECTED.md` | Pass/fail checklist |

## Generating Reports

```sh
# 3. Generate a markdown report (interactive pass/fail checklist)
make test-manual-report
```

## Cleanup

```sh
# 4. Clean up temp dirs and platform markers
make test-manual-teardown
```

## Scenario Coverage

The 10 scenarios cover:

1. Fresh project discovery
2. Skill creation workflows
3. Cross-platform installation
4. JSON output parsing
5. Overlap detection handling
6. Validation error recovery
7. Search and recommendation
8. Multi-skill management
9. Error handling and edge cases
10. Deduplication reasoning

See [`tests/manual/README.md`](https://github.com/devrimcavusoglu/skern/blob/main/tests/manual/README.md) for full details.
