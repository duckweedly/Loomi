---
title: CLI Control Plane
description: Loomi command-line entrypoint for local run, event, tool, and approval workflows.
---

`cmd/loomi` is a thin local control plane over the existing HTTP API. It does not bypass product data, Gateway, Worker, ToolBroker, or approval checks. The current slice exists to dogfood sessions, model/persona selection, run/event streaming, and tool approval from a terminal before adding deeper local shell state.

## Boundary

```text
loomi CLI
-> internal/cli Client
-> /v1/threads, /v1/threads/{id}/messages, /v1/threads/{id}/runs
-> /v1/model-providers, /v1/personas, /v1/tools/catalog
-> /v1/runs/{id}/events/stream
-> /v1/threads/{thread}/runs/{run}/tool-calls/{tool}/approve|deny
```

The CLI keeps all durable state server-side. It only renders threads, personas, model providers, run events, final run IDs, tool catalog entries, and approval decisions. It has no direct database access and no direct runtime/tool executor access.

## Local Defaults

The CLI resolves defaults from `~/.loomi/config.json`, or from `LOOMI_CONFIG` when set. Supported fields are `host`, `mode`, `provider`, `model`, `persona`, and `script`. Environment variables `LOOMI_HOST`, `LOOMI_MODE`, `LOOMI_PROVIDER`, `LOOMI_MODEL`, `LOOMI_PERSONA`, and `LOOMI_SCRIPT` override the file for command execution.

`loomi config set <key> <value>` and `loomi config unset <key>` write only the config file, not environment variables. Writes create the parent directory with `0700` and the file with `0600`. Unsetting a key removes it from the file; command execution then falls back to built-in defaults or environment overrides.

## Commands

- `loomi status` checks `/readyz`.
- `loomi help [version|doctor|run|tools|mcp|lsp|events|runs|approvals|config]` prints focused command help.
- `loomi version` prints build metadata. Release builds can inject `version`, `commit`, and `date` through Go `-ldflags`.
- `scripts/build-cli.sh` builds the local binary into `dist/loomi`, injects `version`, `commit`, and `date`, then runs `dist/loomi version` as a sanity check. `VERSION`, `COMMIT`, `DATE`, and `OUTPUT` can override the detected values.
- `scripts/install-cli.sh` builds the binary and installs it to `~/.local/bin/loomi` by default. `PREFIX` or `TARGET` can change the target path. Existing targets are not replaced unless `LOOMI_INSTALL_OVERWRITE=1` is explicitly set.
- `loomi completion bash|zsh|fish` prints static shell completion scripts for the current command surface.
- `loomi doctor` checks resolved CLI config, API readiness, configured provider availability, and tool catalog availability without mutating state.
- `loomi config show` prints resolved defaults and whether the config file was found.
- `loomi config set <host|mode|provider|model|persona|script> <value>` persists one local default.
- `loomi config unset <host|mode|provider|model|persona|script>` removes one persisted local default.
- `loomi chat` opens an interactive shell over the same run/event loop. Slash commands are local controls: `/help`, `/status`, `/thread`, `/new`, `/model <provider-id-or-model>`, `/persona <id-or-slug>`, `/tools [group]`, `/run [run-id]`, `/approvals [run-id]`, `/events [compact] [run-id]`, `/stop [run-id]`, and `/quit`. When a run blocks on `tool_call_approval_required`, the shell prompts inline for approve, deny, or skip, then continues through the same approval endpoint and SSE resume path.
- `loomi sessions list` reads `/v1/threads`.
- `loomi sessions resume <thread-id>` opens `loomi chat` on an existing thread.
- `loomi models list` reads `/v1/model-providers`.
- `loomi personas list` reads `/v1/personas`.
- `loomi tools list` reads `/v1/tools/catalog`. Text output groups tools by `group` and shows execution state, approval policy, risk level, and enabled state. `--group`, `--enabled-only`, and `--flat` provide focused daily-driver views; JSON output remains the raw filtered catalog list.
- `loomi mcp servers` reads `/v1/mcp/servers` and renders safe local MCP status rows only: slug, display name, transport, enabled state, discovery status, candidate count/names, execution mode, redacted error code, and last discovery metadata. It does not print raw command, args, env, secrets, or host paths.
- `loomi lsp tools` reads `/v1/tools/catalog` and filters to LSP catalog entries. LSP execution still happens only through Work-mode approval-gated tool calls inside runs; the CLI command is a focused visibility surface, not a direct language-server client.
- `loomi artifacts list <thread-id>` and `loomi artifacts read <thread-id> <artifact-id>` read thread-scoped artifact projections from `/v1/threads/{thread}/artifacts`. They return bounded excerpts only; artifact creation remains an approval-gated Work-mode tool call.
- `loomi memory list`, `loomi memory search`, `loomi memory show`, and `loomi memory audit` read existing safe memory and audit APIs. They expose safe summaries/details only and do not create, approve, deny, or delete memories.
- `loomi agent tasks <thread-id>` reads `/v1/threads/{thread}/agent-tasks` and renders coordination-only task records. `loomi agent tools` filters the tool catalog to `agent.*`. The CLI does not start child model runs or external workers.
- `loomi browser tools` filters the tool catalog to `browser.*`. `loomi browser events <run-id>` filters a run event stream to browser tool lifecycle events. Browser execution still happens only through approval-gated Work-mode tool calls; the CLI does not directly control browser sessions.
- `loomi run <prompt>` creates a Work thread by default, appends a user message, starts a `model_gateway` run, and tails SSE until terminal or the next pending approval. It also supports `--prompt-file`, `--timeout`, `--thread`, `--provider`, `--model`, `--persona`, `--script`, `--compact`, `--interactive-approvals`, and `--output text|json|stream-json`.
- `loomi runs status <run-id>` reads the current run projection.
- `loomi runs stop <run-id>` delegates to the existing stop endpoint and renders the stopped/already-terminal result.
- `loomi runs attach <run-id>` prints the current run projection, replays persisted events after `--after` (default `0`), then follows live events from the last replayed sequence.
- `loomi runs follow <run-id>` tails live events only by default. Without `--after`, it first reads the persisted event list to find the current last sequence, then streams future events from that point.
- `loomi events tail <run-id>` streams persisted and live run events. `--tools-only` filters to tool-call events, `--compact` renders shorter one-line summaries, and `--output json` preserves a script-friendly stream.
- `loomi approvals list <run-id>` derives pending approvals from run events.
- `loomi approvals follow <run-id>` streams approval-focused notices and copyable approve/deny commands.
- `loomi approvals approve|deny [--follow] <thread-id> <run-id> <tool-call-id>` delegates to the existing tool-call decision endpoints. With `--follow`, the CLI reads the current last event sequence before the decision, applies the decision, then tails `/events/stream?after_sequence=<last>` so users can continue the run without manually starting `events tail`.

