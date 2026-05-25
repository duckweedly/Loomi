# Quickstart: M7 Tool Call Approval Core

This quickstart defines the intended local smoke path for implementing M7. It is planning guidance; commands should be verified during implementation.

## 1. Preconditions

- M6 worker job pipeline from latest `origin/main` is present.
- Local PostgreSQL is running and migrations through M6 pass readiness.
- Model gateway configuration can run with either a fake provider/test harness or a real provider that can emit a tool request.
- The only executable M7 tool is `runtime.get_current_time`.
- No shell, file read/write, arbitrary network, MCP, browser automation, multi-agent, or memory/RAG tools are enabled.

## 2. Migration and readiness smoke

Expected implementation validation:

```bash
go test ./internal/db ./internal/productdata
```

Manual checks:

1. Apply migrations through `000006_m7_tool_call_approval`.
2. Verify readiness reports the M7 schema version.
3. Roll back and reapply the migration in a local disposable database.
4. Verify `tool_calls` or equivalent projection can enforce unique `(run_id, tool_call_id)`.

## 3. Fake-provider tool request smoke

Use a fake provider or test-only model path that emits a single request:

```json
{
  "tool_call_id": "tc_m7_time_001",
  "tool_name": "runtime.get_current_time",
  "arguments": {
    "timezone": "UTC"
  }
}
```

Expected results:

1. `timezone` is either omitted or `UTC`; omitted defaults to `UTC`, and any other value fails schema validation safely.
2. Run is accepted through existing thread/run start path.
3. History-first SSE emits normal run/model setup events.
4. `tool_call_requested` is persisted with safe metadata.
5. `tool_call_approval_required` is persisted.
6. Current tool-call projection shows `approval_status = required` and `execution_status = blocked`.
7. Worker does not execute the tool while approval is pending.
8. Browser reconnect replays both tool events in order.

## 4. Approve idempotency smoke

Call the approve endpoint repeatedly for the same tool call.

Expected results:

1. First approve records exactly one `tool_call_approved` event.
2. Repeated approve calls return the same approved/current state.
3. Existing M6 worker pipeline resumes or wakes exactly once.
4. Tool execution emits exactly one `tool_call_executing` event.
5. Tool completion emits exactly one `tool_call_succeeded` event.
6. Tool result summary includes only safe fields such as `iso_time`, `timezone`, and `source`.

Suggested check:

```bash
go test ./internal/productdata ./internal/runtime ./internal/httpapi -run 'Tool|Approval|Approve|Idempotent|Worker'
```

## 5. Deny idempotency smoke

Create a second pending tool call and call deny repeatedly.

Expected results:

1. First deny records exactly one `tool_call_denied` event.
2. Repeated deny calls return denied/current state with 200 responses.
3. No `tool_call_executing` or `tool_call_succeeded` event is recorded.
4. The run finalizes through `run_stopped`.
5. UI controls disappear and ToolCallCard shows denied terminal state.

## 6. Validation failure smoke

Test these invalid model requests:

- Unknown tool name.
- Malformed JSON arguments.
- Wrong argument type.
- Unexpected secret-looking argument value.
- Multiple tool calls in one run if MVP supports only one executable call.

Expected results:

1. No execution occurs.
2. `tool_call_failed` is recorded for validation, unsupported-tool, duplicate `tool_call_id`, or MVP multi-tool safe-fail cases.
3. The run finalizes through `run_failed`.
4. Error code/message are redacted and stable.
5. Raw provider payload and raw arguments are not persisted.

## 7. Cancellation smoke

Test stop/cancel while:

1. Tool call is pending approval.
2. Tool call is approved but not yet executing.
3. Tool call is executing.

Expected results:

- Pending/approved/executing tool call reaches `tool_call_cancelled`.
- Run reaches `run_stopped` or existing M6 stopped terminal state.
- No later `tool_call_succeeded` event overwrites cancellation.
- Worker recovery does not duplicate terminal tool events.

## 8. History-first SSE replay smoke

For approve, deny, fail, and cancel paths:

1. Disconnect client after `tool_call_requested`.
2. Reconnect with history-first SSE.
3. Verify ordered replay includes every tool lifecycle event.
4. Verify live continuation produces the same UI state as uninterrupted streaming.

## 9. Browser smoke

Run the API and web app in real API mode.

Expected UI checks:

1. ToolCallCard appears for `runtime.get_current_time`.
2. Tool name and argument summary are visible and safe.
3. Approval-required state shows approve/deny controls.
4. Deny path shows denied and no execution.
5. Approve path shows approved, executing, and result.
6. Failure path shows redacted error.
7. RunRail summarizes waiting/executing/result states.
8. Timeline groups or labels tool events separately from model stream.

If browser smoke cannot run, record the exact reason in the M7 devlog during implementation.

## 10. Validation commands

Backend:

```bash
go test ./...
```

Frontend targeted tests:

```bash
bun test ./web/src/*.test.ts ./web/src/*.test.tsx ./web/src/components/*.test.ts ./web/src/components/*.test.tsx ./web/src/runtime/*.test.ts
```

Frontend build:

```bash
bun run --cwd web build
```

Docs build after docs-site updates:

```bash
bun run --cwd docs-site build
```

## 11. Documentation checks

Implementation must update:

- `docs-site/src/content/docs/architecture/tool-call-approval.md`
- `docs-site/src/content/docs/api/tool-call-approval.md`
- `docs-site/src/content/docs/runbooks/local-m7.md`
- `docs-site/src/content/docs/devlog/2026-05-24-m7-tool-call-approval.md`
- `docs-site/src/content/docs/roadmap/current-status.md`

The M7 devlog must record:

- What was implemented.
- Validation command results.
- Browser smoke results or exact reason skipped.
- Known limitations, especially full multi-step model continuation being out of MVP.
- Confirmation that no shell/file/network/MCP/browser/multi-agent/RAG tool was enabled.
