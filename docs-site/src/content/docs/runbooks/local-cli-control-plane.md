---
title: Local CLI Control Plane Validation
description: Focused validation for cmd/loomi and internal/cli.
---

## Focused Checks

```bash
go test ./cmd/loomi ./internal/cli
go run ./cmd/loomi version
go run ./cmd/loomi completion bash
go run ./cmd/loomi completion zsh
go run ./cmd/loomi completion fish
scripts/build-cli.sh
dist/loomi version
TARGET=/tmp/loomi-bin/loomi scripts/install-cli.sh
/tmp/loomi-bin/loomi version
go run ./cmd/loomi help tools
go run ./cmd/loomi config show
go run ./cmd/loomi doctor
LOOMI_CONFIG=/tmp/loomi-config.json go run ./cmd/loomi config set host http://127.0.0.1:8080
LOOMI_CONFIG=/tmp/loomi-config.json go run ./cmd/loomi config unset host
go run ./cmd/loomi sessions list --host http://127.0.0.1:8080
go run ./cmd/loomi models list --host http://127.0.0.1:8080
go run ./cmd/loomi personas list --host http://127.0.0.1:8080
go run ./cmd/loomi tools list --host http://127.0.0.1:8080
go run ./cmd/loomi tools list --host http://127.0.0.1:8080 --group workspace --enabled-only
go run ./cmd/loomi mcp servers --host http://127.0.0.1:8080
go run ./cmd/loomi lsp tools --host http://127.0.0.1:8080
go run ./cmd/loomi artifacts list <thread-id> --host http://127.0.0.1:8080
go run ./cmd/loomi artifacts read <thread-id> <artifact-id> --host http://127.0.0.1:8080
go run ./cmd/loomi memory list --host http://127.0.0.1:8080
go run ./cmd/loomi memory search "workspace preference" --host http://127.0.0.1:8080
go run ./cmd/loomi memory show <memory-id> --host http://127.0.0.1:8080
go run ./cmd/loomi memory audit --thread-id <thread-id> --host http://127.0.0.1:8080
go run ./cmd/loomi agent tools --host http://127.0.0.1:8080
go run ./cmd/loomi agent tasks <thread-id> --host http://127.0.0.1:8080
go run ./cmd/loomi browser tools --host http://127.0.0.1:8080
go run ./cmd/loomi browser events <run-id> --host http://127.0.0.1:8080 --compact
go run ./cmd/loomi run "Summarize the workspace status" --host http://127.0.0.1:8080 --mode work --provider local_codex
LOOMI_HOST=http://127.0.0.1:18080 LOOMI_PROVIDER=custom LOOMI_MODE=chat go run ./cmd/loomi run --compact --timeout 30s '请只回复 pong'
go run ./cmd/loomi run "Summarize the workspace status" --host http://127.0.0.1:8080 --mode work --provider local_codex --compact
go run ./cmd/loomi run "Summarize the workspace status" --host http://127.0.0.1:8080 --mode work --provider local_codex --interactive-approvals
go run ./cmd/loomi runs status <run-id> --host http://127.0.0.1:8080
go run ./cmd/loomi runs stop <run-id> --host http://127.0.0.1:8080
go run ./cmd/loomi runs attach <active-run-id> --host http://127.0.0.1:8080 --compact
go run ./cmd/loomi runs attach <terminal-run-id> --host http://127.0.0.1:8080 --compact
go run ./cmd/loomi runs follow <run-id> --host http://127.0.0.1:8080 --compact
printf 'Summarize the workspace status' | go run ./cmd/loomi run --prompt-file - --output json --host http://127.0.0.1:8080
go run ./cmd/loomi chat --host http://127.0.0.1:8080
# In chat, try /tools workspace, /run, /approvals, /events compact, and /stop after a run.
go run ./cmd/loomi events tail <run-id> --host http://127.0.0.1:8080 --tools-only --compact
go run ./cmd/loomi approvals follow <run-id> --host http://127.0.0.1:8080
go run ./cmd/loomi approvals approve --follow --host http://127.0.0.1:8080 <thread-id> <run-id> <tool-call-id>
```

Expected evidence:

