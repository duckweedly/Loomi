---
title: 2026-05-28 Preview Toggle Only
description: Removes the titlebar right-panel menu and keeps the titlebar utility scoped to Preview.
---

Changed:

- The titlebar right-side utility now directly toggles the Preview drawer.
- Removed the right-panel dropdown menu from the rendered shell.
- Scoped right-panel item metadata to the single Preview panel.
- Simplified the right drawer so background tasks, terminal, files, diff, and plan placeholders no longer appear from this control.

Validation:

```bash
bun test --cwd web src/App.controls.test.ts src/useWorkspaceShellState.test.ts src/rightPanelItems.test.ts src/components/RunTimeline.runtime.test.ts src/components/RunTimeline.test.tsx src/components/RightToolDrawer.preview.test.tsx
bun run --cwd web build
```
