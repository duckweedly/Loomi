---
title: M53 Nowledge Local Detect
description: Localhost Nowledge detection from Settings > Memory.
---

## Changed

- Added `GET /v1/memory/provider/nowledge/detect`.
- Added real/mock client support and Settings > Memory detect action.

## Safety

- Detector only probes `127.0.0.1:14242/health` with a short timeout.
- It returns detected state, base URL, and safe message only.
- It does not discover API keys or scan remote hosts.

## Validation

- `go test ./internal/httpapi -run 'TestMemoryNowledgeDetectSafeMiss|TestMemoryErrorsReportsProviderDiagnostic' -count=1`
- `bun test --cwd web src/components/SettingsView.runtime.test.tsx src/mockApiClient.test.ts src/memory.test.ts`
- `bun run --cwd web build`
