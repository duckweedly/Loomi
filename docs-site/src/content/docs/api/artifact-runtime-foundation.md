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
    "arguments": ["title", "content", "max_bytes"]
  }
}
```

`artifact.read` and `artifact.list` use `read_only = true`.

## Arguments

`artifact.create_text`:

```json
{
  "title": "Implementation Notes",
  "content": "Bounded UTF-8 text",
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
  "redaction_applied": false
}
```

Events and continuation context persist safe summaries only. They do not include raw unbounded content, executable controls, credentials, local paths, or raw provider payloads.
