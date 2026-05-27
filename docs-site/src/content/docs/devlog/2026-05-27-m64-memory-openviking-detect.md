---
title: M64 Memory OpenViking Detect
description: Safe localhost OpenViking detection for Settings > Memory.
---

## Completed

- Added `GET /v1/memory/provider/openviking/detect`.
- Added real/mock client support and Settings > Memory wiring for the OpenViking detect action.
- Made Nowledge/OpenViking detect failures visible in the provider modal instead of silently dropping failed requests.
- Kept the detector localhost-only, short-timeout, and status-only.

## Safety

- The detector does not read stored OpenViking keys.
- The detector does not send API keys to the probed service.
- The response returns only detected state, default base URL, safe message, and request id.

## Validation

- `go test ./internal/httpapi -run 'TestMemoryOpenVikingDetectSafeMiss|TestMemoryNowledgeDetectSafeMiss' -count=1`
- `bun test --cwd web src/components/SettingsView.runtime.test.tsx src/App.settings.test.tsx src/realApiClient.test.ts`
- `bun run --cwd web build`
- `bun run build` from `docs-site/`
