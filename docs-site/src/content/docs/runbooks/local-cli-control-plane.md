---
title: Local CLI Control Plane Validation
description: Focused validation for cmd/loomi and internal/cli.
---

## Focused Checks

```bash
go test ./cmd/loomi ./internal/cli
go run ./cmd/loomi help tools
go run ./cmd/loomi config show
LOOMI_CONFIG=/tmp/loomi-config.json go run ./cmd/loomi config set host http://127.0.0.1:8080
LOOMI_CONFIG=/tmp/loomi-config.json go run ./cmd/loomi config unset host
go run ./cmd/loomi sessions list --host http://127.0.0.1:8080
go run ./cmd/loomi models list --host http://127.0.0.1:8080
go run ./cmd/loomi personas list --host http://127.0.0.1:8080
go run ./cmd/loomi tools list --host http://127.0.0.1:8080
go run ./cmd/loomi tools list --host http://127.0.0.1:8080 --group workspace --enabled-only
go run ./cmd/loomi run "Summarize the workspace status" --host http://127.0.0.1:8080 --mode work --provider local_codex
go run ./cmd/loomi run "Summarize the workspace status" --host http://127.0.0.1:8080 --mode work --provider local_codex --compact
go run ./cmd/loomi run "Summarize the workspace status" --host http://127.0.0.1:8080 --mode work --provider local_codex --interactive-approvals
printf 'Summarize the workspace status' | go run ./cmd/loomi run --prompt-file - --output json --host http://127.0.0.1:8080
go run ./cmd/loomi chat --host http://127.0.0.1:8080
# In chat, try /tools workspace, /approvals, and /events compact after a run.
go run ./cmd/loomi events tail <run-id> --host http://127.0.0.1:8080 --tools-only --compact
go run ./cmd/loomi approvals follow <run-id> --host http://127.0.0.1:8080
go run ./cmd/loomi approvals approve --follow --host http://127.0.0.1:8080 <thread-id> <run-id> <tool-call-id>
```

Expected evidence:

1. `internal/cli` creates a thread, appends a message, starts a run, consumes SSE, and returns terminal status.
2. `cmd/loomi` routes `sessions list`, `models list`, `personas list`, `tools list`, focused help, and approval decisions without touching runtime internals.
3. Event output includes run event type, sequence, and content/summary without raw secrets.
4. When a run is blocked on a tool call, the final output includes copyable `loomi approvals approve ...` and `loomi approvals deny ...` commands for unresolved approvals only.
5. Text event output shows compact tool `arguments_summary` and `result_summary` values for approval and success events, with common tools rendered as readable fields such as `path=...`, `exit=0`, `stdout="..."`, `links=2`, or `items=3` instead of raw JSON.
6. `loomi run --compact` uses the same short transcript renderer during a live run, while `--output json` and `--output stream-json` remain machine-readable.
7. Approval commands call the existing `/approve` and `/deny` endpoints.
8. `loomi approvals approve --follow` fetches the current last event sequence, applies the approval, then tails new events from `after_sequence` without replaying older output.
9. `loomi run --interactive-approvals` prompts for approve/deny/skip, uses the same approval endpoints, and reconnects with `after_sequence` after an approval decision.
10. `loomi chat` can prompt for approve/deny/skip inline when a run blocks on tool approval, then continue the same run after the decision.
11. `loomi chat` can show `/tools [group]`, `/approvals [run-id]`, and `/events [compact] [run-id]` without leaving the shell.
12. `loomi chat` can start a new thread, resume an existing thread through `sessions resume`, and switch model/persona with slash commands backed by server lists.
13. If the SSE connection closes before a terminal event or pending approval, `internal/cli.Runner` reconnects with `after_sequence` and does not duplicate already-rendered events.
14. `loomi config show` reads `~/.loomi/config.json` or `LOOMI_CONFIG`, then applies `LOOMI_HOST`, `LOOMI_MODE`, `LOOMI_PROVIDER`, `LOOMI_MODEL`, `LOOMI_PERSONA`, and `LOOMI_SCRIPT` overrides.
15. `loomi config set/unset` writes only the config file, creates it with `0600`, and leaves env overrides untouched.
16. `loomi approvals follow` filters the event stream down to tool approval notices and copyable approve/deny commands.
17. `loomi tools list` text output groups tools by catalog group and supports `--group`, `--enabled-only`, and `--flat`.
18. `loomi events tail --tools-only --compact` filters out model delta/final events and renders compact tool call state, arguments, and per-tool result summaries.

If `local_codex` is not enabled, `loomi run` should fail with the API error. Enable it from the existing provider settings/API before using the CLI for live local-agent dogfooding.
