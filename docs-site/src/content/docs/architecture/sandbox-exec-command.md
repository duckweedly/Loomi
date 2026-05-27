---
title: Bounded Command
description: Approval-gated bounded read and validation command execution through the Loomi tool runtime.
---

`sandbox.exec_command` is the first code-agent command primitive after workspace read and mutation tools. The tool name is historical: this slice is not an isolated sandbox. It is intentionally narrow: argv-form command, Work-mode only, explicit approval, relative workspace cwd, bounded read/validation command allowlist, bounded timeout, bounded stdout/stderr previews, and no shell session.

The follow-up process slice adds `sandbox.start_process`, `sandbox.continue_process`, and `sandbox.terminate_process` for short-lived allowlisted commands that need output polling, bounded stdin, or explicit termination. These tools share the same safety boundary as `sandbox.exec_command`; they do not create a persistent shell, PTY, container, or general background worker.

## Arkloop Comparison

Arkloop separates a sandbox service, guest agent, session manager, warm pool, process controller, shell controller, and tool-runtime snapshot. Its process path can route through Docker or Firecracker-style sessions, keeps per-session process refs, reads stdout/stderr through bounded buffers, supports follow/continue/terminate actions, and tears processes down with session lifecycle hooks. Its config layer describes sandbox templates and runtime tiers.

M78 borrows only the mechanism shape: process ids, run/session ownership, bounded output snapshots, explicit continue/terminate actions, and terminal summaries. M81 tightens the next small gap by making stdout cursor semantics stable across long output: `next_cursor` is an absolute captured-byte cursor, the retained preview is bounded to the latest safe tail, and terminal `continue_process` calls return stored state without attempting new stdin writes. M93 adds the service-before-service layer for in-process restore: process handles are snapshotted to a repository, restored into a rebuilt registry, reconciled when the OS process is gone, expired through TTL cleanup, and exposed through safe UI summaries. The default repository is memory-backed, so this does not yet provide API-restart durability. Loomi does not copy Arkloop's config format, tier names, agent protocol, endpoint names, copy, or expression layer.

M93 deliberately does not implement Arkloop's full isolation model. There is no Firecracker microVM, Docker pool, guest user, network namespace, filesystem snapshot sync, artifact extraction, shell checkpoint, PTY resize, browser tier, warm pool, or sandbox template runtime. The current Loomi foundation is a local process registry around host `exec.Cmd`, backed by a minimal in-process process record repository and guarded by Work-mode tool availability, explicit approval, argv-only validation, workspace cwd resolution, timeout/output bounds, and event redaction.

## Runtime Boundary

```text
provider tool request
-> tool_call_requested / approval_required
-> explicit approval
-> worker resume
-> ToolBroker
-> SandboxToolExecutor
-> tool_call_succeeded or tool_call_failed
-> provider continuation
```

The tools are enabled by the same RunContext snapshot used by workspace tools. Chat-mode runs remove sandbox tools from enabled tools, and Gateway rejects sandbox requests that are not present in the enabled-tool snapshot.

## Execution Model

The executor accepts only argv arrays. It resolves `cwd` through the existing workspace root boundary, rejects absolute/traversal/sensitive cwd values, rejects unlisted commands, and executes the command with a bounded context timeout. Output is captured after completion, clipped to configured byte limits, converted to safe UTF-8, and scrubbed of the absolute workspace root before event persistence.

Non-zero process exits are successful tool executions with `exit_code` metadata; validation or spawn failures become `tool_call_failed`.

Process tools use a run-scoped process store with a small in-process restore boundary:

- `sandbox.start_process` starts one allowlisted argv command and returns a `process_id`, status, cwd, exit metadata, and bounded stdout/stderr previews. `stdin: true` is supported only for the narrow stdin process slice, currently argv-form `["cat"]`.
- `sandbox.continue_process` verifies the same `run_id`, returns the current status/output snapshot, supports stdout cursor polling through absolute captured-byte `cursor`/`next_cursor`, and can write bounded `stdin_text` when paired with a monotonically increasing `input_seq`.
- `sandbox.terminate_process` accepts only `process_id`, verifies the same `run_id`, cancels the process, waits briefly, and kills it only if it does not exit.
- The in-process repository stores `run_id`, `process_id`, safe argv summary, cwd alias, status, cursor, started/updated/ended timestamps, exit code, bounded stdout/stderr tail, stdin state, and terminal summary.
- On registry rebuild, terminal records remain readable but cannot accept stdin or close mutations. A restored `running` record without a live local command is marked `failed` with a safe terminal summary instead of starting a replacement process.
- TTL reconciliation marks over-lifetime or idle records `expired`. Cleanup only touches processes that are present in the registry and owned by the originating run; Loomi does not scan or kill unrelated host processes.
- Terminal process results include `terminal_summary` for the UI audit trail.
- Process timeouts are bounded independently from one-shot exec commands and are capped at 120 seconds.
- Long stdout does not grow memory without limit. The process buffer keeps a bounded latest tail while byte counts and `next_cursor` continue to advance. If a caller resumes from a cursor older than the retained window, the response starts from the retained tail and reports truncation.

## Safety Rules

- Shell-form commands are rejected.
- Model-supplied env maps are rejected.
- Read commands are limited to `pwd`, `ls`, `cat`, `head`, `tail`, `sed -n`, `wc`, `rg`, and read-only git subcommands.
- Validation commands are limited to `go test`, `bun test`, `bun run build`, `npm test`, `npm run build`, `pnpm test`, and `pnpm run build`.
- Python, Node, brew, curl/wget, ssh, shells, chmod, kill, rm, package install commands, network clients, write-capable git commands, and arbitrary script execution are rejected before spawn.
- Absolute and traversal path arguments are rejected before spawn.
- Denied, stopped, terminal, duplicate, or out-of-scope tool calls do not execute.
- Process handles are scoped to the originating run; another run cannot continue or terminate them.
- Stdin writes require an explicitly stdin-enabled process, bounded text, and increasing `input_seq`; duplicate, closed, or non-stdin process writes fail before changing process state.
- Once a process has exited, been terminated, failed, or expired, `sandbox.continue_process` is state-only. Read-only continue returns the stored terminal status; stdin or close requests are rejected rather than revived, replayed, or silently applied.
- Terminal runs cannot start new sandbox processes, and a terminal run can only read an existing process handle without mutation.
- Normal run events do not include raw env values, host absolute roots, provider raw payloads, or hidden local state.
- Output previews redact host absolute paths and secret-looking content before entering tool results and UI summaries.

## Current Limitations

- No true streaming stdout/stderr; process output is polling-based through `sandbox.continue_process` cursor snapshots.
- No persistent shell session or PTY.
- No container sandbox, syscall isolation, or per-command filesystem snapshot yet.
- No productdata/Postgres process repository yet; the default memory repository loses process records when the API process restarts.
- The tool still runs on the host workspace, so package scripts require approval and remain high-risk.
