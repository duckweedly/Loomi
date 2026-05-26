---
title: Local M24 Bounded Read-only Command Validation
description: Local validation for approval-gated bounded read-only sandbox.exec_command.
---

## Focused Checks

```bash
go test ./internal/productdata -run 'TestToolCatalogIncludesSandboxExecCommand|TestWorkspaceToolsOnlyEnabledForWorkModeRunContext'
go test ./internal/runtime -run 'TestSandboxExecCommand|TestToolBrokerExecutesSandboxExecCommand|TestWorkerExecutesApprovedSandboxExecCommand|TestGatewayRejectsSandboxExecCommand|TestWorkerDoesNotExecuteSandboxCommand'
go test ./internal/httpapi -run TestM24SandboxExecCommandSmoke
bun test --cwd web ./src/components/SettingsView.tools.test.tsx ./src/components/RunRail.runtime.test.ts
```

Expected evidence:

1. Tool catalog includes `sandbox.exec_command` as high-risk, approval-required, exec-capable, argv-only, read-only, not isolated, and `bounded_read_only_command` scoped.
2. Work mode can enable sandbox exec; Chat mode rejects it.
3. Approved execution runs one argv command and records bounded stdout/stderr and exit metadata.
4. Unapproved, denied, stopped, unsafe cwd, shell-form, file-reading commands, destructive command, and oversized output paths stay bounded and auditable.
5. Settings and RunRail show bounded read-only command risk/audit metadata.

## Full Validation Target

```bash
go test ./...
bun test --cwd web
bun run --cwd web build
bun run --cwd docs-site build
git diff --check
```

Browser smoke must verify Settings Tools and RunRail bounded read-only command visibility with zero console errors.
