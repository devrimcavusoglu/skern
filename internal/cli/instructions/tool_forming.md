### Tool-forming loop (always run first)

Before doing recurring or pattern-following work:

1. `skern skill search <topic>` — see if a relevant skill already exists.
2. Match with score ≥ 0.6: read it via `skern skill show <name>` and follow its body.
3. No match: `skern skill create <name>` — implement, and it's reusable next time.

Score ≥ 0.9 means a near-duplicate exists (creation blocked without `--force`) — read before forcing. Capacity warnings (≥20 project, ≥50 user) appear in JSON output; prefer editing existing skills once you see them.
