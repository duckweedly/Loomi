---
title: Tool Turn Visual Polish
description: Closeout notes for compact tool execution rows and expandable tool details in Chat Canvas.
---

This slice refines tool execution transcript rendering without changing tool runtime semantics:

- Completed tools now default to a compact disclosure row that shows the human-readable tool name, status, and short request context.
- Result previews, web search sources, and longer request/result payloads stay behind expansion, keeping multi-tool turns readable.
- Running, approval-required, and failed tools keep their higher-attention treatment with status summary, phase strip where useful, and approval actions.
- Consecutive tool events in a live run transcript render as one assistant turn activity instead of separate isolated message cards.

Validation:

- `bun test web/src/components/ToolCallCard.test.tsx web/src/components/ChatCanvas.states.test.ts`
- `bun run --cwd web build`
- `bun run build` from `docs-site/`
