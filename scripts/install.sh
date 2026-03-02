#!/bin/bash
# Skern installer — OS-agnostic install script
# Usage: curl -fsSL https://raw.githubusercontent.com/devrimcavusoglu/skern/main/scripts/install.sh | bash
#
# Environment variables:
#   SKERN_INSTALL_DIR  — override install directory (default: ~/.local/bin)
#   SKERN_VERSION      — install a specific version (default: latest)

set -e

REPO="devrimcavusoglu/skern"
BINARY="skern"
DEFAULT_INSTALL_DIR="$HOME/.local/bin"

# --- helpers ---

info() {
    printf '[skern] %s\n' "$1"
}

error() {
    printf '[skern] ERROR: %s\n' "$1" >&2
    exit 1
}

# --- platform detection ---

detect_platform() {
    OS="$(uname -s)"
    ARCH="$(uname -m)"

    case "$OS" in
        Linux)  OS="linux" ;;
        Darwin) OS="darwin" ;;
        *)      error "Unsupported OS: $OS. Skern supports Linux and macOS." ;;
    esac

    case "$ARCH" in
        x86_64|amd64)   ARCH="amd64" ;;
        aarch64|arm64)  ARCH="arm64" ;;
        *)              error "Unsupported architecture: $ARCH. Skern supports amd64 and arm64." ;;
    esac

    PLATFORM="${OS}_${ARCH}"
    info "Detected platform: $PLATFORM"
}

# --- version resolution ---

resolve_version() {
    if [ -n "$SKERN_VERSION" ]; then
        VERSION="$SKERN_VERSION"
        info "Using specified version: $VERSION"
        return
    fi

    info "Fetching latest release..."
    if command -v curl >/dev/null 2>&1; then
        VERSION="$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"tag_name": *"([^"]+)".*/\1/')"
    elif command -v wget >/dev/null 2>&1; then
        VERSION="$(wget -qO- "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"tag_name": *"([^"]+)".*/\1/')"
    else
        error "Neither curl nor wget found. Install one of them or use 'go install' instead."
    fi

    if [ -z "$VERSION" ]; then
        error "Could not determine latest version. Set SKERN_VERSION manually or use 'go install'."
    fi

    info "Latest version: $VERSION"
}

# --- download and verify ---

