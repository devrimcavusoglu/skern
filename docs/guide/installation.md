# Installation

## Quick Install (Linux / macOS)

```sh
curl -fsSL https://raw.githubusercontent.com/devrimcavusoglu/skern/main/scripts/install.sh | bash
```

To install a specific version:

```sh
SKERN_VERSION=v0.0.1 curl -fsSL https://raw.githubusercontent.com/devrimcavusoglu/skern/main/scripts/install.sh | bash
```

## Go Install

Requires Go 1.25+.

```sh
go install github.com/devrimcavusoglu/skern/cmd/skern@latest
```

## Build from Source

```sh
git clone https://github.com/devrimcavusoglu/skern.git
cd skern
make build
```

The binary will be placed in the repository root as `skern`. Move it to a directory in your `PATH` to use it globally.

## Verify Installation

```sh
skern version
```