1. `internal/cli` creates a thread, appends a message, starts a run, consumes SSE, and returns terminal status.
2. `cmd/loomi` routes `sessions list`, `models list`, `personas list`, `tools list`, focused help, and approval decisions without touching runtime internals.
3. `loomi version` prints version, commit, and build date metadata.
4. `loomi completion bash|zsh|fish` prints completion scripts for the current command surface.
5. `scripts/build-cli.sh` writes a local `dist/loomi` binary, injects detected build metadata, and verifies it with `dist/loomi version`.
6. `scripts/install-cli.sh` installs the built binary to `~/.local/bin/loomi` by default, supports `PREFIX`/`TARGET`, and refuses to replace an existing target unless `LOOMI_INSTALL_OVERWRITE=1` is set.
7. `loomi doctor` reports config, API readiness, configured provider status, and tool catalog health; it exits non-zero when API readiness fails. If the default `local_codex` provider is selected but not registered by the API, doctor prints a specific warning and points the operator to `LOOMI_PROVIDER`, `loomi config set provider <id>`, and `loomi models list` without changing the config file.
8. Event output includes run event type, sequence, and content/summary without raw secrets.
9. When a run is blocked on a tool call, the final output includes copyable `loomi approvals approve ...` and `loomi approvals deny ...` commands for unresolved approvals only.
10. Text event output shows compact tool `arguments_summary` and `result_summary` values for approval and success events, with common tools rendered as readable fields such as `path=...`, `exit=0`, `stdout="..."`, `links=2`, or `items=3` instead of raw JSON.
11. `loomi run --compact` uses the same short transcript renderer during a live run, while `--output json` and `--output stream-json` remain machine-readable.
12. Approval commands call the existing `/approve` and `/deny` endpoints.
13. `loomi approvals approve --follow` fetches the current last event sequence, applies the approval, then tails new events from `after_sequence` without replaying older output.
14. `loomi run --interactive-approvals` prompts for approve/deny/skip, uses the same approval endpoints, and reconnects with `after_sequence` after an approval decision.
15. `loomi chat` can prompt for approve/deny/skip inline when a run blocks on tool approval, then continue the same run after the decision.
16. `loomi chat` can show `/tools [group]`, `/approvals [run-id]`, and `/events [compact] [run-id]` without leaving the shell.
17. `loomi runs status <run-id>` and chat `/run [run-id]` render the current run projection.
18. `loomi runs stop <run-id>` and chat `/stop [run-id]` call the existing stop endpoint and render stopped or already-terminal results.
19. `loomi runs attach <run-id>` renders the run projection and replays stored events. For active runs it continues streaming from the last replayed sequence; for terminal runs it exits after history replay and does not open a live stream.
20. `loomi runs follow <run-id>` defaults to future-only streaming by finding the current last sequence before opening SSE; `--after` can override the resume point.
21. `loomi chat` can start a new thread, resume an existing thread through `sessions resume`, and switch model/persona with slash commands backed by server lists.
22. If the SSE connection closes before a terminal event or pending approval, `internal/cli.Runner` reconnects with `after_sequence` and does not duplicate already-rendered events.
23. When `loomi run` has no `--thread`, `internal/cli.Runner` creates a Chat/Work thread with the requested mode and a stable non-empty title derived from the first non-empty prompt line, capped to a short excerpt before the user message and run are created.
24. `loomi config show` reads `~/.loomi/config.json` or `LOOMI_CONFIG`, then applies `LOOMI_HOST`, `LOOMI_MODE`, `LOOMI_PROVIDER`, `LOOMI_MODEL`, `LOOMI_PERSONA`, and `LOOMI_SCRIPT` overrides.
25. `loomi config set/unset` writes only the config file, creates it with `0600`, and leaves env overrides untouched.
26. `loomi approvals follow` filters the event stream down to tool approval notices and copyable approve/deny commands.
27. `loomi tools list` text output groups tools by catalog group and supports `--group`, `--enabled-only`, and `--flat`.
28. `loomi events tail --tools-only --compact` filters out model delta/final events and renders compact tool call state, arguments, and per-tool result summaries.
29. `loomi mcp servers` renders safe MCP status rows without raw command, args, env, secrets, or host paths.
30. `loomi lsp tools` filters the tool catalog to LSP tools and keeps execution behind Work-mode approval-gated runs.
31. `loomi artifacts list/read` uses thread-scoped read-only artifact endpoints and renders bounded text excerpts only.
32. `loomi memory list/search/show/audit` uses safe memory APIs and does not write, approve, deny, or delete memory.
33. `loomi agent tasks/tools` exposes coordination-only task state and catalog entries without launching child model runs.
34. `loomi browser tools/events` exposes browser tool catalog entries and browser run events without direct browser session control.

If `local_codex` is not enabled, `loomi run` should fail with the API error. Enable it from the existing provider settings/API before using the CLI for live local-agent dogfooding.
