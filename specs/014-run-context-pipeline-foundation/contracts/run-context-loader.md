# Contract: RunContext Loader

## Purpose

Prepare the critical execution context for a worker-owned run from durable product data, not from API request memory.

## Loader Input

```json
{
  "run_id": "run_...",
  "thread_id": "thread_...",
  "job_id": "job_...",
  "worker_id": "worker_local_1",
  "attempt": 1,
  "ownership_version": 3
}
```

Rules:

- Input ids come from the claimed background job and current run.
- The loader must verify the run, thread, and job still match.
- Ownership checks remain with existing worker/job guards; the loader must not bypass them.

## Loader Output

```json
{
  "run": {
    "id": "run_...",
    "thread_id": "thread_...",
    "status": "running",
    "source": "model_gateway"
  },
  "thread": {
    "id": "thread_..."
  },
  "messages": [
    { "id": "msg_1", "role": "user", "content_ref": "durable-message-content" }
  ],
  "job": {
    "id": "job_...",
    "attempt": 1,
    "metadata_summary": {
      "message_id": "msg_1",
      "provider_id": "custom",
      "model": "selected-model"
    }
  },
  "provider_route": {
    "provider_id": "custom",
    "model": "selected-model",
    "available": true
  },
  "enabled_tools": [
    {
      "name": "runtime.get_current_time",
      "approval_policy": "always_required"
    }
  ]
}
```

Output rules:

- Message content may be used in memory for provider context, but persisted stage summaries must only include counts and safe ids.
- Provider route must not expose API keys, base authorization headers, raw request bodies, or raw provider errors.
- Enabled tools are allowlisted by runtime definitions and current run/tool state.

## Failure Semantics

The loader fails before runtime invocation when:

- run, thread, or job cannot be found
- run/thread/job ownership boundaries do not match
- required message history is unavailable
- model-gateway run has no usable provider/model route
- job metadata is malformed or references unsupported behavior

Failure output:

```json
{
  "stage": "prepare_context",
  "error_code": "run_context_missing_required_data",
  "message": "Run context could not be prepared from durable state."
}
```

Rules:

- Failure must be redacted before persistence.
- Failure must produce a terminal run failure through existing finalization/worker guards.
- No provider/runtime invocation may happen after loader failure.

## Safe Summary

`prepare_context` completed metadata may include:

- `message_count`
- `has_job_metadata`
- `provider_id`
- `model`
- `enabled_tool_count`
- `has_continuation_projection`

It must not include:

- message text
- provider credentials
- raw provider payloads
- raw tool result payloads
- arbitrary local file, shell, browser, or desktop state
