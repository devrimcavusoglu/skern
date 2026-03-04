# PR Manual Testing

Some skern changes — CLI behavior, skill management, platform adapters, registry logic — require manual verification that CI cannot provide. This page documents the lightweight manual test verification step required on PRs before merge.

## When Is It Required?

Manual test verification is required for PRs that touch:

- CLI commands or flags (`internal/cli/`)
- Skill management logic (`internal/skill/`, `internal/registry/`)
- Platform adapters (`internal/platform/`)
- Overlap detection (`internal/overlap/`)
- Output formatting (`internal/output/`)

### Exemptions

The following PRs are exempt from manual test verification:

- Documentation-only changes (`docs/`, `*.md` outside `internal/`)
- CI/CD configuration changes (`.github/workflows/`)
- Trivially mechanical changes (typo fixes, import reordering, comment edits)

Mark exempt PRs with a comment noting the exemption reason, then check the template checkbox.

## Comment Format

A maintainer posts a PR comment with the following table before merge:

````markdown
### Manual Test Verification

| # | Scenario | Result | Notes |
|---|----------|--------|-------|
| 1 | `skern skill list` shows expected skills | :white_check_mark: | |
| 2 | `skern skill install <name> --platform claude-code` works | :white_check_mark: | |
| 3 | `skern platform status` reflects installed skill | :white_check_mark: | |
| 4 | *<add rows relevant to the PR>* | | |

**Tested at:** `<commit SHA>`
**Platform:** `<OS / shell>`
````

Adapt the rows to the specific PR — only test what the PR changes. Keep it brief.

## Rules

1. **Pass / Fail / Skip** — use :white_check_mark: (pass), :x: (fail), or :fast_forward: (skip with reason in Notes).
2. **All rows must pass** before merge. If any row fails, the PR must be updated and re-tested.
3. **Pin the commit SHA** — record which commit was tested. If new commits are pushed after the verification comment, the test must be re-run.
4. **Brevity** — test only what the PR changes. Don't re-run the full agent test harness (that's a separate concern, see below).

## Labels

Two GitHub labels track manual test status:

| Label | Color | Meaning |
|-------|-------|---------|
| `needs-manual-test` | yellow | PR requires manual test verification before merge |
| `manual-test-verified` | green | Manual test verification has been posted and passes |

Maintainers apply `needs-manual-test` when opening or reviewing a non-exempt PR, and replace it with `manual-test-verified` after a passing verification comment.

## Relationship to the Agent Test Harness

The [agent test harness](/contributing/manual-testing) (`tests/manual/`) is a separate, more comprehensive concern — it covers 10 end-to-end scenarios for pre-release validation. PR manual testing is lighter: it verifies only the specific behavior changed by the PR. The two complement each other but are independent processes.
