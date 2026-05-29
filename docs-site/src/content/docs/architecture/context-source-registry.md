---
title: Context Source Registry
description: Durable source references for future connector and RunContext orchestration.
---

The M32 source registry is intentionally a reference layer, not a connector runtime.

```mermaid
flowchart LR
  User["User or future agent tool"] --> API["/v1/threads/{id}/sources"]
  API --> Validate["Normalize and safety validate"]
  Validate --> Store["context_sources"]
  Store --> List["Thread-scoped source projection"]
  Store -. later .-> RunContext["RunContext source selection"]
  Store -. later .-> Connector["Connector execution"]
```

## Model

`context_sources` belongs to one user and one thread. Each row stores:

- stable `src_...` id
- `kind`
- safe `title`
- normalized `locator`
- optional redacted `summary`
- redacted `metadata`
- `registered` status

## Safety

URL locators are public HTTP(S) only. Query strings and fragments are discarded before storage. Localhost, loopback, private network, link-local, multicast, unspecified hosts, and URL credentials are rejected.

Workspace locators are relative paths only. Traversal, absolute paths, `.git`, `.env*`, private keys, `secrets`, and `credentials` are rejected before persistence.

## Future Use

This registry gives later connector work a safe durable anchor for source selection. Fetching, crawling, sync scheduling, source-specific auth, and source-to-RunContext replay should be added as separate slices with their own approval and observability boundaries.
