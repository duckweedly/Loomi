---
title: 2026-05-26 M24 Bounded Read-only Command
description: Initial approval-gated sandbox.exec_command implementation notes.
---

M24 adds the first command execution primitive after workspace read and mutation tools.

Implemented:

- Added `specs/032-sandbox-exec-command-tools/` with spec, plan, research, data model, contract, quickstart, checklist, and task list.
- Added `sandbox.exec_command` as a builtin high-risk, approval-required, exec-capable bounded read-only command tool. The current slice is not an isolated sandbox.
- Kept sandbox exec Work-mode-only through RunContext enabled tools and Gateway rejection.
- Implemented argv-form execution under the configured workspace root with bounded timeout and stdout/stderr previews.
- Limited the executable slice to `pwd`, `ls`/`ls .`, and `git status`; file-reading commands stay disabled until they can share the workspace resolver and sensitive-path checks.
- Rejected shell-form commands, model-supplied env, out-of-scope cwd, sensitive cwd, path-bearing `ls`, file-reading commands, and destructive command patterns before spawn.
- Added ToolBroker, worker, gateway, and HTTP smoke coverage for approval-gated execution and provider continuation.
- Added Settings and RunRail visibility for bounded read-only command risk and audit metadata.

Later CLI/readiness update:

- Expanded the allowlist from the original `pwd`/`ls`/`git status` slice to the minimum code-agent read and validation set: `cat`, `head`, `tail`, `sed -n`, `wc`, `rg`, read-only git subcommands, `go test`, and test/build commands for Bun/npm/pnpm.
- Changed the advertised scope from `bounded_read_only_command` to `bounded_command` and added `validation_capable`; the tool remains high-risk, approval-required, argv-only, workspace-root scoped, and non-isolated.
- Kept shell-form commands, destructive commands, network clients, package install commands, hidden-search flags, sensitive paths, and output-writing validation flags rejected before spawn.

Later process-control update:

- Added `sandbox.start_process`, `sandbox.continue_process`, and `sandbox.terminate_process` as builtin high-risk, approval-required sandbox group tools.
- Kept process start on the same argv-only read/validation allowlist and workspace-root cwd boundary as `sandbox.exec_command`.
- Added an in-memory run-scoped process store so continuation/termination requires the originating `run_id` and `process_id`.
- Returned bounded stdout/stderr snapshots, `running`/`exited`/`terminated` status, exit code, timeout, byte counts, and truncation metadata for process tools.
- Kept the boundary explicit: no persistent shell, PTY, isolated container, syscall sandbox, or arbitrary background process manager.

Later stdin/cursor readiness update:

- Added `cursor`/`next_cursor` support so `sandbox.continue_process` can return incremental stdout snapshots instead of forcing full-buffer polling.
- Added a narrow stdin-enabled process path: `sandbox.start_process` may set `stdin: true` only for argv-form `cat`, and `sandbox.continue_process` can write bounded `stdin_text` with increasing `input_seq`.
- Added `close_stdin`, `stdin_open`, and `input_seq` result metadata for CLI-style process driving without introducing a shell or PTY.

Focused validation run during implementation:

```bash
go test ./internal/productdata -run 'TestToolCatalogIncludesSandboxExecCommand|TestWorkspaceToolsOnlyEnabledForWorkModeRunContext'
go test ./internal/runtime -run 'TestSandboxExecCommand|TestToolBrokerExecutesSandboxExecCommandThroughOneEntrypoint|TestWorkerExecutesApprovedSandboxExecCommandAndContinuesModel|TestGatewayRejectsSandboxExecCommandInChatMode|TestWorkerDoesNotExecuteSandboxCommandAfterStopOrDenied'
go test ./internal/httpapi -run TestM24SandboxExecCommandSmoke
bun test --cwd web ./src/components/SettingsView.tools.test.tsx ./src/components/RunRail.runtime.test.ts
```

Process-tool focused validation:

```bash
go test ./internal/productdata ./internal/runtime -run 'TestValidateSandbox|TestToolCatalogIncludesSandbox|TestSandbox|TestToolBrokerExecutesSandbox|TestSandboxToolDefinitions|TestGatewayExposesEnabledBuiltinProviderTools'
go test ./internal/httpapi -run TestM24SandboxProcessLoopSmoke -count=1
```

Stdin/cursor focused validation:

```bash
go test ./internal/productdata ./internal/runtime -run 'TestValidateSandbox|TestSandboxProcess|TestToolCatalogIncludesSandbox|TestGatewayExposesEnabledBuiltinProviderTools' -count=1
go test ./internal/httpapi -run TestM24SandboxProcessLoopSmoke -count=1
```

HTTP process smoke coverage:

- Starts a Work-mode run with a persona limited to `sandbox.start_process`, `sandbox.continue_process`, and `sandbox.terminate_process`.
- Uses HTTP approval for each tool call and worker resume for each approved step.
- Runs a FIFO-backed allowlisted `cat stream.txt` process, writes input after the process has started, polls output through `sandbox.continue_process`, and terminates through `sandbox.terminate_process`.
- Verifies `bounded_process` result scope, process output, loop metadata, completion event, and host-root redaction.

Full validation:

```bash
go test ./...
bun test --cwd web
bun run --cwd web build
bun run --cwd docs-site build
git diff --check
```

Browser smoke:

- in-app Browser `iab` was unavailable, so local Playwright was used against `http://127.0.0.1:5178/`.
- Settings > Tools showed `sandbox.exec_command`, `bounded_command scope`, `exec-capable`, `validation-capable`, and `high`.
- RunRail showed bounded read-only command lifecycle rows for waiting and completed tool calls.
- Console errors: 0.
