---
title: Artifact Runtime Tool API
description: Catalog, arguments, and result contract for M28 artifact tools.
---

## Catalog Entries

`GET /v1/tools/catalog` includes:

```json
{
  "name": "artifact.create_text",
  "source": "builtin",
  "group": "artifact",
  "risk_level": "medium",
  "approval_policy": "always_required",
  "execution_state": "executable",
  "safe_metadata": {
    "scope": "artifact",
    "read_only": false,
    "non_executable": true,
    "arguments": ["title", "filename", "mime_type", "display", "content", "max_bytes"]
  }
}
```

`artifact.create_visual` uses the same approval policy with `safe_metadata.renderable = true` and supports `image/svg+xml` or `text/html` previews. `artifact.read` and `artifact.list` use `read_only = true`.

## Arguments

`artifact.create_text`:

```json
{
  "title": "Implementation Notes",
  "filename": "implementation-notes.md",
  "mime_type": "text/markdown",
  "display": "panel",
  "content": "Bounded UTF-8 text",
  "max_bytes": 32768
}
```

`artifact.create_visual`:

```json
{
  "title": "System Diagram",
  "filename": "system-diagram.svg",
  "mime_type": "image/svg+xml",
  "display": "inline",
  "content": "<svg viewBox=\"0 0 600 360\">...</svg>",
  "max_bytes": 32768
}
```

`artifact.read`:

```json
{
  "artifact_id": "art_...",
  "max_bytes": 4096
}
```

`artifact.list`:

```json
{
  "limit": 20
}
```

## Result Summary

```json
{
  "tool": "artifact.create_text",
  "scope": "artifact",
  "operation": "create_text",
  "artifact_id": "art_...",
  "title": "Implementation Notes",
  "artifact_type": "text",
  "content_bytes": 1200,
  "text_excerpt": "Bounded UTF-8 text",
  "truncated": false,
  "redaction_applied": false,
  "artifacts": [
    {
      "key": "art_...",
      "artifact_id": "art_...",
      "title": "Implementation Notes",
      "filename": "implementation-notes.md",
      "mime_type": "text/markdown",
      "display": "panel",
      "content_bytes": 1200,
      "text_excerpt": "Bounded UTF-8 text"
    }
  ]
}
```

Visual artifact result summaries use `operation = "create_visual"` and `artifact_type = "visual"`. Their first `artifacts[]` reference includes bounded render content for the sandboxed Preview drawer:

```json
{
  "tool": "artifact.create_visual",
  "scope": "artifact",
  "operation": "create_visual",
  "artifact_id": "art_...",
  "title": "System Diagram",
  "artifact_type": "visual",
  "artifacts": [
    {
      "key": "art_...",
      "artifact_id": "art_...",
      "title": "System Diagram",
      "filename": "system-diagram.svg",
      "mime_type": "image/svg+xml",
      "display": "inline",
      "content": "<svg viewBox=\"0 0 600 360\">...</svg>"
    }
  ]
}
```

Events and continuation context persist safe summaries only. They do not include raw unbounded content, executable controls, credentials, local paths, or raw provider payloads. Assistant messages may reference a returned artifact with `[title](artifact:<key>)`; the key must come from the tool result and must not be invented.

## Read-only HTTP Projection

The CLI and UI may inspect persisted artifacts through thread-scoped read-only endpoints:

- `GET /v1/threads/:thread_id/artifacts?limit=20`
- `GET /v1/threads/:thread_id/artifacts/:artifact_id?max_bytes=4096`

Responses return safe artifact fields: id, thread/run ids, title, type, content byte count, bounded `text_excerpt`, truncation flag, and timestamps. Read responses may include bounded `content` for the requested artifact; list responses do not include a raw `content` field. Cross-thread reads return `artifact_not_found`.

There is intentionally no HTTP create/update/delete endpoint in this slice. Artifact creation remains an approval-gated Work-mode tool call.
