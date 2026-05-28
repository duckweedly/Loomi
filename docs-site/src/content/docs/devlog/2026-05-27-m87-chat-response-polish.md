---
title: M87 Chat Response Polish
description: Closeout notes for run-scoped thinking hints and concise assistant output rules.
---

M87 tightens the real chat feel without adding runtime tools or architecture:

- Chat Canvas now shows a short run-scoped thinking hint while assistant content is still empty. The hint is picked from Loomi-owned locale copy, stored by `run_id` in browser storage when available, and falls back to a stable run hash during server-side rendering.
- The pending draft surface remains inline and transparent, so generating text does not create a card inside the assistant message bubble.
- Streaming markdown fragments remain hidden until final content is complete; rendered headings continue to omit visible `#` markers.
- Gateway prompt policy now adds concise output rules: answer first, no preface, do not repeat the request, and for code changes report what changed and what was verified.

Validation:

- `bun test --cwd web src/components/ChatCanvas.states.test.ts`
- `go test ./internal/runtime -run TestRunSystemPromptGuidesWorkModeToWorkspaceTools -count=1`
