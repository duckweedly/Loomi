---
title: M70 Memory Provider Error UI
description: Runtime provider error details in Settings > Memory recent errors.
---

## Completed

- Added a Settings > Memory recent error formatter.
- Displayed optional runtime `eventType` and `runId` values from `/v1/memory/errors`.
- Added wrapping style for long runtime ids.
- Preserved the existing provider diagnostic panel and copy.

## Validation

- `bun test --cwd web src/components/SettingsView.runtime.test.tsx src/memory.test.ts`
- `bun run --cwd web build`
- `bun run build` from `docs-site/`
- `git diff --check`

## Notes

- Runtime ids are displayed as opaque diagnostic text only.
- The UI still does not expose prompt text, raw memory content, provider traces, or secrets.
