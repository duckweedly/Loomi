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

`browser.snapshot` and `browser.screenshot` accept `session_id`. `browser.click_link` accepts `session_id`, `link_index`, optional `max_bytes`, and optional `timeout_ms`. `browser.type` accepts `session_id`, `target`, and `text`. `browser.press` accepts `session_id` and `key`.

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

`browser.screenshot`:

```json
{
  "session_id": "br_..."
}
```

`browser.type`:

```json
{
  "session_id": "br_...",
  "target": "q",
  "text": "Loomi"
}
```

`browser.press`:

```json
{
  "session_id": "br_...",
  "key": "Enter"
}
```

The runtime rejects relative URLs, non-HTTP(S) schemes, username/password credentials, localhost, loopback, link-local, private, multicast, unspecified hosts, failed resolution, blocked redirects, unknown sessions, blocked link targets, unknown input targets, and keys outside the bounded allowlist.

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
  "inputs": [
    {
      "index": 0,
      "target": "q",
      "label": "Search"
    }
  ],
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

`browser.screenshot` returns `format = text` and `screenshot_text`; it does not return image bytes. `browser.type` returns `target`, `text_length`, and `form_value_count`, not typed text. `browser.press` returns the key and whether the key was `Enter`.

Events and continuation context persist safe summaries only. They do not include raw HTML, cookies, Set-Cookie values, Authorization values, typed text, provider raw payloads, local paths, or secret-looking content.
