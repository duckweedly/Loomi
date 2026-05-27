---
title: M51 Manual Memory Add
description: User-authored Settings > Memory write path for local approved memories.
---

## Changed

- Added `POST /v1/memory/entries` for one user-authored memory entry.
- Added real/mock API client support and workspace state refresh after create.
- Added a compact manual-add form to Settings > Memory.

## Safety

- Manual create returns safe entry projections only.
- Raw content and content hash do not appear in the response.
- Agent and post-run writes remain approval-gated through write proposals.
- Bulk clear-all remains deferred because it is destructive.

## Validation

- `go test ./internal/httpapi -run 'TestMemoryHandlersCreateManualEntry|TestMemorySnapshotAndImpressionHandlers' -count=1`
- `bun test --cwd web src/components/SettingsView.runtime.test.tsx src/mockApiClient.test.ts src/memory.test.ts`
