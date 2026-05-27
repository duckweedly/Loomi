# Data Model: M15 Chat Real Integrated Smoke Closeout

## M15 Smoke Scenario

- `workspace_id`, `thread_id`, `run_id`: existing identifiers created through the real API/service path.
- `provider_fixture`: deterministic provider behavior with two phases: tool request, continuation final message.
- `sensitive_canaries`: values that must never appear in shareable surfaces.
- Relationships: owns one approved memory fixture, one discovered MCP candidate, one approval projection, one worker execution, one continuation, and one replay evidence set.

## Approved Memory Snapshot

- `entry_id`: approved memory entry id.
- `safe_summary`: redacted text attached to RunContext and visible in safe metadata.
- `raw_canary`: sensitive source value used only to prove redaction.
- Validation: only approved memory may enter RunContext; raw canary must not appear in API/event/docs surfaces.

## MCP Candidate Fixture

- `server_slug`: local deterministic stdio server slug.
- `tool_name`: safe fixture tool name.
- `namespaced_name`: `mcp.<server_slug>.<tool_name>`.
- `candidate_hash`: stable safe hash emitted in discovery/candidate evidence.
- `persona_allowed`: true for this smoke's selected persona snapshot.
- Validation: provider request must match the discovered and persona-allowed namespaced candidate.

## Tool Approval Projection

- `tool_call_id`: provider fixture tool-call id.
- `approval_status`: `required` -> `approved`.
- `execution_status`: `blocked` -> `executing` -> `succeeded`.
- `arguments_summary`: redacted arguments only.
- `result_summary`: redacted result only.

## Replay Evidence Set

- `events`: persisted event history in sequence order.
- Required milestones: queued, worker claimed, pipeline context, `memory_snapshot_loaded`, MCP discovery/candidate hash, approval required, approved, executing, succeeded, continuation, run completed.
- Validation: no sensitive canary appears in any serialized evidence surface.
