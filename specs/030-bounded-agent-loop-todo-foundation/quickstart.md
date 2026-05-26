# Quickstart: M22 Bounded Agent Loop + Todo Foundation

## Backend Smoke

1. Create a fixture workspace with two safe text files.
2. Start a Work mode run with a provider fixture that requests `workspace.glob`.
3. Verify the first tool call emits `tool_call_requested` and `tool_call_approval_required`.
4. Approve the first tool call and process the worker.
5. Verify `tool_call_succeeded` and continuation.
6. Provider fixture requests `workspace.read` during continuation.
7. Verify the second tool call pauses for a new approval and does not execute before approval.
8. Approve the second tool call and process the worker.
9. Verify final assistant message and `run_completed`.
10. Repeat with a fixture that exceeds the loop limit and verify safe `tool_loop_limit_reached` failure.

## UI Smoke

1. Start local API and web in real API mode with the backend smoke seed or equivalent fixture.
2. Open a Work mode thread with loop/todo events.
3. Confirm Work Plan View shows todo items and status changes.
4. Confirm RunRail/Timeline shows first tool, second tool, continuation, loop-limit, failed, and completed states.
5. Confirm Chat mode does not show Work todo state.
6. Confirm no raw file content, host absolute path, secret-looking value, shell command, browser state, or provider payload is visible.

## Required Validation

```bash
go test ./...
bun test --cwd web
bun run --cwd web build
bun run --cwd docs-site build
git diff --check
```
