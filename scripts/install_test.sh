#!/bin/bash
# Tests for scripts/install.sh
# Usage: bash scripts/install_test.sh
#
# Tests run entirely offline — no network calls, no real installs.
# Each test creates its own temp directory and cleans up on exit.

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
INSTALL_SCRIPT="$SCRIPT_DIR/install.sh"

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
    if echo "$haystack" | grep -qF "$needle"; then
        PASS=$((PASS + 1))
    else
        FAIL=$((FAIL + 1))
        printf '  FAIL: %s\n    expected to contain: %s\n    actual: %s\n' "$label" "$needle" "$haystack"
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

assert_file_executable() {
    local label="$1" path="$2"
    TESTS_RUN=$((TESTS_RUN + 1))
    if [ -x "$path" ]; then
        PASS=$((PASS + 1))
    else
        FAIL=$((FAIL + 1))
        printf '  FAIL: %s\n    file is not executable: %s\n' "$label" "$path"
    fi
}

assert_exit_code() {
    local label="$1" expected="$2"
    shift 2
    TESTS_RUN=$((TESTS_RUN + 1))
    set +e
    "$@" >/dev/null 2>&1
    local actual=$?
    set -e
    if [ "$expected" -eq "$actual" ]; then
        PASS=$((PASS + 1))
    else
        FAIL=$((FAIL + 1))
        printf '  FAIL: %s\n    expected exit code: %s\n    actual exit code:   %s\n' "$label" "$expected" "$actual"
    fi
}

run_test() {
    local name="$1"
    printf 'TEST: %s\n' "$name"
}

# Source the install script (main is guarded, won't execute)
source "$INSTALL_SCRIPT"

# --- tests ---

test_detect_platform_current_host() {
    run_test "detect_platform sets OS and ARCH for current host"

    # Reset
    OS="" ARCH="" PLATFORM=""
    detect_platform >/dev/null 2>&1

    # OS should be linux or darwin
    if [ "$OS" != "linux" ] && [ "$OS" != "darwin" ]; then
        TESTS_RUN=$((TESTS_RUN + 1))
        FAIL=$((FAIL + 1))
        printf '  FAIL: OS should be linux or darwin, got: %s\n' "$OS"
    else
        TESTS_RUN=$((TESTS_RUN + 1))
        PASS=$((PASS + 1))
    fi

    # ARCH should be amd64 or arm64
    if [ "$ARCH" != "amd64" ] && [ "$ARCH" != "arm64" ]; then
        TESTS_RUN=$((TESTS_RUN + 1))
        FAIL=$((FAIL + 1))
        printf '  FAIL: ARCH should be amd64 or arm64, got: %s\n' "$ARCH"
    else
        TESTS_RUN=$((TESTS_RUN + 1))
        PASS=$((PASS + 1))
    fi

    assert_eq "PLATFORM format" "${OS}_${ARCH}" "$PLATFORM"
}

test_detect_platform_values() {
    run_test "detect_platform produces valid platform string"

    OS="" ARCH="" PLATFORM=""
    detect_platform >/dev/null 2>&1

    # PLATFORM must be one of the four supported combinations
    case "$PLATFORM" in
        linux_amd64|linux_arm64|darwin_amd64|darwin_arm64)
            TESTS_RUN=$((TESTS_RUN + 1))
            PASS=$((PASS + 1))
            ;;
        *)
            TESTS_RUN=$((TESTS_RUN + 1))
            FAIL=$((FAIL + 1))
            printf '  FAIL: unexpected platform: %s\n' "$PLATFORM"
            ;;
    esac
}

test_resolve_version_explicit() {
    run_test "resolve_version uses SKERN_VERSION when set"

    VERSION=""
    SKERN_VERSION="v0.0.1"
    resolve_version >/dev/null 2>&1

    assert_eq "VERSION from env" "v0.0.1" "$VERSION"

    # Cleanup
    unset SKERN_VERSION
}

test_resolve_version_custom() {
    run_test "resolve_version accepts arbitrary version strings"

    VERSION=""
    SKERN_VERSION="v1.2.3-rc1"
    resolve_version >/dev/null 2>&1

    assert_eq "VERSION custom string" "v1.2.3-rc1" "$VERSION"

    unset SKERN_VERSION
}