download() {
    DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${VERSION}/skern_${VERSION#v}_${PLATFORM}.tar.gz"
    CHECKSUMS_URL="https://github.com/${REPO}/releases/download/${VERSION}/checksums.txt"

    TMPDIR="$(mktemp -d)"
    trap 'rm -rf "$TMPDIR"' EXIT

    TARBALL="$TMPDIR/skern.tar.gz"
    CHECKSUMS_FILE="$TMPDIR/checksums.txt"

    info "Downloading $DOWNLOAD_URL ..."
    if command -v curl >/dev/null 2>&1; then
        curl -fsSL -o "$TARBALL" "$DOWNLOAD_URL" || return 1
        curl -fsSL -o "$CHECKSUMS_FILE" "$CHECKSUMS_URL" || CHECKSUMS_FILE=""
    elif command -v wget >/dev/null 2>&1; then
        wget -qO "$TARBALL" "$DOWNLOAD_URL" || return 1
        wget -qO "$CHECKSUMS_FILE" "$CHECKSUMS_URL" || CHECKSUMS_FILE=""
    fi

    # verify checksum
    if [ -n "$CHECKSUMS_FILE" ] && [ -f "$CHECKSUMS_FILE" ]; then
        EXPECTED_NAME="skern_${VERSION#v}_${PLATFORM}.tar.gz"
        EXPECTED_SUM="$(grep "$EXPECTED_NAME" "$CHECKSUMS_FILE" | awk '{print $1}')"

        if [ -n "$EXPECTED_SUM" ]; then
            if command -v sha256sum >/dev/null 2>&1; then
                ACTUAL_SUM="$(sha256sum "$TARBALL" | awk '{print $1}')"
            elif command -v shasum >/dev/null 2>&1; then
                ACTUAL_SUM="$(shasum -a 256 "$TARBALL" | awk '{print $1}')"
            else
                info "Warning: no sha256sum or shasum found, skipping checksum verification."
                ACTUAL_SUM=""
            fi

            if [ -n "$ACTUAL_SUM" ] && [ "$ACTUAL_SUM" != "$EXPECTED_SUM" ]; then
                error "Checksum mismatch! Expected $EXPECTED_SUM, got $ACTUAL_SUM. Aborting."
            elif [ -n "$ACTUAL_SUM" ]; then
                info "Checksum verified."
            fi
        else
            info "Warning: could not find checksum for $EXPECTED_NAME, skipping verification."
        fi
    else
        info "Warning: checksums file not available, skipping verification."
    fi

    # extract
    tar -xzf "$TARBALL" -C "$TMPDIR"

    if [ ! -f "$TMPDIR/$BINARY" ]; then
        error "Binary not found after extraction. Archive may be corrupt."
    fi

    return 0
}

# --- install ---

install_binary() {
    INSTALL_DIR="${SKERN_INSTALL_DIR:-$DEFAULT_INSTALL_DIR}"
    mkdir -p "$INSTALL_DIR"

    mv "$TMPDIR/$BINARY" "$INSTALL_DIR/$BINARY"
    chmod +x "$INSTALL_DIR/$BINARY"

    info "Installed $BINARY to $INSTALL_DIR/$BINARY"

    # check if install dir is in PATH
    case ":$PATH:" in
        *":$INSTALL_DIR:"*) ;;
        *)
            info ""
            info "WARNING: $INSTALL_DIR is not in your PATH."
            info "Add it by appending this to your shell profile (~/.bashrc, ~/.zshrc, etc.):"
            info ""
            info "  export PATH=\"$INSTALL_DIR:\$PATH\""
            info ""
            ;;
    esac
}

# --- fallback: go install ---

go_install() {
    info "Falling back to 'go install'..."
    if ! command -v go >/dev/null 2>&1; then
        error "Go is not installed. Install Go 1.23+ from https://go.dev/dl/ or download a binary from GitHub Releases."
    fi

    GO_VERSION="$(go version | grep -oE 'go[0-9]+\.[0-9]+' | sed 's/go//')"
    MAJOR="$(echo "$GO_VERSION" | cut -d. -f1)"
    MINOR="$(echo "$GO_VERSION" | cut -d. -f2)"
    if [ "$MAJOR" -lt 1 ] || { [ "$MAJOR" -eq 1 ] && [ "$MINOR" -lt 23 ]; }; then
        error "Go 1.23+ is required (found $GO_VERSION). Please upgrade Go."
    fi

    if [ -n "$SKERN_VERSION" ]; then
        go install "github.com/${REPO}/cmd/skern@${SKERN_VERSION}"
    else
        go install "github.com/${REPO}/cmd/skern@latest"
    fi

    info "Installed via 'go install'. Binary is in your GOBIN or GOPATH/bin."
}

# --- main ---

main() {
    info "Skern installer"
    info ""

    detect_platform
    resolve_version

    if download; then
        install_binary
    else
        info "Binary download failed."
        go_install
    fi

    # verify installation
    if command -v "$BINARY" >/dev/null 2>&1; then
        info ""
        info "Success! Installed $($BINARY version 2>/dev/null || echo "$BINARY")"
    else
        info ""
        info "Installation complete. You may need to restart your shell or update your PATH."
    fi
}

# Guard: only run main when executed directly or piped (not sourced for testing).
# When piped (curl | bash), BASH_SOURCE[0] is empty and $0 is "bash".
# When sourced, BASH_SOURCE[0] is set but differs from $0.
if [[ "${BASH_SOURCE[0]}" == "$0" ]] || [[ -z "${BASH_SOURCE[0]}" ]]; then
    main
fi
