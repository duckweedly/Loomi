# Quickstart: M24 Sandbox Exec Command Tools

## Focused Validation

```bash
go test ./internal/productdata -run 'TestToolCatalogIncludesSandboxExecCommand|TestSandboxToolsOnlyEnabledForWorkModeRunContext'
go test ./internal/runtime -run 'TestSandboxExecCommand|TestToolBrokerExecutesSandboxExecCommand|TestWorkerExecutesApprovedSandboxExecCommand'
go test ./internal/httpapi -run TestM24SandboxExecCommandSmoke
bun test --cwd web ./src/components/SettingsView.tools.test.tsx ./src/components/RunRail.runtime.test.ts
```

Expected evidence:

1. Catalog includes `sandbox.exec_command` as high-risk, approval-required, exec-capable, not isolated, and `bounded_read_only_command` scoped.
2. Work mode can enable sandbox exec; Chat mode rejects it.
3. Approved execution runs one safe argv command and records bounded output.
4. Unapproved, denied, stopped, unsafe cwd, shell-form, file-reading command, path-bearing `ls`, destructive command, and oversized output paths do not violate safety boundaries.
5. Settings and RunRail show sandbox exec risk/audit metadata.

## Full Validation Target

```bash
go test ./...
bun test --cwd web
bun run --cwd web build
bun run --cwd docs-site build
git diff --check
```

Browser smoke must verify Settings Tools and RunRail sandbox exec visibility with zero console errors.