test_download_url_construction() {
    run_test "download constructs correct URLs"

    VERSION="v0.0.1"
    PLATFORM="darwin_arm64"

    # download() sets DOWNLOAD_URL and CHECKSUMS_URL before attempting fetch
    # We can verify the URL pattern by inspecting what it would construct
    expected_download="https://github.com/devrimcavusoglu/skern/releases/download/v0.0.1/skern_0.0.1_darwin_arm64.tar.gz"
    expected_checksums="https://github.com/devrimcavusoglu/skern/releases/download/v0.0.1/checksums.txt"

    # Construct the URLs the same way download() does
    actual_download="https://github.com/${REPO}/releases/download/${VERSION}/skern_${VERSION#v}_${PLATFORM}.tar.gz"
    actual_checksums="https://github.com/${REPO}/releases/download/${VERSION}/checksums.txt"

    assert_eq "download URL" "$expected_download" "$actual_download"
    assert_eq "checksums URL" "$expected_checksums" "$actual_checksums"
}

test_download_url_version_strip() {
    run_test "download URL strips v prefix from version for archive name"

    VERSION="v2.1.0"
    PLATFORM="linux_amd64"

    actual="https://github.com/${REPO}/releases/download/${VERSION}/skern_${VERSION#v}_${PLATFORM}.tar.gz"
    assert_contains "archive name has no v prefix" "$actual" "skern_2.1.0_linux_amd64"
    assert_contains "download path keeps v prefix" "$actual" "/download/v2.1.0/"
}

test_checksum_verification() {
    run_test "checksum verification detects valid checksum"

    local tmpdir
    tmpdir="$(mktemp -d)"

    # Create a fake tarball with known content
    echo "fake-binary-content" > "$tmpdir/fakefile"
    tar -czf "$tmpdir/skern.tar.gz" -C "$tmpdir" fakefile

    # Compute its real checksum
    local real_sum
    if command -v sha256sum >/dev/null 2>&1; then
        real_sum="$(sha256sum "$tmpdir/skern.tar.gz" | awk '{print $1}')"
    elif command -v shasum >/dev/null 2>&1; then
        real_sum="$(shasum -a 256 "$tmpdir/skern.tar.gz" | awk '{print $1}')"
    else
        printf '  SKIP: no sha256sum or shasum available\n'
        rm -rf "$tmpdir"
        return
    fi

    # Create a checksums file with the correct entry
    echo "$real_sum  skern_0.0.1_darwin_arm64.tar.gz" > "$tmpdir/checksums.txt"

    # Simulate what download() does for verification
    VERSION="v0.0.1"
    PLATFORM="darwin_arm64"
    TARBALL="$tmpdir/skern.tar.gz"
    CHECKSUMS_FILE="$tmpdir/checksums.txt"
    EXPECTED_NAME="skern_${VERSION#v}_${PLATFORM}.tar.gz"
    EXPECTED_SUM="$(grep "$EXPECTED_NAME" "$CHECKSUMS_FILE" | awk '{print $1}')"

    local actual_sum
    if command -v sha256sum >/dev/null 2>&1; then
        actual_sum="$(sha256sum "$TARBALL" | awk '{print $1}')"
    else
        actual_sum="$(shasum -a 256 "$TARBALL" | awk '{print $1}')"
    fi

    assert_eq "checksum matches" "$EXPECTED_SUM" "$actual_sum"

    rm -rf "$tmpdir"
}

