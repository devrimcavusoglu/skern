# Installation

Skern ships pre-built binaries for macOS, Linux, and Windows. Pick the section that matches your operating system.

The full canonical install guide lives at [INSTALL.md](https://github.com/devrimcavusoglu/skern/blob/main/INSTALL.md) in the repo — including custom install paths, manual install, and uninstall instructions. This page covers the most common paths.

## macOS / Linux

```sh
curl -fsSL https://raw.githubusercontent.com/devrimcavusoglu/skern/main/scripts/install.sh | bash
```

The installer detects amd64/arm64, downloads the matching tarball, verifies its SHA-256 against the release's `checksums.txt`, and installs to `~/.local/bin/skern` by default.

## Windows (PowerShell)

```powershell
irm https://raw.githubusercontent.com/devrimcavusoglu/skern/main/scripts/install.ps1 | iex
```

Installs to `%LOCALAPPDATA%\skern\bin\skern.exe`. If your PowerShell execution policy blocks the script:

```powershell
powershell -ExecutionPolicy Bypass -Command "irm https://raw.githubusercontent.com/devrimcavusoglu/skern/main/scripts/install.ps1 | iex"
```

## Pin a Version

Set `SKERN_VERSION` before running the installer:

```sh
SKERN_VERSION=v0.2.1 curl -fsSL https://raw.githubusercontent.com/devrimcavusoglu/skern/main/scripts/install.sh | bash
```

```powershell
$env:SKERN_VERSION = 'v0.2.1'; irm https://raw.githubusercontent.com/devrimcavusoglu/skern/main/scripts/install.ps1 | iex
```

## Custom Install Directory

Both installers honor `SKERN_INSTALL_DIR`:

```sh
SKERN_INSTALL_DIR=/usr/local/bin curl -fsSL https://raw.githubusercontent.com/devrimcavusoglu/skern/main/scripts/install.sh | bash
```

## From Source (Go 1.25+)

```sh
go install github.com/devrimcavusoglu/skern/cmd/skern@latest
```

When built this way, version info falls back to `runtime/debug.ReadBuildInfo` since ldflags aren't injected.

## Build Locally

```sh
git clone https://github.com/devrimcavusoglu/skern.git
cd skern
make build
```

The binary is placed at the repository root as `skern`. Move it onto your `PATH` to use it globally.

## Manual Install

Download the archive for your platform from [the releases page](https://github.com/devrimcavusoglu/skern/releases/latest) and extract the binary. Archive naming:

- `skern_<version>_darwin_<arch>.tar.gz`
- `skern_<version>_linux_<arch>.tar.gz`
- `skern_<version>_windows_<arch>.zip`

`<arch>` is `amd64` or `arm64`. Each release also publishes `checksums.txt` with SHA-256 sums.

## Verify Installation

```sh
skern --version
```

If the command is not found, the install directory is not yet on your `PATH`. The shell installer prints PATH instructions for your shell at the end — follow those, then open a new shell.

## Uninstall

Delete the binary at the install location. The installer does not modify any other files.
