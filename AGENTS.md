# Agent Guidance

> Primary onboarding and guardrail document for any LLM reading, writing, or reviewing code in this repository. Read before making changes.

## Code style

- No comments — use descriptive function/service names instead.
- No code markers like `// ... existing code ...` in edits.
- Go imports: stdlib, then external, then internal (`github.com/flohoss/chat/...`), each block alphabetical.
- CLEAN code: small functions, single responsibility, descriptive names, no dead code, no overengineering.

## Git

- Do not commit automatically — wait until explicitly asked.
- One commit per concern — never batch unrelated changes.
- Title only, no body. Capitalize first letter after the prefix:
  - `[fix]` bug fix
  - `[feature]` new functionality
  - `[improve]` improvement to existing functionality
  - `[refactor]` formatting, renaming, structural-only
  - `[meta]` deployment, CI
  - `[docs]` documentation

## Verification

The dev server runs `templ generate --watch --proxy` — it compiles `*.templ` into `*_templ.go`, runs `go run .` behind the proxy, and **hot-reloads the browser on every file change**. Do not run `go build` or `templ generate` manually. Do not edit `*_templ.go` by hand — edit the matching `*.templ` source and let the watcher regenerate. Errors appear live in the running server/browser.

After any code change, run formatting before committing (do not skip even for small edits):

- `docker compose run --rm go fmt ./...`

Only commit if formatting passes.
