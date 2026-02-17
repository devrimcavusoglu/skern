#!/usr/bin/env bash
set -euo pipefail

# Scribe Manual Test Scenarios — Setup Script
# Creates isolated test environments in /tmp/scribe-manual-tests/
# with pre-populated skills and git repos for each scenario.

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
SCENARIOS_DIR="$SCRIPT_DIR/scenarios"
TEST_ROOT="/tmp/scribe-manual-tests"
CREATED_DIRS_FILE="$TEST_ROOT/.created_dirs"

echo "=== Scribe Manual Test Setup ==="
echo ""

# --- Preflight checks ---

if ! command -v scribe &>/dev/null; then
  echo "ERROR: scribe not found in PATH. Run 'make build' and add to PATH first."
  exit 1
fi

SCRIBE_VERSION="$(scribe version 2>/dev/null | head -1)"
echo "  scribe: $SCRIBE_VERSION"
echo "  test root: $TEST_ROOT"
echo ""

# --- Clean previous runs ---

if [ -d "$TEST_ROOT" ]; then
  echo "  Removing previous test root..."
  rm -rf "$TEST_ROOT"
fi
mkdir -p "$TEST_ROOT"

# --- Platform marker dirs ---
# Create platform detection dirs if they don't exist (for 3-platform scenarios).
# Track what we create so teardown can clean up.

touch "$CREATED_DIRS_FILE"

ensure_dir() {
  local dir="$1"
  if [ ! -d "$dir" ]; then
    mkdir -p "$dir"
    echo "$dir" >> "$CREATED_DIRS_FILE"
    echo "  Created platform marker: $dir"
  fi
}

ensure_dir "$HOME/.agents"
ensure_dir "$HOME/.config/opencode"

# --- Helpers ---

# Write a SKILL.md directly into .scribe/skills/<name>/ (project scope)
write_skill() {
  local base_dir="$1"
  local name="$2"
  local description="$3"
  local body="${4:-}"

  local skill_dir="$base_dir/.scribe/skills/$name"
  mkdir -p "$skill_dir"

  cat > "$skill_dir/SKILL.md" <<SKILLEOF
---
name: $name
description: $description
version: 0.1.0
author:
  name: test-harness
  type: human
---

$body
SKILLEOF
}

# Initialize a git repo in the given directory
init_git() {
  local dir="$1"
  (
    cd "$dir"
    git init -q
    git add -A
    git commit -q -m "Initial commit (test scenario)" --allow-empty
  )
}

# Copy scenario definition files (AGENTS.md, PROMPT.md, EXPECTED.md) into test dir
copy_scenario_files() {
  local scenario_name="$1"
  local dest_dir="$2"
  local src_dir="$SCENARIOS_DIR/$scenario_name"

  if [ -f "$src_dir/AGENTS.md" ]; then
    cp "$src_dir/AGENTS.md" "$dest_dir/AGENTS.md"
  fi
  if [ -f "$src_dir/PROMPT.md" ]; then
    cp "$src_dir/PROMPT.md" "$dest_dir/PROMPT.md"
  fi
  if [ -f "$src_dir/EXPECTED.md" ]; then
    cp "$src_dir/EXPECTED.md" "$dest_dir/EXPECTED.md"
  fi
}

# --- Scenario setup ---

setup_scenario() {
  local num="$1"
  local name="$2"
  local dir="$TEST_ROOT/$num-$name"
  mkdir -p "$dir"
  copy_scenario_files "$num-$name" "$dir"
  echo "$dir"
}

# Scenario 01: Fresh Project — empty registry
echo "Setting up 01-fresh-project..."
DIR="$(setup_scenario 01 fresh-project)"
(cd "$DIR" && scribe init --quiet 2>/dev/null)
init_git "$DIR"

# Scenario 02: Existing Skills — 3 pre-populated skills
echo "Setting up 02-existing-skills..."
DIR="$(setup_scenario 02 existing-skills)"
(cd "$DIR" && scribe init --quiet 2>/dev/null)
write_skill "$DIR" "go-formatter" "Formats Go source files using gofmt and goimports"
write_skill "$DIR" "python-linter" "Lints Python code with ruff and reports issues"
write_skill "$DIR" "markdown-toc" "Generates table of contents for markdown files"
init_git "$DIR"

# Scenario 03: Overlap Detection — 2 skills with overlapping descriptions
echo "Setting up 03-overlap-detection..."
DIR="$(setup_scenario 03 overlap-detection)"
(cd "$DIR" && scribe init --quiet 2>/dev/null)
write_skill "$DIR" "code-review" "Review code changes and suggest code improvements"
write_skill "$DIR" "lint-python" "Lint Python source code and report lint errors"
init_git "$DIR"

