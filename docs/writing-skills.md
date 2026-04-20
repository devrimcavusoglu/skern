# Writing Effective Skills

This guide covers how to write high-quality, discoverable, well-structured skills for skern. These guidelines are adapted from the [superpowers writing-skills](https://github.com/obra/superpowers/tree/main/skills/writing-skills) project and are built into skern's scaffolding and validation.

## Skill Types

| Type | Description | Example Name |
|------|-------------|--------------|
| **Technique** | Concrete method with actionable steps | `condition-based-waiting` |
| **Pattern** | Mental model for problem-solving | `flatten-with-flags` |
| **Reference** | API docs, syntax guides, lookup tables | `office-docs` |

## When to Create a Skill

Create a skill when:

- The technique was not intuitively obvious
- You would reference it across multiple projects
- The pattern applies broadly beyond one codebase
- Others would benefit from it

Do **not** create a skill for:

- One-time solutions or narratives about specific sessions
- Patterns already well-documented in official tooling

## Description: Claude Search Optimization (CSO)

The `description` field in your SKILL.md frontmatter is the most critical field for agent discoverability. Agents match skills by description, so how you write it directly affects whether your skill gets loaded.

**Start with "Use when..."** and describe the triggering conditions:

```yaml
# Good
description: |
  Use when tests hang or flake due to timing-dependent assertions.
  Symptoms: intermittent CI failures, setTimeout in test code.

# Bad — summarizes the workflow (agents may shortcut)
description: |
  A skill that teaches agents how to write better async tests
  using condition-based waiting patterns.
```

**Why this matters:** When a description summarizes the skill's workflow, an agent may follow the description as a shortcut instead of reading the full SKILL.md body.

**Keyword tips:**

- Include error messages agents might encounter
- Add symptom words: "flaky", "hanging", "race condition", "timeout"
- Use tool names and common synonyms

## Naming Conventions

- Use `kebab-case` with lowercase letters, numbers, and hyphens
- Prefer verb-first active voice: `creating-skills` not `skill-creation`
- Be specific: `condition-based-waiting` not `async-test-helpers`

Names must match `^[a-z0-9]+(-[a-z0-9]+)*$` and be 1-64 characters.

## Recommended Body Structure

When you run `skern skill create`, the scaffold includes these sections by default:

### Overview

1-2 sentences describing the core principle or technique. Keep it concise.

### When to Use

List triggering conditions as bullet points:

- Symptoms that signal this skill is needed
- Specific use cases
- When **not** to use this skill (counter-examples help agents avoid false matches)

### Core Pattern

The main technique or pattern. For techniques, use a before/after comparison:

```markdown
## Core Pattern

**Before (anti-pattern):**
`setTimeout(check, 1000)` — arbitrary delay, still flaky

**After (correct):**
`waitFor(() => expect(value).toBe(true))` — condition-based, deterministic
```

### Quick Reference

A scannable table or bullet list for fast lookup. Agents should be able to use this section as a cheat sheet without reading the full skill.

### Common Mistakes

Frequent errors and their fixes. Helps agents self-correct.

## Token Efficiency

Skills are loaded into agent context windows, so brevity matters:

| Skill Category | Target Word Count |
|----------------|-------------------|
| Getting-started workflows | < 150 words |
| Frequently-loaded skills | < 200 words |
| Other skills | < 500 words |

**Tips:**

- Use cross-references instead of repeating shared concepts
- Compress examples to the minimum that demonstrates the point
- Move heavy reference material (100+ lines) to supporting files

## Code Examples

- **One excellent example beats many mediocre ones**
- Use the most relevant language for the skill's domain
- Keep examples copy-pasteable and runnable

## File Organization

| Layout | When to Use |
|--------|------------|
| Single `SKILL.md` | All content fits in one file |
| `SKILL.md` + supporting files | Heavy reference (100+ lines) or reusable tool code |

Supporting files live in the same directory as `SKILL.md`. Reference them in the body.

## Anti-Patterns to Avoid

- **Narrative storytelling** — "In session 2025-10-03, I encountered..." Skills are reusable references, not journals.
- **Multi-language examples** — Pick one language. Multiple languages dilute quality and waste tokens.
- **Generic labels** — Avoid `helper1`, `step3`, `utils`. Use descriptive names.
- **Workflow summaries in description** — The description field is for triggering conditions only.

## Validation Hints

Skern's validator provides hints when your skill deviates from these guidelines:

- **Description missing trigger prefix** — Hints when the description doesn't start with "Use when", "Use for", "Use to", etc.
- **Missing recommended sections** — Hints when the body is missing "When to Use", "Core Pattern", "Quick Reference", or "Common Mistakes" sections.
- **Body too short** — Hints when the body has fewer than 20 words.
- **Description too vague** — Hints when the description has fewer than 3 words.

These are non-blocking hints (not errors), designed to guide you toward better skills.

## Next Steps

- [Quick Start](/guide/quick-start) — create and install your first skill
- [Skill Format](/concepts/skill-format) — full SKILL.md specification
- [Validation Reference](/reference/validation) — all validation rules
