# Contract: M27 Browser Automation Foundation

## Tool Catalog

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
    "network_access": "public_http_only",
    "stateful": true
  }
}
```

The same catalog shape applies to `browser.snapshot` and `browser.click_link`.

Chat mode catalog resolution must omit browser tools. Work mode may include them when persona allowlist permits them.

## Tool Arguments

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
  "session_id": "br_run_1_tc_1"
}
```

`browser.click_link`:

```json
{
  "session_id": "br_run_1_tc_1",
  "link_index": 0,
  "max_bytes": 32768,
  "timeout_ms": 5000
}
```

## Tool Result

```json
{
  "tool": "browser.open",
  "scope": "browser",
  "operation": "open",
  "session_id": "br_run_1_tc_1",
  "url": "https://example.com/docs",
  "final_url": "https://example.com/docs",
  "status_code": 200,
  "content_type": "text/html",
  "title": "Example Docs",
  "text_excerpt": "Bounded page text...",
  "links": [
    { "index": 0, "text": "Next", "href": "https://example.com/next", "host": "example.com", "blocked": false }
  ],
  "bytes_read": 12000,
  "byte_limit": 32768,
  "truncated": false,
  "redaction_applied": false
}
```

## Rejections

The runtime rejects before network execution when:

- run context is Chat mode or tool is not enabled
- URL or clicked target is malformed, relative, credentialed, or non-HTTP(S)
- host is local/private/link-local/multicast/unspecified
- DNS resolution fails or resolves to blocked IPs
- redirect target fails URL validation
- session id is unknown, expired, or belongs to a different run
- link index is missing or out of range
- run is denied, stopped, terminal, duplicate, or out-of-scope
