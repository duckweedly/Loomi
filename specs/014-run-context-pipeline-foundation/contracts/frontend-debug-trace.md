# Contract: Frontend Timeline and Debug Trace

## Mapping

Backend stage events map into runtime events as pipeline/debug trace rows.

| Backend Event | Frontend Type | Group | Status |
| --- | --- | --- | --- |
| `pipeline_stage_started` | `pipeline.stage.started` | `worker-job` or `pipeline` | `running` |
| `pipeline_stage_completed` | `pipeline.stage.completed` | `worker-job` or `pipeline` | `running` |
| `pipeline_stage_failed` | `pipeline.stage.failed` | `worker-job` or `pipeline` | `failed` |

If implementation retains M6 `pipeline_step_*` backend names, the frontend must still display the M9 stage names from metadata.

## Timeline Behavior

Timeline/debug views must show:

1. Context prepared
2. Tools resolved
3. Runtime invoked
4. Finalized

Rules:

- Live SSE and history replay must produce the same ordered stage trace.
- Stage details must be safe summaries.
- Tool-call events remain grouped with tool-call UI; stage trace should not duplicate raw tool results.
- Model deltas/final messages remain normal message/runtime events.
- Terminal failure rows should point to the failed stage when available.

## Debug Summary Labels

Suggested labels:

- `prepare_context`: "Context prepared"
- `resolve_tools`: "Tools resolved"
- `invoke_runtime`: "Runtime invoked"
- `finalize`: "Finalized"

These labels are UI text only; event metadata remains stable stage names.

## Replay Requirements

- Reconnecting after a run finishes rebuilds the same stage trace as a live run.
- Dedupe continues to use event id/sequence.
- A failed stage marks the run as failed/stopped according to the terminal event.
- A missing stage event must not crash the Timeline; it should simply omit that row.
