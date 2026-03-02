#!/bin/bash
# Smoke and E2E tests for the skern binary.
# Usage: bash scripts/smoke_test.sh [path/to/skern]
#
# Tests run against a built binary, using temp directories.
# No network calls, no real platform installations.

set -e

SKERN="${1:-./skern}"

if [ ! -x "$SKERN" ]; then
    echo "ERROR: skern binary not found at '$SKERN'. Run 'make build' first."
    exit 1
fi

# Resolve to absolute path
SKERN="$(cd "$(dirname "$SKERN")" && pwd)/$(basename "$SKERN")"

PASS=0
FAIL=0
TESTS_RUN=0

# --- test helpers ---

assert_eq() {
    local label="$1" expected="$2" actual="$3"
    TESTS_RUN=$((TESTS_RUN + 1))
    if [ "$expected" = "$actual" ]; then
        PASS=$((PASS + 1))
    else
        FAIL=$((FAIL + 1))
        printf '  FAIL: %s\n    expected: %s\n    actual:   %s\n' "$label" "$expected" "$actual"
    fi
}

assert_contains() {
    local label="$1" haystack="$2" needle="$3"
    TESTS_RUN=$((TESTS_RUN + 1))
    # Strip whitespace for JSON-friendly matching
    local compact
    compact="$(echo "$haystack" | tr -d '[:space:]')"
    local compact_needle
    compact_needle="$(echo "$needle" | tr -d '[:space:]')"
    if echo "$compact" | grep -qF "$compact_needle"; then
        PASS=$((PASS + 1))
    else
        FAIL=$((FAIL + 1))
        printf '  FAIL: %s\n    expected to contain: %s\n    actual: %s\n' "$label" "$needle" "$haystack"
    fi
}

assert_not_contains() {
    local label="$1" haystack="$2" needle="$3"
    TESTS_RUN=$((TESTS_RUN + 1))
    local compact
    compact="$(echo "$haystack" | tr -d '[:space:]')"
    local compact_needle
    compact_needle="$(echo "$needle" | tr -d '[:space:]')"
    if ! echo "$compact" | grep -qF "$compact_needle"; then
        PASS=$((PASS + 1))
    else
        FAIL=$((FAIL + 1))
        printf '  FAIL: %s\n    expected NOT to contain: %s\n    actual: %s\n' "$label" "$needle" "$haystack"
    fi
}

assert_exit_code() {
    local label="$1" expected="$2" actual="$3"
    TESTS_RUN=$((TESTS_RUN + 1))
    if [ "$expected" = "$actual" ]; then
        PASS=$((PASS + 1))
    else
        FAIL=$((FAIL + 1))
        printf '  FAIL: %s\n    expected exit code: %s\n    actual exit code:   %s\n' "$label" "$expected" "$actual"
    fi
}

assert_file_exists() {
    local label="$1" path="$2"
    TESTS_RUN=$((TESTS_RUN + 1))
    if [ -f "$path" ]; then
        PASS=$((PASS + 1))
    else
        FAIL=$((FAIL + 1))
        printf '  FAIL: %s\n    file does not exist: %s\n' "$label" "$path"
    fi
}

assert_dir_exists() {
    local label="$1" path="$2"
    TESTS_RUN=$((TESTS_RUN + 1))
    if [ -d "$path" ]; then
        PASS=$((PASS + 1))
    else
        FAIL=$((FAIL + 1))
        printf '  FAIL: %s\n    directory does not exist: %s\n' "$label" "$path"
    fi
}

assert_file_not_exists() {
    local label="$1" path="$2"
    TESTS_RUN=$((TESTS_RUN + 1))
    if [ ! -f "$path" ]; then
        PASS=$((PASS + 1))
    else
        FAIL=$((FAIL + 1))
        printf '  FAIL: %s\n    file should not exist: %s\n' "$label" "$path"
    fi
}

run_test() {
    local name="$1"
    printf 'Running %s ... ' "$name"
    shift
    "$@"
    printf 'ok\n'
}

# --- setup / teardown ---