# Scenario 04: Multi-Platform Install — 1 skill, all 3 platform dirs
echo "Setting up 04-multi-platform-install..."
DIR="$(setup_scenario 04 multi-platform-install)"
(cd "$DIR" && scribe init --quiet 2>/dev/null)
write_skill "$DIR" "deploy-helper" "Assists with deployment steps and checklists"
mkdir -p "$DIR/.claude" "$DIR/.agents" "$DIR/.opencode"
init_git "$DIR"

# Scenario 05: Full Lifecycle JSON — empty registry, .claude/ present
echo "Setting up 05-full-lifecycle-json..."
DIR="$(setup_scenario 05 full-lifecycle-json)"
(cd "$DIR" && scribe init --quiet 2>/dev/null)
mkdir -p "$DIR/.claude"
init_git "$DIR"

# Scenario 06: Error Recovery — 1 skill, .claude/ present
echo "Setting up 06-error-recovery..."
DIR="$(setup_scenario 06 error-recovery)"
(cd "$DIR" && scribe init --quiet 2>/dev/null)
write_skill "$DIR" "test-runner" "Run test suites and report test results for the project"
mkdir -p "$DIR/.claude"
init_git "$DIR"

# Scenario 07: Scoped Skill Management — empty registry
echo "Setting up 07-scoped-skill-management..."
DIR="$(setup_scenario 07 scoped-skill-management)"
(cd "$DIR" && scribe init --quiet 2>/dev/null)
init_git "$DIR"

# Scenario 08: Deduplication Advisory — 5 overlapping skills
echo "Setting up 08-deduplication-advisory..."
DIR="$(setup_scenario 08 deduplication-advisory)"
(cd "$DIR" && scribe init --quiet 2>/dev/null)
write_skill "$DIR" "test-runner" "Run test suites and report test results for the project"
write_skill "$DIR" "run-tests" "Run test suites and report results across the project"
write_skill "$DIR" "test-runner-v2" "Run test suites and report test results with coverage"
write_skill "$DIR" "code-review" "Review code changes and suggest code improvements"
write_skill "$DIR" "code-reviewer" "Review code changes and provide code improvement suggestions"
init_git "$DIR"

# Scenario 09: Template Skills — empty registry, template files provided
echo "Setting up 09-template-skills..."
DIR="$(setup_scenario 09 template-skills)"
(cd "$DIR" && scribe init --quiet 2>/dev/null)
# Copy template files
if [ -d "$SCENARIOS_DIR/09-template-skills/templates" ]; then
  mkdir -p "$DIR/templates"
  cp "$SCENARIOS_DIR/09-template-skills/templates/"*.md "$DIR/templates/"
fi
init_git "$DIR"

# Scenario 10: Platform Status Matrix — 3 skills, partial installs
echo "Setting up 10-platform-status-matrix..."
DIR="$(setup_scenario 10 platform-status-matrix)"
(cd "$DIR" && scribe init --quiet 2>/dev/null)
write_skill "$DIR" "go-formatter" "Formats Go source files using gofmt and goimports"
write_skill "$DIR" "db-migrate" "Run database migrations and track schema changes"
write_skill "$DIR" "api-docs" "Generate API documentation from source code annotations"
mkdir -p "$DIR/.claude" "$DIR/.agents" "$DIR/.opencode"
# Partial installs: go-formatter -> claude-code + codex-cli, db-migrate -> claude-code only
(cd "$DIR" && scribe skill install go-formatter --platform claude-code --scope project --quiet 2>/dev/null || true)
(cd "$DIR" && scribe skill install go-formatter --platform codex-cli --scope project --quiet 2>/dev/null || true)
(cd "$DIR" && scribe skill install db-migrate --platform claude-code --scope project --quiet 2>/dev/null || true)
init_git "$DIR"

# --- Summary ---

echo ""
echo "=== Setup Complete ==="
echo ""
echo "Scenarios created in: $TEST_ROOT"
echo ""
ls -1d "$TEST_ROOT"/*/  2>/dev/null | while read -r d; do
  echo "  $(basename "$d")/"
done
echo ""
echo "Next steps:"
echo "  1. cd into a scenario directory"
echo "  2. Read PROMPT.md for the agent prompt"
echo "  3. Read EXPECTED.md for the pass criteria"
echo "  4. Open your AI agent and test"
echo "  5. Run 'make test-manual-report' when done"
