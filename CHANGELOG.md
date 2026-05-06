# Changelog

All notable changes to skern are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed

- **`skill create --from-template <path>` now requires a skill directory.**
  A skill is a folder, so `--from-template` accepts only a directory containing
  a `SKILL.md`. Passing a bare file (a `SKILL.md` or a body-only markdown file)
  is rejected with a clear error that points at the parent directory. The
  template's frontmatter (`description`, `tags`, `metadata.author`,
  `metadata.version`) is preserved on the new skill (#82) and every sibling
  file or subdirectory (e.g., `references/`, `templates/`, `VENDORED.md`) is
  copied alongside the new `SKILL.md` (#83). The CLI `<name>` argument always
  wins over the template's `name`; other flags override template values when
  explicitly set, otherwise template values are preserved. **Breaking:** the
  previous behavior of treating a non-frontmatter markdown file as a raw body
  is removed — wrap such bodies in a directory with a `SKILL.md` instead.

## [v0.2.1] — 2026-05-03

Cross-platform install. No code changes; release pipeline + install UX only.

### Added

- **Windows binaries.** Releases now publish `skern_<version>_windows_amd64.zip`
  and `skern_<version>_windows_arm64.zip` alongside the existing macOS and Linux
  tarballs. Triggered by adding `windows` to the goreleaser build matrix and a
  `format_overrides` rule that ships Windows as zip per convention.
- **`scripts/install.ps1`** — PowerShell installer mirroring `scripts/install.sh`.
  Detects amd64/arm64, downloads the matching zip, verifies SHA-256 against the
  release's `checksums.txt`, extracts to `%LOCALAPPDATA%\skern\bin`, and warns
  if the install dir is not on `PATH`. Honors `SKERN_INSTALL_DIR` and
  `SKERN_VERSION` environment variables, same as the Unix script.
- **`INSTALL.md`** at repo root — single canonical install guide with one
  command per OS (macOS / Linux / Windows), plus version-pinning, manual
  install, source build, and uninstall sections. Designed as a clean structured
  doc that an LLM agent can follow end-to-end.

### Changed

- README install section now shows the three OS one-liners side-by-side and
  links to `INSTALL.md` for full coverage.

[v0.2.1]: https://github.com/devrimcavusoglu/skern/compare/v0.2.0...v0.2.1

## [v0.2.0] — 2026-05-03

Dynamic skill loading release. **Contains breaking changes.**

### Breaking changes

- **JSON shape of `skill install` and `skill uninstall` results changed.** The
  response now uses a top-level `platform` and `capacity` block alongside a
  `skills[]` array, replacing the previous `platforms[]` array. Consumers that
  parse install/uninstall JSON must migrate. ([#52], [#76])
- **`--platform all` is no longer accepted.** Each `skill install` /
  `skill uninstall` invocation targets exactly one platform. To deploy a skill
  across multiple platforms, loop the call per platform. ([#76])

### Added

- Batch install/uninstall: pass multiple skill names per invocation. ([#76])
- Capacity reporting on every install/uninstall response — `count`, `threshold`,
  `headroom`, and `over-budget` flag — so agents can react to capacity pressure
  without an extra query. ([#76])
- `--enforce-budget` opt-in flag on `skill install` to refuse installs that
  would exceed configured per-platform capacity thresholds. ([#76])
- `--with-platforms` flag on `skern skill list` to surface per-platform
  installation state inline. ([#76])
- Skill import from URL or git repository. ([#73])
- Skill creation guidelines documented in `docs/writing-skills.md`, adapted
  from superpowers/writing-skills. ([#74])

### Changed

- `skern skill --help` now groups subcommands into **Registry commands** (where
  the skill lives in skern itself) and **Platform commands** (where the skill is
  deployed to a specific platform), making the registry → platform direction
  explicit. Per-command help strings reinforce the model: `create`/`remove` say
  "skern's registry," `install` says "registered skills onto a platform," and
  `uninstall` notes the registry is untouched.
- The "skill not found" error from `skill install` now reads "not registered in
  skern" and points users at `skill create` / `skill import` instead of just
  `skill list`.

### Fixed

- Version info falls back to `runtime/debug.ReadBuildInfo` when ldflags aren't
  injected (e.g. `go install`-built binaries). ([#72])

[v0.2.0]: https://github.com/devrimcavusoglu/skern/compare/v0.1.1...v0.2.0
[#52]: https://github.com/devrimcavusoglu/skern/issues/52
[#72]: https://github.com/devrimcavusoglu/skern/pull/72
[#73]: https://github.com/devrimcavusoglu/skern/pull/73
[#74]: https://github.com/devrimcavusoglu/skern/pull/74
[#76]: https://github.com/devrimcavusoglu/skern/pull/76

## [v0.1.1] — 2026-03-19

### Added

- `skern skill diff` command for comparing skill manifests. ([#66])
- Skill versioning and version management — semver in frontmatter, upgrade
  detection. ([#67])

### Fixed

- False-positive warnings from parsing inline code spans and URLs as file paths
  during validation. ([#69])

[v0.1.1]: https://github.com/devrimcavusoglu/skern/compare/v0.1.0...v0.1.1
[#66]: https://github.com/devrimcavusoglu/skern/pull/66
[#67]: https://github.com/devrimcavusoglu/skern/pull/67
[#69]: https://github.com/devrimcavusoglu/skern/pull/69

## [v0.1.0] — 2026-03-07

First milestone release covering M0–M5.

### Added

- `skern skill edit` / `skill update` command. ([#64])
- Skill tags/categories in SKILL.md frontmatter. ([#63])
- Stylistic lint checks in `skern skill validate`. ([#62])
- `--force` flag on `skill install` for overwrite. ([#60])
- Parse errors surfaced in `Registry.List()`. ([#61])
- Strengthened semver validation. ([#56])
- Documentation site (VitePress) at skern.dev.

### Changed

- Replaced package-level mutable state with `CommandContext`. ([#59])
- Unified overlap and discovery scoring systems. ([#58])
- Split `output` package into multiple files. ([#57])
- Slimmer README; detailed content moved to the docs site.

### Fixed

- `fsync` before close in `copyFile`. ([#55])
- Deduplicated keywords in `extractKeywords()`. ([#54])
- Install script execution when piped via `curl | bash`.

[v0.1.0]: https://github.com/devrimcavusoglu/skern/compare/v0.0.1...v0.1.0
[#54]: https://github.com/devrimcavusoglu/skern/pull/54
[#55]: https://github.com/devrimcavusoglu/skern/pull/55
[#56]: https://github.com/devrimcavusoglu/skern/pull/56
[#57]: https://github.com/devrimcavusoglu/skern/pull/57
[#58]: https://github.com/devrimcavusoglu/skern/pull/58
[#59]: https://github.com/devrimcavusoglu/skern/pull/59
[#60]: https://github.com/devrimcavusoglu/skern/pull/60
[#61]: https://github.com/devrimcavusoglu/skern/pull/61
[#62]: https://github.com/devrimcavusoglu/skern/pull/62
[#63]: https://github.com/devrimcavusoglu/skern/pull/63
[#64]: https://github.com/devrimcavusoglu/skern/pull/64

## [v0.0.1] — 2026-03-02

Initial public release.

[v0.0.1]: https://github.com/devrimcavusoglu/skern/releases/tag/v0.0.1
