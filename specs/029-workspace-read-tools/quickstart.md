# Quickstart: M21 Workspace Read Tools

## Backend Smoke

1. Create a fixture workspace with normal files, a large text file, `.env`, `secrets/token.txt`, and a symlink pointing outside the root.
2. Set `LOOMI_WORKSPACE_ROOT` to the fixture root.
3. Start a Work mode run with a provider tool call for `workspace.glob`.
4. Verify the run emits `tool_call_requested` and `tool_call_approval_required` before execution.
5. Approve the tool call.
6. Verify worker emits `tool_call_executing`, `tool_call_succeeded`, and continuation events.
7. Repeat for `workspace.read` and `workspace.grep`.
8. Verify traversal, absolute outside path, sensitive files, and symlink escape fail without leaking content.

## Required Validation

```bash
go test ./...
bun test --cwd web
bun run --cwd web build
bun run --cwd docs-site build
git diff --check
```

## Manual UI Smoke

1. Open Settings > Tools.
2. Confirm workspace tools appear under the workspace group with read-only and executable metadata.
3. Confirm no local absolute path is shown.
4. Open a timeline containing workspace tool events.
5. Confirm requested, approval-required, approved, executing, succeeded, and failed states are legible in Work/Chat timelines.