setup_env() {
    TMPDIR_ROOT="$(mktemp -d)"
    export HOME="$TMPDIR_ROOT/home"
    mkdir -p "$HOME"
    PROJECT_DIR="$TMPDIR_ROOT/project"
    mkdir -p "$PROJECT_DIR"

    # Create platform detection dirs so skern detects them
    mkdir -p "$HOME/.claude"
    mkdir -p "$HOME/.codex"
    mkdir -p "$HOME/.config/opencode"
}

teardown_env() {
    rm -rf "$TMPDIR_ROOT"
}

# ============================================================
# Smoke tests — basic CLI behavior
# ============================================================

test_version_output() {
    setup_env
    local out
    out="$($SKERN version 2>&1)"
    assert_contains "version output contains 'skern'" "$out" "skern"
    teardown_env
}

test_help_output() {
    setup_env
    local out
    out="$($SKERN --help 2>&1)"
    assert_contains "help shows 'skill'" "$out" "skill"
    assert_contains "help shows 'platform'" "$out" "platform"
    assert_contains "help shows 'init'" "$out" "init"
    assert_contains "help shows 'version'" "$out" "version"
    teardown_env
}

test_version_json() {
    setup_env
    local out
    out="$($SKERN version --json 2>&1)"
    assert_contains "version json has version key" "$out" '"version"'
    assert_contains "version json has commit key" "$out" '"commit"'
    teardown_env
}

test_unknown_command_fails() {
    setup_env
    local rc=0
    $SKERN notacommand >/dev/null 2>&1 || rc=$?
    assert_exit_code "unknown command exits non-zero" "1" "$rc"
    teardown_env
}

test_skill_help() {
    setup_env
    local out
    out="$($SKERN skill --help 2>&1)"
    assert_contains "skill help shows create" "$out" "create"
    assert_contains "skill help shows list" "$out" "list"
    assert_contains "skill help shows search" "$out" "search"
    assert_contains "skill help shows validate" "$out" "validate"
    assert_contains "skill help shows install" "$out" "install"
    teardown_env
}

test_platform_help() {
    setup_env
    local out
    out="$($SKERN platform --help 2>&1)"
    assert_contains "platform help shows list" "$out" "list"
    assert_contains "platform help shows status" "$out" "status"
    teardown_env
}

# ============================================================
# Init tests
# ============================================================

test_init_creates_directory() {
    setup_env
    local out
    out="$(cd "$PROJECT_DIR" && $SKERN init 2>&1)"
    assert_dir_exists "init creates .skern/skills dir" "$PROJECT_DIR/.skern/skills"
    teardown_env
}

test_init_json() {
    setup_env
    local out
    out="$(cd "$PROJECT_DIR" && $SKERN init --json 2>&1)"
    assert_contains "init json has path" "$out" '"path"'
    teardown_env
}

# ============================================================
# Skill lifecycle — E2E
# ============================================================

test_skill_create() {
    setup_env
    local out
    out="$($SKERN skill create test-smoke --description "A smoke test skill" --json 2>&1)"
    assert_contains "create returns skill name" "$out" '"name":"test-smoke"'
    assert_file_exists "SKILL.md created in registry" "$HOME/.skern/skills/test-smoke/SKILL.md"
    teardown_env
}

test_skill_create_with_author() {
    setup_env
    local out
    out="$($SKERN skill create authored-skill --description "Skill with author" --author "tester" --author-type human --json 2>&1)"
    assert_contains "create returns skill name" "$out" '"name":"authored-skill"'
    # Verify author in SKILL.md
    local content
    content="$(cat "$HOME/.skern/skills/authored-skill/SKILL.md")"
    assert_contains "SKILL.md contains author name" "$content" "tester"
    teardown_env
}

test_skill_create_invalid_name() {
    setup_env
    local rc=0
    $SKERN skill create "INVALID NAME" --description "bad" >/dev/null 2>&1 || rc=$?
    assert_exit_code "invalid name exits with validation error" "2" "$rc"
    teardown_env
}

