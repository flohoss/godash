# Agent Guidance

> **Purpose.** This file is the primary onboarding and guardrail document for any LLM
> (Claude, GPT, Gemini, Cursor, Copilot, etc.) that reads, writes, or reviews code in
> this repository. Read it before making changes. It is intentionally rule-oriented and
> self-contained.

## Code style

- **No comments.** If something needs clarification, use a function or service name that explains it.
- **No code markers** like `// ... existing code ...` in edits.
- Go imports: stdlib, then external, then internal (`github.com/flohoss/chat/...`), each block alphabetical.

## Git

Split commit message to a meaningful scope!

**Commit message format**

- Prefix with exactly one of:
  - `[fix]` — fixes a bug
  - `[feature]` — adds new functionality
  - `[improve]` — improves existing functionality
  - `[meta]` — changes outside the code base (e.g. deployment setup)
  - `[docs]` — documentation (README, these docs, etc.)
  - `[refactor]` — formatting, renaming, structural-only changes
- Capitalize the first letter after the prefix.
- **Title only — no body.**

## Verification after changes

After any code change, **always run these before committing** — do not skip even for small edits:

- **Backend:** `docker compose run --rm go fmt ./...`

Only commit if all commands pass.
