# Loomi

Loomi is a local-first agent workspace for Chat, Work, tools, memory, and project context. It is built around a visible run timeline: model messages, tool calls, approvals, worker steps, memory snapshots, artifacts, and failures are recorded so the user can see what happened and recover cleanly.

## What It Does

- Chat with configured or local providers through the real run, event, and worker pipeline.
- Run Work-mode tasks with permissioned tools for workspace files, web search, web fetch, browser sessions, artifacts, memory, and coordination records.
- Keep tool calls observable: safe read-only actions can run automatically where allowed, while writes, commands, browser actions, and other risky operations remain approval-gated.
- Store provider configuration, tool state, selected workspace roots, memory, artifacts, and run history in the backend instead of relying on mock UI state.
- Provide a desktop-feeling interface for conversations, work runs, provider setup, web search keys, memory, tools, and runtime diagnostics.

## Current Status

Loomi now includes a Go API and worker backend, PostgreSQL-backed product data, a React/Electron desktop shell, model gateway integration, streaming run events, provider configuration, approval-gated tool execution, workspace tools, web search/fetch, browser automation foundations, artifact metadata, memory tools, and Work-mode projections.

It is still under active development. The current focus is making the real desktop experience reliable: provider setup, directory access, tool selection, readable tool summaries, run recovery, memory behavior, and simple end-to-end testing.

## Run Locally

Start the API:

```bash
DATABASE_URL="postgres://loomi:loomi@127.0.0.1:55433/loomi_m2?sslmode=disable" \
APP_ENV=local \
HTTP_ADDR=127.0.0.1:18080 \
go run ./cmd/loomi-api
```

Start the desktop shell:

```bash
VITE_LOOMI_API_BASE_URL=http://127.0.0.1:18080 bun run --cwd web desktop:dev
```

## Repository Layout

- `cmd/` - API, CLI, seed, and executable entry points.
- `internal/` - backend product data, runtime, HTTP API, worker, providers, tools, memory, and diagnostics.
- `web/` - React/Electron desktop shell and frontend runtime.
- `migrations/` - database schema migrations.
- `specs/` - Spec Kit feature specs, plans, and tasks.
- `docs-site/` - Starlight documentation for architecture, APIs, runbooks, roadmap, and devlogs.
- `docs/` - older planning notes and reference material.

## Development

```bash
go test ./...
bun test --cwd web
bun run --cwd web build
bun run --cwd docs-site build
git diff --check
```

## Safety Model

Loomi treats tools and local context as permissioned capabilities. File access is scoped to a selected workspace root, sensitive paths are denied, secrets are redacted from events and UI, and risky actions are approval-gated and visible in the run timeline.
