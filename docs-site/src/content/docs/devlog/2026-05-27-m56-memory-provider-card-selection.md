---
title: M56 Memory Provider Card Selection
description: Card-based provider selector for Settings > Memory.
---

## Completed

- Replaced the Memory provider segmented selector with selectable provider cards.
- Added responsive styling so the cards collapse to one column on narrow screens.
- Preserved the existing provider update and Configure modal behavior.

## Validation

- `bun test --cwd web src/components/SettingsView.runtime.test.tsx src/mockApiClient.test.ts src/memory.test.ts`
- `bun run --cwd web build`
- browser smoke for Nowledge card + Configure modal (`loomi-memory-m56-provider-cards-smoke.png`)
- `bun run --cwd docs-site build`
- `git diff --check`
