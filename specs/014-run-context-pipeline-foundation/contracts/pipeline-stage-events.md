# Contract: Pipeline Stage Events

## Stage Names

Normal MVP order:

```text
prepare_context
resolve_tools
invoke_runtime
finalize
```

Rules:

- The first M9 slice is linear.
- No arbitrary workflow graph, plugin runtime, or user-authored middleware is introduced.
- Future stages can be inserted through the stage registration/composition point when separately specified.

## Event Types

| Event Type | Category | Required Metadata | Meaning |
| --- | --- | --- | --- |
| `pipeline_stage_started` | `progress` | `stage`, `job_id`, `attempt` | Worker entered a pipeline stage |
| `pipeline_stage_completed` | `progress` | `stage`, `job_id`, `attempt`, safe summary | Worker completed a pipeline stage |
| `pipeline_stage_failed` | `error` | `stage`, `job_id`, `attempt`, `error_code` | Stage failed safely |

Existing M6 event names `pipeline_step_started` and `pipeline_step_completed` may be retained as backend compatibility names if implementation chooses. The user-facing dotted runtime names should still group them as pipeline stage trace.

## Success Ordering

```text
job_claimed
pipeline_stage_started        stage=prepare_context
pipeline_stage_completed      stage=prepare_context
pipeline_stage_started        stage=resolve_tools
pipeline_stage_completed      stage=resolve_tools
pipeline_stage_started        stage=invoke_runtime
model/runtime events...
pipeline_stage_completed      stage=invoke_runtime
pipeline_stage_started        stage=finalize
assistant/final run events...
pipeline_stage_completed      stage=finalize
run_completed
```

If current terminal event ordering requires `run_completed` inside finalization, the trace must still make it clear that `finalize` started before terminal completion and completed at or before the terminal boundary visible to replay.

## Failure Ordering

Prepare-context failure:

```text
pipeline_stage_started        stage=prepare_context
pipeline_stage_failed         stage=prepare_context
run_failed
```

Runtime failure:

```text
pipeline_stage_started        stage=invoke_runtime
model/provider/tool failure events...
pipeline_stage_failed         stage=invoke_runtime
run_failed
```

Rules:

- A failed stage must not be followed by normal later stages unless the existing worker retry/cancellation semantics explicitly restart the attempt.
- Terminal writes remain ownership-guarded.
- Stop requests may cause `run_stopped` rather than `run_failed`.

## Safe Metadata

Allowed:

- stage name
- run/thread/job ids
- attempt and ownership version
- message count
- enabled tool names/counts
- provider/model safe labels
- model phase label
- redacted error code/message

Forbidden:

- credentials and authorization headers
- raw provider request/response bodies
- raw tool result payloads
- message text in stage metadata
- shell output, file contents, desktop/browser captured state
- hidden prompt/system context beyond safe labels
