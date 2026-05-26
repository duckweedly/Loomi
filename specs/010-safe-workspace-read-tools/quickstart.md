# Quickstart: M8 Safe Workspace Read Tools

## Local Validation Flow

1. Start the API and web app with the same commands used for M7.
2. Trigger a controlled tool request for `workspace.glob`.
3. Verify the UI shows approval-required state.
4. Approve the tool call.
5. Verify the result shows bounded relative paths.
6. Repeat for `workspace.grep` and `workspace.read_file`.
7. Attempt unsafe paths such as `../`, `.env`, `.ssh/id_ed25519`, `secrets/token.txt`, and absolute home paths; verify safe failures.
8. Reconnect or reload the UI and verify history-first replay preserves tool states.

## Validation Commands

```bash
go test ./...
bun test --cwd web
bun run --cwd web build
(cd docs-site && bun run build)
git diff --check
```

## Browser Smoke

- Open Settings/Runtime or the run detail view used by M7.
- Start or replay a run containing each M8 read tool.
- Approve one call and deny one call.
- Confirm ToolCallCard, RunRail, and Timeline show distinct read-tool lifecycle states.
- Confirm the browser console has no errors.
