# Scenario 9: Skill Creation from Templates

## Assets provided

- `templates/review.md` — Code review skill body template
- `templates/test-helper.md` — Test writing skill body template

## Prompt to give the agent

> Create two skills using the template files in the templates/ directory:
> 1. A skill called "code-review" using templates/review.md as the body
> 2. A skill called "test-helper" using templates/test-helper.md as the body
> Then validate both skills.

## What to observe

1. Does the agent discover and use the `--from-template` flag?
2. Does it use the correct path to the template files?
3. Are the skills created with the template content as their body?
4. Do both skills pass validation?
5. Does the agent use `skern skill show` to verify the body content was applied?
