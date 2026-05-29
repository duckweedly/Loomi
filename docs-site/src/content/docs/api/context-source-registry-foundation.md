---
title: Context Source Registry API
description: Thread-scoped safe source registration for later connector and context routing work.
---

M32 adds a narrow thread-scoped source registry. It records what a run or user may treat as context, but it does not fetch, crawl, sync, authenticate, or execute external connectors.

## Create Source

```http
POST /v1/threads/{thread_id}/sources
```

```json
{
  "kind": "url",
  "title": "Docs",
  "locator": "https://example.com/docs?token=secret",
  "summary": "Public product docs",
  "metadata": {}
}
```

Response:

```json
{
  "source": {
    "id": "src_...",
    "thread_id": "thr_...",
    "kind": "url",
    "title": "Docs",
    "locator": "https://example.com/docs",
    "summary": "Public product docs",
    "status": "registered",
    "metadata": {},
    "created_at": "2026-05-28T00:00:00Z",
    "updated_at": "2026-05-28T00:00:00Z"
  }
}
```

Supported `kind` values:

| Kind | Locator contract |
| --- | --- |
| `url` | Public HTTP(S) URL. Query and fragment are stripped before persistence. |
| `github_repo` | Public `github.com/{owner}/{repo}` URL. Extra path, query, fragment, and `.git` suffix are normalized away. |
| `workspace_path` | Relative workspace path only. Absolute, traversal, `.git`, `.env*`, private key, `secrets`, and `credentials` paths are rejected. |
| `note` | Bounded redacted text locator for manual context labels. |

URL sources reject credentials, localhost, loopback, private, link-local, multicast, and unspecified hosts.

## List Sources

```http
GET /v1/threads/{thread_id}/sources?limit=20
```

Sources are scoped to the current local identity and thread. Cross-thread reads return an empty list or `thread_not_found`; no global source registry is exposed in this slice.

## Boundaries

- No OAuth, GitHub API calls, crawling, sync workers, web search, browser automation, MCP connection, or marketplace behavior.
- No raw token, query string, credential, host absolute path, or sensitive workspace path is persisted.
- The registry is a durable reference layer for later connector orchestration and RunContext source selection.
