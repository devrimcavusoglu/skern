# Expected Behavior — Scenario 11

## Pass criteria

- [ ] Agent reads Go source files and identifies formatting/doc-comment inconsistencies
- [ ] Agent discovers scribe by reading AGENTS.md
- [ ] Agent **autonomously decides** to create a skill (not told to)
- [ ] Agent searches or recommends before creating (no duplicates)
- [ ] Agent creates at least one skill with valid name, description, and meaningful body
- [ ] Agent validates the created skill(s)
- [ ] Skill body contains actionable instructions (not just a title)

## Bonus (nice-to-have)

- [ ] Agent installs the skill to claude-code platform
- [ ] Agent creates multiple complementary skills (e.g., formatting + doc comments)
- [ ] Agent also fixes the existing Go files (formatting, doc comments)
- [ ] Agent commits changes or suggests doing so

## Key signal

The agent was never told to "create a skill." The prompt asks about code consistency "going forward" for "any agent working in this repo." A passing agent must bridge the gap from project need to skill creation on its own.

## Verification commands

```sh
# Skills were created:
scribe skill list --json

# Skills are valid:
scribe skill validate <name> --json

# Skill body is non-trivial (more than just frontmatter):
cat .scribe/skills/<name>/SKILL.md
```