test_skill_list_empty() {
    setup_env
    local out
    out="$($SKERN skill list --json 2>&1)"
    assert_contains "empty list has count 0" "$out" '"count":0'
    teardown_env
}

test_skill_show() {
    setup_env
    $SKERN skill create show-me --description "Show test" >/dev/null 2>&1
    local out
    out="$($SKERN skill show show-me --json 2>&1)"
    assert_contains "show returns name" "$out" '"name":"show-me"'
    assert_contains "show returns description" "$out" "Show test"
    teardown_env
}

test_skill_validate() {
    setup_env
    $SKERN skill create valid-skill --description "Validate test" >/dev/null 2>&1
    local out
    out="$($SKERN skill validate valid-skill --json 2>&1)"
    assert_contains "validate returns valid true" "$out" '"valid":true'
    teardown_env
}

test_skill_search() {
    setup_env
    $SKERN skill create search-target --description "Findable skill" >/dev/null 2>&1
    local out
    out="$($SKERN skill search "search" --json 2>&1)"
    assert_contains "search finds the skill" "$out" "search-target"
    teardown_env
}

test_skill_remove() {
    setup_env
    $SKERN skill create remove-me --description "Will be removed" >/dev/null 2>&1
    local out
    out="$($SKERN skill remove remove-me --json 2>&1)"
    assert_contains "remove returns name" "$out" '"name":"remove-me"'
    assert_file_not_exists "SKILL.md removed" "$HOME/.skern/skills/remove-me/SKILL.md"
    teardown_env
}

test_skill_show_nonexistent() {
    setup_env
    local rc=0
    $SKERN skill show nonexistent >/dev/null 2>&1 || rc=$?
    assert_exit_code "show nonexistent exits with error" "1" "$rc"
    teardown_env
}

# ============================================================
# Platform tests
# ============================================================

test_platform_list() {
    setup_env
    local out
    out="$($SKERN platform list --json 2>&1)"
    assert_contains "platform list has platforms key" "$out" '"platforms"'
    teardown_env
}

test_platform_status_empty() {
    setup_env
    local out
    out="$($SKERN platform status --json 2>&1)"
    assert_contains "empty status has status key" "$out" '"status"'
    teardown_env
}

# ============================================================
# Install / Uninstall E2E
# ============================================================

test_skill_install_single_platform() {
    setup_env
    $SKERN skill create install-test --description "Install test" >/dev/null 2>&1
    local out
    out="$($SKERN skill install install-test --platform claude-code --json 2>&1)"
    assert_contains "install returns skill name" "$out" '"skill":"install-test"'
    assert_file_exists "SKILL.md installed to claude-code" "$HOME/.claude/skills/install-test/SKILL.md"
    teardown_env
}

test_skill_install_all_platforms() {
    setup_env
    $SKERN skill create install-all --description "Install all test" >/dev/null 2>&1
    local out
    out="$($SKERN skill install install-all --platform all --json 2>&1)"
    assert_contains "install returns skill name" "$out" '"skill":"install-all"'
    assert_file_exists "installed to claude-code" "$HOME/.claude/skills/install-all/SKILL.md"
    assert_file_exists "installed to codex-cli" "$HOME/.agents/skills/install-all/SKILL.md"
    assert_file_exists "installed to opencode" "$HOME/.config/opencode/skills/install-all/SKILL.md"
    teardown_env
}

test_skill_uninstall() {
    setup_env
    $SKERN skill create uninstall-test --description "Uninstall test" >/dev/null 2>&1
    $SKERN skill install uninstall-test --platform claude-code >/dev/null 2>&1
    local out
    out="$($SKERN skill uninstall uninstall-test --platform claude-code --json 2>&1)"
    assert_contains "uninstall returns skill name" "$out" '"skill":"uninstall-test"'
    assert_file_not_exists "SKILL.md removed from platform" "$HOME/.claude/skills/uninstall-test/SKILL.md"
    teardown_env
}

# ============================================================
# Full lifecycle E2E
# ============================================================

