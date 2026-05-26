---
title: Browser Automation Tool API
description: Catalog, arguments, and result contract for M27 browser tools.
---

## Catalog Entries

`GET /v1/tools/catalog` includes:

```json
{
  "name": "browser.open",
  "source": "builtin",
  "group": "browser",
  "risk_level": "medium",
  "approval_policy": "always_required",
  "execution_state": "executable",
  "safe_metadata": {
    "scope": "browser",
    "read_only": true,
    "network_access": "public_http_only",
    "stateful_session": true,
    "arguments": ["url", "max_bytes", "timeout_ms"]
  }
}
```

`browser.snapshot` accepts `session_id`. `browser.click_link` accepts `session_id`, `link_index`, optional `max_bytes`, and optional `timeout_ms`.

Chat mode RunContext resolution omits all browser tools. Work mode can include them when persona allowlist permits them.

## Arguments

`browser.open`:

```json
{
  "url": "https://example.com/docs",
  "max_bytes": 32768,
  "timeout_ms": 5000
}
```

`browser.snapshot`:

```json
{
  "session_id": "br_..."
}
```

`browser.click_link`:

```json
{
  "session_id": "br_...",
  "link_index": 0,
  "max_bytes": 32768,
  "timeout_ms": 5000
}
```

The runtime rejects relative URLs, non-HTTP(S) schemes, username/password credentials, localhost, loopback, link-local, private, multicast, unspecified hosts, failed resolution, blocked redirects, unknown sessions, and blocked link targets.

## Result Summary

```json
{
  "tool": "browser.open",
  "scope": "browser",
  "operation": "open",
  "session_id": "br_...",
  "url": "https://example.com/docs",
  "final_url": "https://example.com/docs",
  "status_code": 200,
  "content_type": "text/html",
  "title": "Example Docs",
  "text_excerpt": "Bounded visible text...",
  "links": [
    {
      "index": 0,
      "text": "Next",
      "url": "https://example.com/docs/next",
      "blocked": false
    }
  ],
  "link_count": 1,
  "bytes_read": 12000,
  "byte_limit": 32768,
  "truncated": false,
  "redaction_applied": false
}
```

Events and continuation context persist safe summaries only. They do not include raw HTML, cookies, Set-Cookie values, Authorization values, provider raw payloads, local paths, or secret-looking content.
