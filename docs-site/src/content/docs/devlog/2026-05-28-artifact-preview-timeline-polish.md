---
title: 2026-05-28 Artifact Preview Timeline Polish
description: Makes generated documents first-class preview resources in the chat timeline.
---

Changed:

- Added a frontend `PreviewArtifact` projection that extracts document/resource metadata from tool-call result summaries and run events.
- Render completed artifact/document tools as compact document resource cards instead of inline request/result panels.
- Extract assistant-provided Markdown document payloads from `md` fenced blocks or accidental inline `md#...` output into the same compact document card path.
- Connected document resource cards to the existing right Preview drawer through `openArtifact`.
- Replaced the Preview placeholder with a resource preview state that renders Markdown/text excerpts when available.
- Normalized accidental `md#...` heading prefixes before Markdown rendering.
- Normalized ordinary `markdown#...` answer payloads as inline Markdown prose instead of extracting them as generated document cards.
- Extended `artifact.create_text` result summaries with safe `artifacts[]` resource refs containing `key`, `filename`, `mime_type`, and `display`.
- Added assistant `artifact:<key>` link extraction so final replies can cite generated documents without dumping the document body into chat.
- Kept a small waiting-for-model transcript state after terminal tool events and before continuation text arrives.

Boundaries:

- This is a frontend projection over existing safe result summaries; it does not expose raw unbounded tool output.
- Approval-required tools still render their confirmation controls.
- Non-artifact tools keep the existing compact/expandable tool-card path.

Validation:

```bash
bun test web/src/runtime/artifactPreview.test.ts web/src/runtime/markdownNormalize.test.ts web/src/components/ToolCallCard.test.tsx web/src/components/RightToolDrawer.preview.test.tsx web/src/useWorkspaceShellState.test.ts
bun test web/src/runtime/messageArtifactPreview.test.ts web/src/components/ChatCanvas.states.test.ts web/src/components/RightToolDrawer.preview.test.tsx
bun test --cwd web src/runtime/markdownNormalize.test.ts src/runtime/messageArtifactPreview.test.ts src/components/ChatCanvas.states.test.ts
bun run --cwd web build
go test ./internal/runtime ./internal/productdata
```
