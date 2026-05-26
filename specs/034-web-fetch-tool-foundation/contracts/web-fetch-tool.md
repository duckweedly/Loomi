# Contract: M26 Web Fetch Tool Foundation

## Tool Catalog

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
    "network_access": "public_http_only"
  }
}
```

Chat mode catalog resolution must omit `web.fetch`. Work mode may include it when persona allowlist permits it.

## Tool Arguments

Provider tool-call arguments:

```json
{
  "url": "https://example.com/docs",
  "max_bytes": 20000,
  "timeout_ms": 5000
}
```

`max_bytes` and `timeout_ms` are optional and clamped to runtime bounds.

## Tool Result

Safe result envelope:

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
  "byte_limit": 20000,
  "truncated": false,
  "redaction_applied": false
}
```

## Rejections

The runtime rejects before network execution when:

- run context is Chat mode or tool is not enabled
- scheme is not HTTP(S)
- URL is relative, malformed, or credentialed
- host is local/private/link-local/multicast/unspecified
- DNS resolution produces only blocked IPs
- redirect target fails the same URL validation
- run is denied, stopped, terminal, duplicate, or out-of-scope
