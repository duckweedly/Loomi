---
title: 2026-05-26 CLI Control Plane
description: Initial loomi CLI slice for run, event, tool, and approval workflows.
---

Implemented:

- Added `cmd/loomi` as the first general Loomi CLI entrypoint.
- Added `internal/cli` HTTP client, SSE reader, run runner, and text/JSON rendering helpers.
- Supported `status`, `tools list`, `run`, `events tail`, `approvals list`, `approvals approve`, and `approvals deny`.
- Added first daily-driver commands: `chat`, `sessions list`, `sessions resume`, `models list`, and `personas list`.
- Added `loomi run --prompt-file`, stdin prompt support with `--prompt-file -`, `--timeout`, and `--output text|json|stream-json`.
- Added a small interactive chat shell with `/status`, `/thread`, `/new`, `/model`, `/persona`, and `/quit`.
- Added runner SSE reconnect with `after_sequence` and event de-duplication when a stream closes before terminal status or pending approval.
- Added read-only CLI defaults from `~/.loomi/config.json`, `LOOMI_CONFIG`, and `LOOMI_*` env overrides, plus `loomi config show`.
- Added `loomi approvals follow` for approval-focused streaming notices.
- Made `loomi run` fold streamed tool-call events into unresolved approval prompts and print copyable approve/deny commands at the end of the run.
- Made pending approval derivation ignore stale approval-required events once the same tool call is approved, denied, executing, succeeded, failed, or cancelled.
- Kept CLI execution behind existing HTTP APIs and approval endpoints; the CLI does not invoke ToolBroker or database code directly.
- Added tests for run creation/SSE consumption, tool-call approval decisions, and command routing.
- Expanded `sandbox.exec_command` from the original tiny read-only slice to a bounded code-agent read/validation allowlist while keeping approval, argv-only, workspace scope, and high-risk metadata.

Later daily-driver update:

- Text event output now renders compact tool `arguments_summary` on requested/approval events and compact `result_summary` on succeeded events.
- Added `loomi approvals approve --follow` and `loomi approvals deny --follow`; the CLI captures the current last event sequence before applying the decision, then streams only later events from the same run.
- Kept `--follow` on the HTTP/SSE path only; it does not bypass approval, worker, Gateway, ToolBroker, or run event persistence.

Later config update:

- Added `loomi config set <key> <value>` and `loomi config unset <key>` for `host`, `mode`, `provider`, `model`, `persona`, and `script`.
- Config writes use the file selected by `LOOMI_CONFIG` or `~/.loomi/config.json`, create parent directories with `0700`, and write the JSON file with `0600`.
- Kept environment overrides read-only and higher priority than the file.

Later interactive approval update:

- Added `loomi run --interactive-approvals` for text-mode runs.
- The Runner now accepts an approval callback, prompts for approve/deny/skip, calls the existing approval decision endpoint, and reconnects via `after_sequence` when the stream closes after a decision.
- Kept stdin prompt mode separate from `--prompt-file -`, because approval choices also need stdin.

Later help/tools update:

- Added `loomi help [run|tools|approvals|config]` for compact daily-driver command help.
- Changed text `loomi tools list` to group tools by catalog group and show execution state, approval policy, risk level, and enabled state.
- Added `loomi tools list --group <name>`, `--enabled-only`, and `--flat`; JSON output remains filtered raw catalog data.

Later event-tail update:

- Added `loomi events tail --tools-only` to filter the SSE stream down to tool call events.
- Added `loomi events tail --compact` for short one-line event summaries, including tool state, tool name, tool call id, and compact args/result/error metadata.
- Kept `--output json` available for script-friendly event streaming.

Later per-tool rendering update:

- Added readable text summaries for common workspace, sandbox, browser, artifact, web, LSP, todo, and coordination tool args/results.
- Kept JSON and stream-json output unchanged; unknown tool metadata still falls back to compact JSON.

Later run transcript update:

- Added `loomi run --compact` so live runs can use the same short renderer as `events tail --compact`.
- Kept compact mode text-only; JSON and stream-json remain stable for scripts.

Later chat approval update:

- Added inline approve/deny/skip prompts to `loomi chat` when a run reaches `tool_call_approval_required`.
- Reused the existing approval decision endpoint and Runner `after_sequence` continuation path; chat still does not bypass ToolBroker, worker, or persisted run events.

Later chat controls update:

- Added `loomi chat` slash commands `/tools [group]`, `/approvals [run-id]`, and `/events [compact] [run-id]`.
- Defaults use the most recent chat run id, so users can inspect tools, pending approvals, and compact event history without leaving the shell.

Focused validation:

```bash
go test ./cmd/loomi ./internal/cli
go test ./internal/runtime -run 'TestSandboxExecCommand'
```

Daily-driver focused validation:

```bash
go test ./cmd/loomi ./internal/cli -count=1
```

Next CLI gaps:

- Tool batch coverage beyond the current workspace/todo/patch/sandbox/web slices.
- Richer long-running run controls such as stop/resume once the backing APIs are ready for CLI use.
