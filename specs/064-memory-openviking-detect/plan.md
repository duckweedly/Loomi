# Plan: M64 Memory OpenViking Detect

## Scope

Add one detection endpoint and one Settings wiring path for OpenViking. Reuse the existing M53 Nowledge detection pattern and the M55 provider configuration modal.

## Backend

- Extend `handleMemoryByID` with `provider/openviking/detect`.
- Implement `handleOpenVikingProviderDetect` with a 1200 ms localhost-only probe.
- Return the same `memoryProviderDetectResponse` shape used by Nowledge.

## Frontend

- Add `detectOpenVikingMemoryProvider` to `ApiClient`, `realApiClient`, `mockApiClient`, app state, and `App` wiring.
- Add an OpenViking detect button in the provider configuration modal.
- On detected result, fill the OpenViking base URL draft.

## Docs

- Update memory architecture, API, runbook, devlog, and this Spec Kit feature directory.

## Validation

```bash
go test ./internal/httpapi -run 'TestMemoryOpenVikingDetectSafeMiss|TestMemoryNowledgeDetectSafeMiss' -count=1
bun test --cwd web src/components/SettingsView.runtime.test.tsx src/App.settings.test.tsx src/realApiClient.test.ts
bun run --cwd web build
bun run build
```
