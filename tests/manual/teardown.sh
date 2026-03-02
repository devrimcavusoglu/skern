#!/usr/bin/env bash
set -euo pipefail

# Skern Manual Test Scenarios — Teardown Script
# Removes home directory markers created by setup.sh and cleans up test scenarios.

TEST_ROOT="/tmp/skern-manual-tests"
CREATED_DIRS_FILE="$TEST_ROOT/.created_dirs"

echo "=== Skern Manual Test Teardown ==="

# Remove home directory markers that were created by setup
if [ -f "$CREATED_DIRS_FILE" ]; then
  while IFS= read -r dir; do
    [ -z "$dir" ] && continue
    if [ -d "$dir" ]; then
      # Only remove if the directory is empty (we created it, don't delete user content)
      if [ -z "$(ls -A "$dir" 2>/dev/null)" ]; then
        rmdir "$dir"
        echo "  Removed empty marker: $dir"
      else
        echo "  Skipped (not empty): $dir"
      fi
    fi
  done < "$CREATED_DIRS_FILE"
else
  echo "  No .created_dirs file found — nothing to clean in home directory."
fi

# Remove test root
if [ -d "$TEST_ROOT" ]; then
  rm -rf "$TEST_ROOT"
  echo "  Removed test root: $TEST_ROOT"
else
  echo "  Test root already removed: $TEST_ROOT"
fi

echo ""
echo "=== Teardown Complete ==="
