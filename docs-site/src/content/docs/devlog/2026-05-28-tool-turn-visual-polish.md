---
title: Tool Turn Visual Polish
description: Closeout notes for compact tool execution rows and expandable tool details in Chat Canvas.
---

This slice refines tool execution transcript rendering without changing tool runtime semantics:

- Completed tools now default to a compact disclosure row that shows the human-readable tool name, status, and short request context.
- Result previews, web search sources, and longer request/result payloads stay behind expansion, keeping multi-tool turns readable.
- Running, approval-required, and failed tools keep their higher-attention treatment with status summary, phase strip where useful, and approval actions.
- Consecutive tool events in a live run transcript render as one assistant turn activity instead of separate isolated message cards.
- If a run fails, stops, or enters recovery after terminal tool events and before final assistant text, Chat Canvas now appends an explicit assistant status row after the tool group instead of leaving the turn visually blank.
- Work-mode prompting now forbids final answers that say Loomi still needs to keep reading source files when workspace tools are available; the model should request the next workspace read/search/list call in the same run instead of pausing for the user to say "continue".

Validation:

- `bun test web/src/components/ToolCallCard.test.tsx web/src/components/ChatCanvas.states.test.ts`
- `bun test --cwd web ./src/components/ChatCanvas.states.test.ts -t 'keeps failure feedback visible|keeps recovery feedback visible|keeps a waiting-for-model state'`
- `bun run --cwd web build`
- `bun run build` from `docs-site/`
