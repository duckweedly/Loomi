---
title: Web Fetch Tool API
description: Catalog, arguments, and result contract for web.fetch.
---

## Catalog Entry

`GET /v1/tools/catalog` includes:

```json
{
  "name": "web.fetch",
  "source": "builtin",
  "group": "web",
  "risk_level": "medium",
  "approval_policy": "always_required",
  "execution_state": "executable",
  "safe_metadata": {
    "scope": "web",
    "read_only": true,
    "network_access": "public_http_only",
    "arguments": ["url", "max_bytes", "timeout_ms"]
  }
}
```

Chat mode RunContext resolution omits the tool. Work mode can include it when persona allowlist permits it.

## Arguments

```json
{
  "url": "https://example.com/docs",
  "max_bytes": 32768,
  "timeout_ms": 5000
}
```

`url` is required. `max_bytes` and `timeout_ms` are optional and clamped by runtime bounds.

The runtime rejects relative URLs, non-HTTP(S) schemes, username/password credentials, localhost, loopback, link-local, private, multicast, unspecified hosts, failed resolution, and blocked redirects.

## Result Summary

```json
{
  "tool": "web.fetch",
  "scope": "web",
  "operation": "fetch",
  "url": "https://example.com/docs",
  "final_url": "https://example.com/docs",
  "status_code": 200,
  "content_type": "text/html",
  "title": "Example Docs",
  "text_excerpt": "Bounded text excerpt...",
  "bytes_read": 12000,
  "byte_limit": 32768,
  "truncated": false,
  "redaction_applied": false
}
```

Unsupported non-text content returns metadata with `unsupported_content = true` and no body excerpt.
