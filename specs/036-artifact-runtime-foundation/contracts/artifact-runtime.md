# Contract: M28 Artifact Runtime Foundation

## Catalog

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

## Create Arguments

```json
{
  "title": "Implementation Notes",
  "content": "Bounded UTF-8 text",
  "max_bytes": 32768
}
```

## Read Arguments

```json
{
  "artifact_id": "art_...",
  "max_bytes": 4096
}
```

## List Arguments

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

## Rejections

- Chat mode RunContext filtering
- missing title/content/artifact_id
- unsupported type
- oversized content
- invalid UTF-8
- unknown artifact
- out-of-scope artifact
- denied/stopped/terminal calls
