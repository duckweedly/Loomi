---
title: M55 Memory Provider Config Modal
description: Compact Memory provider configuration flow.
---

## Completed

- Moved Nowledge and OpenViking provider detail fields into a modal opened from Settings > Memory.
- Kept the main Memory provider surface focused on enablement, post-run organization, provider choice, status, diagnostics, and Configure.
- Kept Nowledge local detection inside the modal and preserved the existing safe update path.

## Validation

- `bun test --cwd web src/components/SettingsView.runtime.test.tsx src/mockApiClient.test.ts src/memory.test.ts`
- `bun run --cwd web build`
- browser smoke for Configure + Nowledge detect (`loomi-memory-m55-config-modal-smoke.png`)
- `bun run --cwd docs-site build`
- `git diff --check`
