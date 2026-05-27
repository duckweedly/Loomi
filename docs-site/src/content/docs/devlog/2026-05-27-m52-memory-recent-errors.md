---
title: M52 Memory Recent Errors
description: Safe provider diagnostic errors in Settings > Memory.
---

## Changed

- Added `GET /v1/memory/errors`.
- Added frontend error event mapping, state loading, and provider panel rendering.

## Safety

- Errors are derived from memory provider diagnostic state.
- API keys, Authorization headers, upstream traces, request bodies, local paths, and secret-like values are not returned.

## Validation

- `go test ./internal/httpapi -run 'TestMemoryErrorsReportsProviderDiagnostic|TestMemoryHandlersCreateManualEntry' -count=1`
- `bun test --cwd web src/memory.test.ts src/components/SettingsView.runtime.test.tsx src/mockApiClient.test.ts`
