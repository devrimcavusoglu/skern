# Continue

[Continue](https://continue.dev/) is an open-source AI code assistant. Skern installs skills into its skills directory.

## Skill Paths

| Scope | Path |
|-------|------|
| User-level | `~/.continue/skills/<name>/SKILL.md` |
| Project-level | `.continue/skills/<name>/SKILL.md` |

Continue uses dedicated user and project directories — it does not share the `.agents/skills/` directory.

## Install a Skill

```sh
skern skill install code-review --platform continue
skern skill install code-review --platform continue --scope project
```

## Uninstall a Skill

```sh
skern skill uninstall code-review --platform continue
```

## Detection

Skern detects Continue by the presence of `~/.continue/`.
