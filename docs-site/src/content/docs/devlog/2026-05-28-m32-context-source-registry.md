---
title: M32 Context Source Registry Foundation
description: Safe thread-scoped source registration without connector execution.
---

M32 fills the next Arkloop/Craft parity gap at the context boundary: Loomi now has a durable place to register source references before adding heavier connector execution or activity ingestion.

Implemented:

- Added `context_sources` storage for thread-scoped source references.
- Added `POST /v1/threads/{thread_id}/sources` and `GET /v1/threads/{thread_id}/sources`.
- Supported `url`, `github_repo`, `workspace_path`, and `note` source kinds.
- Normalized public URLs by stripping query and fragment.
- Rejected localhost/private URL hosts, URL credentials, absolute/traversal workspace paths, `.env*`, `.git`, private key, `secrets`, and `credentials` paths.
- Kept metadata and summary redacted before persistence.

Boundaries:

- No connector marketplace, OAuth, GitHub API integration, crawler, browser, web search, MCP sync, activity recorder, or run-time source fetching.
- No global source registry yet; this slice is thread-scoped.
- No UI changes in this work session.

Validation:

```bash
go test ./internal/productdata ./internal/httpapi -run 'Test.*ContextSource|Test.*Source' -count=1
```
