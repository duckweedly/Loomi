---
title: Web Tools API
description: Catalog, arguments, and result contracts for web.fetch and web.search.
---

## Catalog Entries

`GET /v1/tools/catalog` includes:

```json
[
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
  },
  {
    "name": "web.search",
    "source": "builtin",
    "group": "web",
    "risk_level": "medium",
    "approval_policy": "read_only",
    "execution_state": "executable",
    "safe_metadata": {
      "scope": "web",
      "read_only": true,
      "network_access": "search_provider_api",
      "providers": ["tavily", "brave"],
      "arguments": ["query", "provider", "limit", "timeout_ms"]
    }
  }
]
```

Chat mode RunContext can include `web.search` when persona allowlist permits it. `web.fetch` remains Work-mode only. Work mode can include both web tools.

## web.fetch Arguments

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

## web.search Arguments

```json
{
  "query": "latest AI news",
  "provider": "tavily",
  "limit": 5,
  "timeout_ms": 5000
}
```

`query` is required. `provider` is optional and may be `tavily` or `brave`; runtime defaults to Tavily when `LOOMI_TAVILY_API_KEY` is configured, otherwise Brave when `LOOMI_BRAVE_SEARCH_API_KEY` is configured. `limit` and `timeout_ms` are optional and clamped by runtime bounds.

## Web Search Config API

`GET /v1/web-search/config` returns only booleans:

```json
{
  "config": {
    "has_tavily_key": true,
    "has_brave_key": false,
    "enabled": true
  }
}
```

`POST /v1/web-search/config` accepts `tavily_api_key` and/or `brave_api_key`. At least one key enables web search. Keys are stored through the product data store so a local API restart can reload them. Responses never echo key values.

## web.search Result Summary

```json
{
  "tool": "web.search",
  "scope": "web",
  "operation": "search",
  "provider": "tavily",
  "query": "latest AI news",
  "result_count": 1,
  "items": [
    {
      "title": "Example News",
      "url": "https://example.com/news",
      "snippet": "Bounded public result snippet..."
    }
  ],
  "redaction_applied": false
}
```

The runtime never stores provider API keys, raw search response bodies, headers, cookies, or raw page content in tool result summaries.
