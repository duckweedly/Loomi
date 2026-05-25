---
title: M16 Work Mode Payload
description: Work mode foundation 复用现有 API 的 safe metadata 约定。
---

M16 没有新增后端 endpoint。前端使用现有 thread/message/run/event API，并在 run event metadata 中识别可选 Work mode 字段。

M17 仍不新增 HTTP endpoint。可重复 evidence 通过 `cmd/loomi-seed` 的 local-dev/test scenario 写入现有 productdata service/repository：`LOOMI_SEED_SCENARIO=m17-work-artifact go run ./cmd/loomi-seed`。这不是生产事件写接口。

## Existing Inputs

- `Thread.mode`: `work` 启用 Work Plan View；`chat` 不启用。
- `Message.content`: fallback goal/step 来源。
- `Run.status`: 当前进度状态。
- `Run.events[]`: recent progress 和 metadata 投影来源。
- `RunEvent.metadata`: 可选 safe metadata。

## Optional Metadata

```json
{
  "work_goal": "Ship M16 work mode foundation",
  "work_steps": [
    {
      "id": "step-render",
      "title": "Render Work Plan View",
      "status": "running",
      "summary": "Projected from existing run events"
    }
  ],
  "work_artifacts": [
    {
      "id": "artifact-plan",
      "title": "M16 Work Plan",
      "type": "markdown",
      "source_thread_id": "thread-brief",
      "source_run_id": "run-1",
      "summary": "Safe metadata preview only",
      "created_at": "2026-05-25 10:25",
      "updated_at": "Now",
      "redaction_applied": true
    }
  ]
}
```

## Redaction Contract

- Secret-looking strings are displayed as `[redacted]`.
- Executable metadata keys such as command, path, file, shell, browser, filesystem, execute, and URL are omitted from artifact cards and never rendered as actions.
- Unknown fields are ignored for artifact cards.
- No file, browser, or shell execution is triggered by this payload.
- `redaction_applied` may be displayed as a marker such as "Redacted unsafe metadata" without revealing removed fields.

## M17 Evidence Seed Output

The local seed logs these identifiers for browser smoke:

- `thread_id`
- `message_id`
- `run_id`
- `event_id`

Expected seeded thread id: `thr_m17_work_artifact`.
