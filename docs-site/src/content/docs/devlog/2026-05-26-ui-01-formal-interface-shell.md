---
title: 2026-05-26 UI-01 Formal Interface Shell
description: First-round shell redesign notes, validation targets, and known follow-up scope.
---

## Scope

UI-01 creates the first reviewable light interface shell for Loomi:

- light desktop-window outer shell
- narrower native-feeling sidebar
- large white chat canvas
- centered content column
- fixed bottom composer
- preserved Chat/Work, Stop, provider warning, Settings, Tools, and RunRail entry points

The reference image is used only for layout proportion and shell mechanism. Loomi keeps its own identity, wording, and existing controls.

## Validation Targets

Closeout for this slice requires:

```bash
go test ./...
bun test --cwd web
bun run --cwd web build
bun run --cwd docs-site build
git diff --check
```

Browser smoke must open the local web app and verify:

- overall layout direction
- Chat mode accepts input
- Work mode accepts input and can start a run
- active run Stop is visible
- Settings > Tools opens
- console error count is 0
- screenshot path is recorded

## Non-goals

No pixel-level clone, no new feature, no state-management rewrite, no backend API change, no database change, no new tool capability, no runtime behavior change, no provider behavior change, no memory behavior change, and no M38/activity recorder continuation.
