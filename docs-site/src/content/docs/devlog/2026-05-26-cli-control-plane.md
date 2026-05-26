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

Later run stop update:

- Added `loomi runs status <run-id>` and chat `/run [run-id]` for current run status.
- Added `loomi runs stop <run-id>` and chat `/stop [run-id]`.
- Stop commands call the existing `/v1/runs/{runID}/stop` endpoint and render stopped/already-terminal results; no new runtime stop semantics were introduced.

Later run attach/follow update:

- Added `loomi runs attach <run-id>` for disconnected long-run resume: it renders the current run projection, replays persisted events after `--after`, and then streams new events from the last replayed sequence.
- Added `loomi runs follow <run-id>` for future-only tailing: by default it reads the current last event sequence and starts SSE after that point; `--after` can override the resume point.
- Reused the same compact/tools-only/json event rendering path as `loomi events tail`; no new backend endpoint or runtime semantics were introduced.

Later MCP/LSP visibility update:

- Added `loomi mcp servers` over the existing `/v1/mcp/servers` endpoint. Text output includes only safe status metadata and discovered candidate names; raw command, args, env, secrets, and host paths stay hidden.
- Added `loomi lsp tools` as a focused view over `/v1/tools/catalog` filtered to LSP tools.
- Kept LSP as approval-gated Work-mode tool execution inside runs; the CLI does not become a direct language-server client.

Later artifact visibility update:

- Added thread-scoped read-only artifact projection endpoints: `GET /v1/threads/:thread_id/artifacts` and `GET /v1/threads/:thread_id/artifacts/:artifact_id`.
- Added `loomi artifacts list <thread-id>` and `loomi artifacts read <thread-id> <artifact-id>` over those endpoints.
- Kept artifact creation behind approval-gated Work-mode tool calls; CLI reads bounded safe excerpts and does not expose a create/update/delete HTTP path.

Later memory visibility update:

- Added `loomi memory list`, `loomi memory search`, `loomi memory show`, and `loomi memory audit` over the existing safe memory APIs.
- Commands support scope/source filters and JSON output for scripting.
- Kept mutation out of the CLI slice: no memory write-proposal, approve, deny, delete, or direct create command was added.

Later agent/browser visibility update:

- Added `GET /v1/threads/:thread_id/agent-tasks` as a read-only projection of coordination-only agent task records.
- Added `loomi agent tasks <thread-id>` and `loomi agent tools`.
- Added `loomi browser tools` and `loomi browser events <run-id>` for browser catalog/event visibility.
- Kept both surfaces bounded: no direct child model execution, external worker launch, or browser session control was added.

Later doctor update:

- Added `loomi doctor` for read-only local health checks across resolved config, API readiness, configured provider status, and tool catalog availability.
- Doctor returns non-zero when API readiness fails, so scripts can detect a broken local control plane before starting dogfood runs.

Later version update:

- Added `loomi version` with text and JSON output.
- Release builds can inject `version`, `commit`, and `date` via Go `-ldflags`; dev builds report `dev`, `unknown`, and `unknown`.

Later local build update:

- Added `scripts/build-cli.sh` to build the local CLI binary into `dist/loomi`.
- The script detects git version and commit metadata, stamps UTC build time, supports `VERSION`, `COMMIT`, `DATE`, and `OUTPUT` overrides, and runs `dist/loomi version` after building.
- Added `scripts/install-cli.sh` for a local install path. It defaults to `~/.local/bin/loomi`, supports `PREFIX` and `TARGET`, and refuses to replace an existing target unless `LOOMI_INSTALL_OVERWRITE=1` is set.
- Added `loomi completion bash|zsh|fish` for shell completion scripts.
- Kept the scripts local-only: they do not publish release artifacts.

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
- Release packaging around the CLI binary.
