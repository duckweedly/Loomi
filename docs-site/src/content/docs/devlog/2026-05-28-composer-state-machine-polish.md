---
title: Composer State Machine Polish
description: First pass on Loomi composer run-state, context menu, and workspace context chip handling.
---

## Completed

- Active runs keep the primary composer action in stop mode, and that stop button stays enabled even when send is otherwise blocked.
- Submitting a prompt clears the draft and pending attachments, then returns focus to the textarea so follow-up typing does not require re-clicking.
- The `+` context menu keeps the existing Loomi floating menu primitive, but its composer placement is clamped to the viewport and falls below the trigger when there is not enough room above.
- The workspace chip now reads as business context with explicit empty/selected states, hover/focus styling, and an accessible label instead of behaving like a generic file picker.

## Validation

- `bun test web/src/components/Composer.test.ts`
- `bun test web/src/components/Composer.test.ts web/src/components/LoomiMenu.test.ts`
- `bun test --cwd web`
- `bun run --cwd web build`
- `bun run build` from `docs-site/`

