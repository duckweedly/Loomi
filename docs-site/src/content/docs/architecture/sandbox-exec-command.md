---
title: Bounded Command
description: Approval-gated bounded read and validation command execution through the Loomi tool runtime.
---

`sandbox.exec_command` is the first code-agent command primitive after workspace read and mutation tools. The tool name is historical: this slice is not an isolated sandbox. It is intentionally narrow: argv-form command, Work-mode only, explicit approval, relative workspace cwd, bounded read/validation command allowlist, bounded timeout, bounded stdout/stderr previews, and no shell session.

The follow-up process slice adds `sandbox.start_process`, `sandbox.continue_process`, and `sandbox.terminate_process` for short-lived allowlisted commands that need output polling, bounded stdin, or explicit termination. These tools share the same safety boundary as `sandbox.exec_command`; they do not create a persistent shell, PTY, container, or general background worker.

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

Process tools use an in-memory run-scoped process store:

- `sandbox.start_process` starts one allowlisted argv command and returns a `process_id`, status, cwd, exit metadata, and bounded stdout/stderr previews. `stdin: true` is supported only for the narrow stdin process slice, currently argv-form `["cat"]`.
- `sandbox.continue_process` verifies the same `run_id`, returns the current status/output snapshot, supports stdout cursor polling through `cursor`/`next_cursor`, and can write bounded `stdin_text` when paired with a monotonically increasing `input_seq`.
- `sandbox.terminate_process` accepts only `process_id`, verifies the same `run_id`, cancels the process, waits briefly, and kills it only if it does not exit.
- Process timeouts are bounded independently from one-shot exec commands and are capped at 120 seconds.

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
- Normal run events do not include raw env values, host absolute roots, provider raw payloads, or hidden local state.

## Current Limitations

- No true streaming stdout/stderr; process output is polling-based through `sandbox.continue_process` cursor snapshots.
- No persistent shell session or PTY.
- No container sandbox, syscall isolation, or per-command filesystem snapshot yet.
- The tool still runs on the host workspace, so package scripts require approval and remain high-risk.
