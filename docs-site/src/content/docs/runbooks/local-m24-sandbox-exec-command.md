---
title: Local Bounded Command Validation
description: Local validation for approval-gated bounded sandbox.exec_command.
---

## M78 Sandbox Process Foundation

M78 extends the M24 command slice with the minimum controlled process lifecycle:

- `sandbox.start_process` starts an approved allowlisted argv command and returns `process_id`, `status`, `next_cursor`, bounded stdout/stderr previews, and safe lifecycle metadata.
- `sandbox.continue_process` polls later output by `cursor` and returns promptly when there is no new output.
- `sandbox.terminate_process` cancels a run-scoped process and records `terminal_summary`.

This remains a local host process foundation. It is not Docker, Firecracker, a shell, a PTY, or a general terminal service.

## Focused Checks

```bash
go test ./internal/productdata -run 'TestToolCatalogIncludesSandboxExecCommand|TestWorkspaceToolsOnlyEnabledForWorkModeRunContext'
go test ./internal/runtime -run 'TestSandboxExecCommand|TestToolBrokerExecutesSandboxExecCommand|TestWorkerExecutesApprovedSandboxExecCommand|TestGatewayRejectsSandboxExecCommand|TestWorkerDoesNotExecuteSandboxCommand'
go test ./internal/httpapi -run TestM24SandboxExecCommandSmoke
bun test --cwd web ./src/components/SettingsView.tools.test.tsx ./src/components/RunRail.runtime.test.ts
```

Expected evidence:

1. Tool catalog includes `sandbox.exec_command` as high-risk, approval-required, exec-capable, argv-only, validation-capable, not isolated, and `bounded_command` scoped.
2. Work mode can enable sandbox exec; Chat mode rejects it.
3. Approved execution runs one argv command and records bounded stdout/stderr and exit metadata.
4. Unapproved, denied, stopped, unsafe cwd, shell-form, sensitive path, destructive command, network command, arbitrary script, and oversized output paths stay bounded and auditable.
5. Settings and RunRail show bounded command risk/audit metadata.

## Process Tool Checks

```bash
go test ./internal/productdata ./internal/runtime -run 'TestValidateSandbox|TestToolCatalogIncludesSandbox|TestSandbox|TestToolBrokerExecutesSandbox|TestSandboxToolDefinitions|TestGatewayExposesEnabledBuiltinProviderTools'
go test ./internal/httpapi -run TestM24SandboxProcessLoopSmoke -count=1
bun test --cwd web ./src/components/RunRail.runtime.test.ts ./src/components/ToolCallCard.test.tsx
```

Expected evidence:

1. Tool catalog includes `sandbox.start_process`, `sandbox.continue_process`, and `sandbox.terminate_process` as high-risk, approval-required, exec-capable, argv-only, validation-capable, not isolated, and `bounded_process` scoped.
2. `sandbox.start_process` accepts only the bounded argv/cwd/timeout/output request shape, the same read/validation allowlist as one-shot exec, plus the narrow `stdin: true` process slice for argv-form `cat`.
3. `sandbox.continue_process` accepts `process_id`, optional `cursor`, and optional bounded stdin fields; `sandbox.terminate_process` accepts only `process_id`.
4. Process handles are run-scoped; another run cannot continue or terminate them.
5. Runtime results expose status, exit code, timeout, bounded stdout/stderr, byte counts, truncation flags, `next_cursor`, `stdin_open`, and `input_seq` without leaking host workspace roots.
6. HTTP smoke covers `start_process -> continue_process -> terminate_process` through separate approval/resume cycles and completes the run after provider continuation.
7. UI summaries label process lifecycle tools distinctly and show `process_id`, `next_cursor`, and `terminal_summary` without host paths or secret-looking output.

## M81 Process Lifecycle Recovery Checks

M81 focuses on process output/lifecycle/recovery usability, not on a new sandbox service.

```bash
go test ./internal/runtime -run 'TestSandboxProcessContinueCursorReadsBoundedLongOutput|TestSandboxProcessContinueAfterExitReturnsTerminalSummary|TestSandboxProcessContinueAfterTerminateOnlyReturnsSafeState|TestSandboxProcessOutputRedactionCoversPathsAndSecrets' -count=1
```

Expected evidence:

1. Long stdout remains bounded while `next_cursor` continues to advance with captured bytes.
2. Reusing the previous cursor reads only new retained output and does not replay old output.
3. Exited processes return `status=exited`, `exit_code`, and `terminal_summary` on continue.
4. Terminated processes can be continued only as a safe state read; stdin text and close requests do not perform a new action.
5. Cross-run process access stays rejected.
6. Absolute host paths, workspace roots, and secret-looking content are redacted in process output previews.

## Full Validation Target

```bash
go test ./...
bun test --cwd web
bun run --cwd web build
bun run --cwd docs-site build
git diff --check
```

Browser smoke must verify Settings Tools and RunRail bounded read-only command visibility with zero console errors.
