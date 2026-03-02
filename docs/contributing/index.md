# Contributing

Contributions to skern are welcome. This guide covers the basics of getting started.

## Getting the Source

```sh
git clone https://github.com/devrimcavusoglu/skern.git
cd skern
```

## Requirements

- Go 1.23+
- `golangci-lint` (for linting)
- `make`

## Workflow

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests and linting
5. Submit a pull request

## Quick Links

- [Development](/contributing/development) — build, test, and lint commands
- [Manual Testing](/contributing/manual-testing) — agent test harness for pre-release testing

## Issue Tracking

The project uses `br` (beads-rust) for issue tracking. Reference issues in commits as `br#<id>`.

```sh
br list               # Open issues
br create "Title"     # New issue
br close <id>         # Close an issue
```

## License

By contributing, you agree that your contributions will be licensed under the Apache 2.0 License.
