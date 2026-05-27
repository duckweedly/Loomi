# Data Model: M29 Multi-agent Runtime Foundation

## AgentTask

- `id`: stable local child task id
- `thread_id`: owning thread
- `run_id`: source run
- `role`: bounded role label, such as `researcher`, `implementer`, or `reviewer`
- `goal`: bounded task goal
- `status`: `spawned` or `completed`
- `result_summary`: optional bounded completion summary
- `created_at`: service timestamp
- `updated_at`: service timestamp

## Tool Arguments

`agent.spawn`:

- `role`: required safe role label
- `goal`: required bounded goal

`agent.list`:

- `limit`: optional bounded result count

`agent.complete`:

- `task_id`: required existing child task id
- `result_summary`: required bounded completion summary

## Result Summary

- `tool`
- `scope = agent`
- `operation`
- `task_id`
- `role`
- `goal`
- `status`
- `result_summary`
- `redaction_applied`