`loomi run` defaults to `provider_id=local_codex` so local dogfooding fails visibly when the bridge is not enabled instead of silently falling back to a mock.

When streamed events include unresolved `tool_call_approval_required` entries, `loomi run` prints the matching `loomi approvals approve ...` and `loomi approvals deny ...` commands in its final result. The pending list is folded from the event stream, so later approved, denied, executing, succeeded, failed, or cancelled tool-call events remove stale approval prompts. Text output renders compact `arguments_summary` and `result_summary` for tool events so terminal users can inspect what is about to run and what came back without switching to JSON. Common workspace, sandbox, browser, artifact, web, LSP, todo, and coordination tools get per-tool summaries such as `path=...`, `exit=0`, `stdout="..."`, `links=2`, or `items=3`; unknown tools still fall back to compact JSON. `loomi run --compact` uses the same short transcript renderer as `loomi events tail --compact`, keeping long local dogfood runs easier to scan while leaving JSON and stream-json unchanged.

With `--interactive-approvals`, `loomi run` prompts for `approve`, `deny`, or `skip` when a tool approval event arrives, then calls the same approval decision endpoint and continues the run event stream with `after_sequence`. The flag is text-output only and cannot be combined with `--prompt-file -`, because stdin is reserved for approval choices. `loomi chat` always uses the same inline approval prompt for interactive chat input.

## Runner Reconnect

`internal/cli.Runner` reconnects the event stream up to three times when the SSE connection closes before a terminal run event or pending approval. Reconnect uses `after_sequence` and de-duplicates event IDs, so streamed model deltas and tool events are not rendered twice during a short disconnect.

## Remaining Gap

The next CLI hardening slice should add broader tool batch coverage and richer resume controls for long interactive runs. The current scripts deliberately stop at local build/install flows; they do not publish release artifacts.
