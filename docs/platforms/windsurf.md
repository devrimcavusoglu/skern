# Windsurf

[Windsurf](https://windsurf.com/) is Codeium's agentic IDE. Skern installs skills into its skills directory.

## Skill Paths

| Scope | Path |
|-------|------|
| User-level | `~/.codeium/windsurf/skills/<name>/SKILL.md` |
| Project-level | `.windsurf/skills/<name>/SKILL.md` |

Windsurf does not share the `.agents/skills/` project directory with other adapters — its project skills live at `.windsurf/skills/`.

## Install a Skill

```sh
skern skill install code-review --platform windsurf
skern skill install code-review --platform windsurf --scope project
```

## Uninstall a Skill

```sh
skern skill uninstall code-review --platform windsurf
```

## Detection

Skern detects Windsurf by the presence of `~/.codeium/windsurf/` or `~/.windsurf/`.
