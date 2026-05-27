---
title: M47 Memory Provider Config Details
description: Safe Nowledge and OpenViking memory provider configuration shape.
---

M47 moves the memory provider settings from a generic semantic placeholder toward the target provider mechanism.

## Completed

- Added migration `000015` for OpenViking and Nowledge provider configuration fields.
- Added backend config/status structs for OpenViking root, embedding, VLM, rerank, and Nowledge base URL/API key/timeout.
- Kept provider secrets write-only: HTTP responses expose key presence booleans and never return raw keys.
- Updated `/v1/memory/provider` request/response mapping.
- Updated Settings > Memory with Local / Nowledge / OpenViking selection and provider-specific config fields.

## Validation

- `go test ./internal/productdata ./internal/httpapi -run 'TestMemoryProvider|TestPrepareRunContextIncludesMemoryProviderReadiness' -count=1`
- `bun test --cwd web src/memory.test.ts src/components/SettingsView.runtime.test.tsx`
- `bun run --cwd web build`
- `bun run --cwd docs-site build`
- Browser smoke against Settings > Memory with Nowledge and OpenViking fields.

## Still Deferred

- Real OpenViking and Nowledge adapter execution.
- Snapshot and impression rebuild APIs.
- Notebook dual-provider tools.
- Activity recorder ingestion.
