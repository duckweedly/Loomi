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

Focused validation run during implementation:

```bash
go test ./internal/productdata -run 'TestToolCatalogIncludesSandboxExecCommand|TestWorkspaceToolsOnlyEnabledForWorkModeRunContext'
go test ./internal/runtime -run 'TestSandboxExecCommand|TestToolBrokerExecutesSandboxExecCommandThroughOneEntrypoint|TestWorkerExecutesApprovedSandboxExecCommandAndContinuesModel|TestGatewayRejectsSandboxExecCommandInChatMode|TestWorkerDoesNotExecuteSandboxCommandAfterStopOrDenied'
go test ./internal/httpapi -run TestM24SandboxExecCommandSmoke
bun test --cwd web ./src/components/SettingsView.tools.test.tsx ./src/components/RunRail.runtime.test.ts
```

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
- Settings > Tools showed `sandbox.exec_command`, `bounded_read_only_command scope`, `exec-capable`, `read-only`, and `high`.
- RunRail showed bounded read-only command lifecycle rows for waiting and completed tool calls.
- Console errors: 0.
