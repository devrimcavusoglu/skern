# Contributing

Contributions to skern are welcome. This guide covers the basics of getting started.

## Getting the Source

```sh
git clone https://github.com/devrimcavusoglu/skern.git
cd skern
```

## Requirements

- Go 1.25+
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
- [PR Manual Testing](/contributing/pr-manual-testing) — manual test verification for pull requests

## Issue Tracking

The project uses GitHub Issues for tracking. Reference issues in commits as `#<number>`.

```sh
gh issue list                             # List open issues
gh issue create --title "Title" --body "" # New issue
gh issue close <number>                   # Close an issue
```

## License

By contributing, you agree that your contributions will be licensed under the Apache 2.0 License.
