# Scenario 9: Skill Creation from Templates

## Assets provided

- `templates/code-review/` — Skill directory template for a code review skill
- `templates/test-helper/` — Skill directory template for a test writing skill

Each template directory contains a `SKILL.md` (with frontmatter) and may contain
companion files alongside it.

## Prompt to give the agent

> Create two skills using the template directories under templates/:
> 1. A skill called "code-review" using templates/code-review as the source
> 2. A skill called "test-helper" using templates/test-helper as the source
> Then validate both skills.

## What to observe

1. Does the agent discover and use the `--from-template` flag?
2. Does it pass the skill *directory* (not the SKILL.md file) to `--from-template`?
3. Are the new skills created with the template's frontmatter and body preserved?
4. Are companion files (if any) present alongside the new SKILL.md?
5. Do both skills pass validation?
6. Does the agent use `skern skill show` to verify the body content was applied?
