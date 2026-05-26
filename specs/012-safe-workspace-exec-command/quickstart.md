# Quickstart: M10 Safe Workspace Exec Command

## Local Validation Flow

1. Trigger a controlled `workspace.exec_command` request such as `["printf", "hello"]`.
2. Confirm approval-required state.
3. Deny one request and confirm no execution.
4. Approve one request and confirm exit code and bounded stdout/stderr result.
5. Try cwd escape, shell wrappers, destructive commands, timeout, and large output cases.
6. Reload the UI and verify history-first replay preserves exec command states.

## Validation Commands

```bash
go test ./...
bun test --cwd web
bun run --cwd web build
(cd docs-site && bun run build)
git diff --check
```
