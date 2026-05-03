# Installing skern

Pick the section that matches your operating system. Each section contains exactly one install command. After running it, verify with `skern --version`.

## macOS

```sh
curl -fsSL https://raw.githubusercontent.com/devrimcavusoglu/skern/main/scripts/install.sh | bash
```

## Linux

```sh
curl -fsSL https://raw.githubusercontent.com/devrimcavusoglu/skern/main/scripts/install.sh | bash
```

## Windows

```powershell
irm https://raw.githubusercontent.com/devrimcavusoglu/skern/main/scripts/install.ps1 | iex
```

If your PowerShell execution policy blocks the script:

```powershell
powershell -ExecutionPolicy Bypass -Command "irm https://raw.githubusercontent.com/devrimcavusoglu/skern/main/scripts/install.ps1 | iex"
```

## Verify

```sh
skern --version
```

If the command is not found, the install directory is not yet on your `PATH`. The installer prints PATH instructions for your shell at the end of its output — follow those, then open a new shell.

## Default install locations

| OS              | Path                            |
|-----------------|---------------------------------|
| macOS / Linux   | `~/.local/bin/skern`            |
| Windows         | `%LOCALAPPDATA%\skern\bin\skern.exe` |

Override with the `SKERN_INSTALL_DIR` environment variable.

## Pin a specific version

```sh
SKERN_VERSION=v0.2.1 curl -fsSL https://raw.githubusercontent.com/devrimcavusoglu/skern/main/scripts/install.sh | bash
```

```powershell
$env:SKERN_VERSION = 'v0.2.1'; irm https://raw.githubusercontent.com/devrimcavusoglu/skern/main/scripts/install.ps1 | iex
```

## Build from source

Requires Go 1.25+:

```sh
go install github.com/devrimcavusoglu/skern/cmd/skern@latest
```

## Manual install

Download the archive for your platform from the [releases page](https://github.com/devrimcavusoglu/skern/releases/latest), extract it, and place the `skern` (or `skern.exe`) binary on your `PATH`. Archive naming:

- `skern_<version>_darwin_<arch>.tar.gz`
- `skern_<version>_linux_<arch>.tar.gz`
- `skern_<version>_windows_<arch>.zip`

`<arch>` is `amd64` or `arm64`. Each release also publishes `checksums.txt` with SHA-256 sums.

## Uninstall

Delete the binary at the install location shown above. The installer does not modify any other files.