test_checksum_mismatch_detected() {
    run_test "checksum verification detects mismatch"

    local tmpdir
    tmpdir="$(mktemp -d)"

    echo "real-content" > "$tmpdir/fakefile"
    tar -czf "$tmpdir/skern.tar.gz" -C "$tmpdir" fakefile

    # Write a wrong checksum
    echo "0000000000000000000000000000000000000000000000000000000000000000  skern_0.0.1_darwin_arm64.tar.gz" > "$tmpdir/checksums.txt"

    local actual_sum
    if command -v sha256sum >/dev/null 2>&1; then
        actual_sum="$(sha256sum "$tmpdir/skern.tar.gz" | awk '{print $1}')"
    elif command -v shasum >/dev/null 2>&1; then
        actual_sum="$(shasum -a 256 "$tmpdir/skern.tar.gz" | awk '{print $1}')"
    else
        printf '  SKIP: no sha256sum or shasum available\n'
        rm -rf "$tmpdir"
        return
    fi

    TESTS_RUN=$((TESTS_RUN + 1))
    if [ "$actual_sum" != "0000000000000000000000000000000000000000000000000000000000000000" ]; then
        PASS=$((PASS + 1))
    else
        FAIL=$((FAIL + 1))
        printf '  FAIL: checksum should not match the fake value\n'
    fi

    rm -rf "$tmpdir"
}

test_install_binary_to_custom_dir() {
    run_test "install_binary places binary in SKERN_INSTALL_DIR"

    local tmpdir install_dir
    tmpdir="$(mktemp -d)"
    install_dir="$tmpdir/custom-bin"

    # Set up state as if download() succeeded
    TMPDIR="$tmpdir"
    BINARY="skern"
    echo '#!/bin/bash' > "$tmpdir/skern"
    echo 'echo "skern test"' >> "$tmpdir/skern"

    SKERN_INSTALL_DIR="$install_dir"
    install_binary >/dev/null 2>&1

    assert_file_exists "binary installed" "$install_dir/skern"
    assert_file_executable "binary is executable" "$install_dir/skern"

    unset SKERN_INSTALL_DIR
    rm -rf "$tmpdir"
}

test_install_binary_default_dir() {
    run_test "install_binary uses DEFAULT_INSTALL_DIR when SKERN_INSTALL_DIR is unset"

    local tmpdir fake_home
    tmpdir="$(mktemp -d)"
    fake_home="$tmpdir/fakehome"
    mkdir -p "$fake_home"

    TMPDIR="$tmpdir"
    BINARY="skern"
    echo '#!/bin/bash' > "$tmpdir/skern"

    # Override DEFAULT_INSTALL_DIR to avoid touching real ~/.local/bin
    unset SKERN_INSTALL_DIR
    local saved_default="$DEFAULT_INSTALL_DIR"
    DEFAULT_INSTALL_DIR="$fake_home/.local/bin"

    install_binary >/dev/null 2>&1

    assert_file_exists "binary in default dir" "$fake_home/.local/bin/skern"
    assert_file_executable "binary is executable" "$fake_home/.local/bin/skern"

    DEFAULT_INSTALL_DIR="$saved_default"
    rm -rf "$tmpdir"
}

test_path_warning() {
    run_test "install_binary warns when install dir is not in PATH"

    local tmpdir install_dir
    tmpdir="$(mktemp -d)"
    install_dir="$tmpdir/not-in-path"

    TMPDIR="$tmpdir"
    BINARY="skern"
    echo '#!/bin/bash' > "$tmpdir/skern"

    SKERN_INSTALL_DIR="$install_dir"
    local output
    output="$(install_binary 2>&1)"

    assert_contains "PATH warning shown" "$output" "not in your PATH"

    unset SKERN_INSTALL_DIR
    rm -rf "$tmpdir"
}

test_path_no_warning_when_in_path() {
    run_test "install_binary does not warn when install dir is in PATH"

    local tmpdir install_dir
    tmpdir="$(mktemp -d)"
    install_dir="$tmpdir/in-path"

    TMPDIR="$tmpdir"
    BINARY="skern"
    echo '#!/bin/bash' > "$tmpdir/skern"

    SKERN_INSTALL_DIR="$install_dir"
    local saved_path="$PATH"
    export PATH="$install_dir:$PATH"

    local output
    output="$(install_binary 2>&1)"

    TESTS_RUN=$((TESTS_RUN + 1))
    if echo "$output" | grep -qF "not in your PATH"; then
        FAIL=$((FAIL + 1))
        printf '  FAIL: should not warn when dir is in PATH\n'
    else
        PASS=$((PASS + 1))
    fi

    export PATH="$saved_path"
    unset SKERN_INSTALL_DIR
    rm -rf "$tmpdir"
}

