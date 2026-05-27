# Contract: M15 Chat Real Integrated Smoke

## Command Contract

The gated smoke command must be explicit so normal `go test ./...` remains deterministic and cheap:

```bash
LOOMI_M15_REAL_CHAT_SMOKE=1 go test ./internal/httpapi -run TestM15ChatRealIntegratedSmoke -count=1 -v
```

Expected result:

- The command exits successfully.
- The test logs the run id, tool-call id, final assistant message, and event type sequence.
- Without the gate, the test is skipped and explains the required environment variable.

## Evidence Contract

The smoke must verify these observable milestones from real API/service/worker execution:

- chat run queued through HTTP/API path
- worker/run pipeline context prepared
- `memory_snapshot_loaded`
- MCP discovery and stable candidate hash
- `tool_call_approval_required`
- HTTP approval accepted
- `tool_call_approved`
- `tool_call_executing`
- `tool_call_succeeded`
- provider continuation started/completed
- final assistant message persisted
- run completed

## Redaction Contract

The smoke must assert sensitive canaries are absent from:

- run creation or fetch API payloads
- RunContext safe summary/debug payloads
- persisted run events
- tool result summary
- final assistant message
- docs examples added for M15
