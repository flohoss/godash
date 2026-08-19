# Agent Guidance

Read before making changes. Rule-oriented and self-contained.

## Principles

- **CLEAN code.** Small functions, single responsibility, descriptive names, no dead code, no overengineering.
- **No comments.** Use descriptive function or service names instead.
- **No code markers** like `// ... existing code ...` in edits.
- Go imports: stdlib, then external, then internal (`github.com/flohoss/godash/...`), each block alphabetical.
- Never edit generated files (`*_templ.go`).

## Tooling — always via Docker Compose, never on the host

- **Dev server:** `docker compose up` (auto-rebuild + hot-reload via `templ generate --watch --proxy`). The server compiles `*.templ` into `*_templ.go` and **hot-reloads the browser on every file change**. Do not run `go build` or `templ generate` manually. Do not edit `*_templ.go` by hand — edit the matching `*.templ` source and let the watcher regenerate. Errors appear live in the running server/browser.
- **Format:** `docker compose run --rm go fmt ./...`

Run formatting after every code change, even small edits. Only commit if formatting passes.

## Common Commands

```sh
docker compose run --rm npm install
docker compose run --rm --entrypoint npx npm --yes npm-check-updates -u && docker compose run --rm npm install
docker compose run --rm go get -u ./...
docker compose run --rm go mod tidy
docker compose run --rm go fmt ./...
```

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