test_full_lifecycle() {
    setup_env

    # 1. Init project
    (cd "$PROJECT_DIR" && $SKERN init --json >/dev/null 2>&1)

    # 2. Create skill
    local out
    out="$($SKERN skill create lifecycle-skill --description "Full lifecycle test" --author "smoke" --author-type agent --json 2>&1)"
    assert_contains "lifecycle: create ok" "$out" '"name":"lifecycle-skill"'

    # 3. Validate
    out="$($SKERN skill validate lifecycle-skill --json 2>&1)"
    assert_contains "lifecycle: validate ok" "$out" '"valid":true'

    # 4. List should show 1 skill
    out="$($SKERN skill list --json 2>&1)"
    assert_contains "lifecycle: list count 1" "$out" '"count":1'

    # 5. Install to all platforms
    out="$($SKERN skill install lifecycle-skill --platform all --json 2>&1)"
    assert_file_exists "lifecycle: claude-code installed" "$HOME/.claude/skills/lifecycle-skill/SKILL.md"
    assert_file_exists "lifecycle: codex-cli installed" "$HOME/.agents/skills/lifecycle-skill/SKILL.md"
    assert_file_exists "lifecycle: opencode installed" "$HOME/.config/opencode/skills/lifecycle-skill/SKILL.md"

    # 6. Platform status should show installed
    out="$($SKERN platform status --json 2>&1)"
    assert_contains "lifecycle: status has skill" "$out" "lifecycle-skill"

    # 7. Uninstall from one
    $SKERN skill uninstall lifecycle-skill --platform claude-code >/dev/null 2>&1
    assert_file_not_exists "lifecycle: claude-code uninstalled" "$HOME/.claude/skills/lifecycle-skill/SKILL.md"
    assert_file_exists "lifecycle: codex-cli still installed" "$HOME/.agents/skills/lifecycle-skill/SKILL.md"

    # 8. Remove from registry
    $SKERN skill remove lifecycle-skill >/dev/null 2>&1
    assert_file_not_exists "lifecycle: removed from registry" "$HOME/.skern/skills/lifecycle-skill/SKILL.md"

    # 9. Show should fail
    local rc=0
    $SKERN skill show lifecycle-skill >/dev/null 2>&1 || rc=$?
    assert_exit_code "lifecycle: show after remove fails" "1" "$rc"

    teardown_env
}

# ============================================================
# Run all tests
# ============================================================

echo "=== Skern Smoke & E2E Tests ==="
echo ""

# Smoke tests
run_test "test_version_output" test_version_output
run_test "test_help_output" test_help_output
run_test "test_version_json" test_version_json
run_test "test_unknown_command_fails" test_unknown_command_fails
run_test "test_skill_help" test_skill_help
run_test "test_platform_help" test_platform_help

# Init
run_test "test_init_creates_directory" test_init_creates_directory
run_test "test_init_json" test_init_json

# Skill CRUD
run_test "test_skill_create" test_skill_create
run_test "test_skill_create_with_author" test_skill_create_with_author
run_test "test_skill_create_invalid_name" test_skill_create_invalid_name
run_test "test_skill_list_empty" test_skill_list_empty
run_test "test_skill_show" test_skill_show
run_test "test_skill_validate" test_skill_validate
run_test "test_skill_search" test_skill_search
run_test "test_skill_remove" test_skill_remove
run_test "test_skill_show_nonexistent" test_skill_show_nonexistent

# Platform
run_test "test_platform_list" test_platform_list
run_test "test_platform_status_empty" test_platform_status_empty

# Install / Uninstall
run_test "test_skill_install_single_platform" test_skill_install_single_platform
run_test "test_skill_install_all_platforms" test_skill_install_all_platforms
run_test "test_skill_uninstall" test_skill_uninstall

# Full lifecycle
run_test "test_full_lifecycle" test_full_lifecycle

# --- summary ---
echo ""
echo "=== Results: $TESTS_RUN tests, $PASS passed, $FAIL failed ==="

if [ "$FAIL" -gt 0 ]; then
    exit 1
fi
