---
title: M16 Work Mode Payload
description: Work mode foundation 复用现有 API 的 safe metadata 约定。
---

M16 没有新增后端 endpoint。前端使用现有 thread/message/run/event API，并在 run event metadata 中识别可选 Work mode 字段。

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
      "updated_at": "Now"
    }
  ]
}
```

## Redaction Contract

- Secret-looking strings are displayed as `[redacted]`.
- Executable metadata keys such as command, path, file, shell, browser, filesystem, execute, and URL are not rendered as artifact actions.
- Unknown fields are ignored for artifact cards.
- No file, browser, or shell execution is triggered by this payload.
