# Changelog

All notable changes to skern are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- **`--category` filter on `skern skill list`.** Filters by namespaced tag
  (`--category lang:python`), repeatable, comma-lists values, OR within a
  category and AND across categories. `--include-untagged` treats a skill with
  no tag in a requested category as matching it. Tag validation now enforces a
  single canonical form: lowercase alphanumeric segments joined by hyphens,
  with at most one colon separating category from value. Tag *filters* remain
  case-insensitive, so legacy hand-edited uppercase tags still match.
  ([#96], [#98])

### Changed

- **Docs and agent guide restructured.** `AGENTS.md` now carries conventions,
  architecture, and design decisions only; contributor setup (build/test
  matrix, directory tree, dependencies) lives in
  `docs/contributing/development.md`, with each cross-linking the other
  instead of restating it. Closed milestone history (M0–M7) was dropped in
  favor of this changelog. Design decision 3 was amended: adapters remain
  declarative, but pure-copy is no longer an invariant — a `Spec` may carry
  post-install behavior where a platform's loader requires it.
- **`skern skill --help` lists every supported platform** instead of a
  hardcoded first three, derived from the spec table.

### Fixed

- **Unrecognized frontmatter keys are no longer dropped.** Every `SKILL.md`
  rewrite — `skill create --from-template`, `skill import`, `skill edit`,
  `skill version` — used to re-serialize only the keys skern models, silently
  discarding everything else (`compatibility`, `handoffs`, `metadata.phases`,
  a `metadata.tags` list, …). Unmodeled keys — top-level, `metadata.*`, and
  nested inside `metadata.author` / `modified-by` entries — now round-trip
  (values preserved as parsed; modeled keys written first, extras sorted
  after; YAML 1.1-only scalars such as bare dates and `yes`/`no` are
  canonicalized). `skill diff` reports differing pass-through keys under
  their own names. ([#100])
- **Release workflow is idempotent on duplicate tag-push deliveries.** A
  redelivered tag push no longer fails the run or produces a partial release.
  ([#95])

### Internal

- GitHub Actions modernized: Node 24 action majors, floating tool versions
  pinned. ([#90], [#97])
- Corrected stale claims in the agent guide: `santhosh-tekuri/jsonschema` was
  listed as a planned dependency but was never adopted (validation is
  bespoke); the per-platform project capacity path was documented as
  `<platform>/.skern/skills/` rather than the platform's own project skills
  directory; the milestone-numbered branch convention was replaced with the
  `<type>/<slug>` scheme actually in use.

[#90]: https://github.com/devrimcavusoglu/skern/issues/90
[#95]: https://github.com/devrimcavusoglu/skern/pull/95
[#96]: https://github.com/devrimcavusoglu/skern/issues/96
[#97]: https://github.com/devrimcavusoglu/skern/pull/97
[#98]: https://github.com/devrimcavusoglu/skern/pull/98
[#100]: https://github.com/devrimcavusoglu/skern/issues/100

## [v0.3.1] — 2026-05-13

Patch release: widen skill-name validation to accept dot-namespaced names.
Strict superset of the v0.3.0 regex — every previously valid name remains
valid, no migration required.

### Added

- **Dot-namespaced skill names are now valid.** `ValidateName` accepts dots as
  a segment separator in addition to hyphens, so names like `myorg.bootstrap`
  and `codebase-intelligence.scan` pass validation. The hyphen-only form
  (`code-review`) remains valid; this is a strict superset of the previous
  regex, no migration required. Enables installers to preserve their
  namespace prefix on platforms that derive slash commands from directory
  names (e.g. GitHub Copilot maps `.copilot/skills/myorg.bootstrap/` to
  `/myorg.bootstrap`). ([#92], [#93])

[#92]: https://github.com/devrimcavusoglu/skern/issues/92
[#93]: https://github.com/devrimcavusoglu/skern/pull/93

## [v0.3.0] — 2026-05-07

Platform registry rewrite, five new adapters, agent-instruction snippet
writer, and a focused breaking change to `--from-template`.

### Breaking changes

- **`skill create --from-template <path>` now requires a skill directory.**
  A skill is a folder, so `--from-template` accepts only a directory containing
  a `SKILL.md`. Passing a bare file (a `SKILL.md` or a body-only markdown file)
  is rejected with a clear error pointing at the parent directory. The
  previous behavior of treating a non-frontmatter markdown file as a raw body
  is removed — wrap such bodies in a directory with a `SKILL.md` instead. ([#84])

### Added

- **`skern init` can now write a skern usage snippet into agent instruction
  files** (`AGENTS.md`, `CLAUDE.md`, `.claude/CLAUDE.md`). Off by default;
  opt in with `--instructions` (or accept the interactive prompt on a TTY).
  The snippet is wrapped in `<!-- skern:instructions:start -->` /
  `<!-- skern:instructions:end -->` markers so re-running updates the block
  in place. Three additional flags shape the output: `--tool-forming-loop`
  appends a search-before-create workflow section (off by default),
  `--target <path>` overrides auto-discovery for explicit files, and
  `--print-instructions` emits the snippet to stdout without writing files.
  The `InitResult` JSON envelope grows an `instructions` field reporting
  what was written. ([#87])
- **Five new platform adapters: `cursor`, `gemini-cli`, `github-copilot`,
  `windsurf`, `continue`.** All five accept the same `--platform` flag, route
  installs to the platform's expected skill directory, and participate in
  `skern platform list`/`status` matrices. Path conventions follow
  [vercel-labs/skills](https://github.com/vercel-labs/skills#supported-agents).
  ([#81])
- **Declarative platform registry.** Adapters are now defined as one row in
  `internal/platform/spec.go` — a `Spec` carrying name, user dir, project dir,
  and home-relative detection paths. A single generic `Adapter` struct
  implements the `Platform` interface from any spec row, replacing the
  per-platform Go files (`claude.go`, `codex.go`, `opencode.go`). Adding a
  platform is a one-line PR. ([#81])
- **Windows added to the CI matrix** — every PR now runs the test suite on
  Windows in addition to macOS and Linux. ([#86])
- **`--from-template` preserves template frontmatter and copies sibling
  assets.** The template's `description`, `tags`, `metadata.author`, and
  `metadata.version` are preserved on the new skill (#82) and every sibling
  file or subdirectory (`references/`, `templates/`, `VENDORED.md`, …) is
  copied alongside the new `SKILL.md` (#83). The CLI `<name>` argument always
  wins over the template's name; other flags override template values when
  explicitly set, otherwise template values are preserved.

### Changed

- **Platform detection is per-platform, not per-directory.** Several adapters
  share `.agents/skills/` as their project dir; detection now keys on each
  adapter's distinct user-level config dir (`~/.cursor`, `~/.gemini`,
  `~/.copilot`, `~/.codex`, etc.) so `platform list` doesn't false-positive
  for platforms whose CLI isn't installed.
- **CLI flag help and error messages enumerate the registered platforms
  dynamically** — adding a platform updates `--platform` help and the
  "unknown platform" error text without touching the CLI.
- **Documentation site comprehensively refreshed** — eight platform pages,
  reference covers `skill import` / `skill version` / `skill diff`, validation
  page split into errors/warnings/hints, installation page covers all three
  OSes, agent-setup leads with `init --instructions`.

[v0.3.0]: https://github.com/devrimcavusoglu/skern/compare/v0.2.1...v0.3.0
[#81]: https://github.com/devrimcavusoglu/skern/pull/81
[#84]: https://github.com/devrimcavusoglu/skern/pull/84
[#86]: https://github.com/devrimcavusoglu/skern/pull/86
[#87]: https://github.com/devrimcavusoglu/skern/pull/87

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
