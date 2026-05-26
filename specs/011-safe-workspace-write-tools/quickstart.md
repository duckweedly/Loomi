# Quickstart: M9 Safe Workspace Write Tools

## Local Validation Flow

1. Trigger a controlled `workspace.write_file` request for a safe file under the workspace.
2. Verify approval-required state.
3. Deny one request and verify no file mutation.
4. Approve one request and verify the file is written once.
5. Trigger `workspace.edit` on a file where `old_text` occurs exactly once.
6. Approve it and verify one replacement.
7. Try missing-match, duplicate-match, path traversal, symlink escape, `.env`, `.ssh`, `secrets/`, and missing parent directory cases; verify safe failures without mutation.
8. Reload the UI and verify history-first replay preserves write/edit tool states.

## Validation Commands

```bash
go test ./...
bun test --cwd web
bun run --cwd web build
(cd docs-site && bun run build)
git diff --check
```

## Browser Smoke

- Load a run containing workspace write/edit tool lifecycle events.
- Confirm ToolCallCard, RunRail, and Timeline show requested, approval-required, executing, and terminal states.
- Confirm write/edit summaries are readable and do not expose raw secrets.
- Confirm browser console has no errors.