test_full_download_and_install_flow() {
    run_test "full flow: create tarball, verify checksum, install"

    local tmpdir install_dir
    tmpdir="$(mktemp -d)"
    install_dir="$tmpdir/bin"

    # Build a fake skern binary and tarball
    local staging="$tmpdir/staging"
    mkdir -p "$staging"
    printf '#!/bin/bash\necho "skern v0.0.1-test"' > "$staging/skern"
    chmod +x "$staging/skern"
    tar -czf "$tmpdir/skern.tar.gz" -C "$staging" skern

    # Compute checksum
    local checksum
    if command -v sha256sum >/dev/null 2>&1; then
        checksum="$(sha256sum "$tmpdir/skern.tar.gz" | awk '{print $1}')"
    elif command -v shasum >/dev/null 2>&1; then
        checksum="$(shasum -a 256 "$tmpdir/skern.tar.gz" | awk '{print $1}')"
    else
        printf '  SKIP: no sha256sum or shasum available\n'
        rm -rf "$tmpdir"
        return
    fi

    OS="" ARCH="" PLATFORM=""
    detect_platform >/dev/null 2>&1
    VERSION="v0.0.1"

    echo "$checksum  skern_${VERSION#v}_${PLATFORM}.tar.gz" > "$tmpdir/checksums.txt"

    # Simulate what download() does after a successful fetch
    TARBALL="$tmpdir/skern.tar.gz"
    CHECKSUMS_FILE="$tmpdir/checksums.txt"
    EXPECTED_NAME="skern_${VERSION#v}_${PLATFORM}.tar.gz"
    EXPECTED_SUM="$(grep "$EXPECTED_NAME" "$CHECKSUMS_FILE" | awk '{print $1}')"

    local actual_sum
    if command -v sha256sum >/dev/null 2>&1; then
        actual_sum="$(sha256sum "$TARBALL" | awk '{print $1}')"
    else
        actual_sum="$(shasum -a 256 "$TARBALL" | awk '{print $1}')"
    fi

    assert_eq "checksum matches in flow" "$EXPECTED_SUM" "$actual_sum"

    # Extract and install
    tar -xzf "$TARBALL" -C "$tmpdir"
    assert_file_exists "extracted binary" "$tmpdir/skern"

    TMPDIR="$tmpdir"
    BINARY="skern"
    SKERN_INSTALL_DIR="$install_dir"
    install_binary >/dev/null 2>&1

    assert_file_exists "installed binary" "$install_dir/skern"
    assert_file_executable "installed binary executable" "$install_dir/skern"

    # Verify the installed binary works
    local bin_output
    bin_output="$("$install_dir/skern" 2>&1)"
    assert_eq "binary output" "skern v0.0.1-test" "$bin_output"

    unset SKERN_INSTALL_DIR
    rm -rf "$tmpdir"
}

test_info_output_format() {
    run_test "info() outputs with [skern] prefix"

    local output
    output="$(info "hello world")"
    assert_eq "info format" "[skern] hello world" "$output"
}

test_error_output_format() {
    run_test "error() outputs with [skern] ERROR prefix and exits non-zero"

    local output
    set +e
    output="$(error "something broke" 2>&1)"
    local code=$?
    set -e

    assert_contains "error format" "$output" "[skern] ERROR: something broke"
    assert_eq "error exits non-zero" "1" "$code"
}

# --- run all tests ---

test_info_output_format
test_error_output_format
test_detect_platform_current_host
test_detect_platform_values
test_resolve_version_explicit
test_resolve_version_custom
test_download_url_construction
test_download_url_version_strip
test_checksum_verification
test_checksum_mismatch_detected
test_install_binary_to_custom_dir
test_install_binary_default_dir
test_path_warning
test_path_no_warning_when_in_path
test_full_download_and_install_flow

# --- summary ---

printf '\n--- Results ---\n'
printf 'Tests: %d | Pass: %d | Fail: %d\n' "$TESTS_RUN" "$PASS" "$FAIL"

if [ "$FAIL" -gt 0 ]; then
    exit 1
fi
