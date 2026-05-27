---
title: M88 Chat Response Polish Followups
description: Follow-up polish for run-level thinking, concise replies, and large tool result continuation.
---

M88 continues the Arkloop-inspired response polish while keeping Loomi's own copy, visuals, and safety boundary:

- Run Rail now shows a run-level thinking line when the assistant draft is still empty, so the user sees progress even before a message bubble exists.
- Thinking copy uses text shimmer and a short incremental update when elapsed seconds change; there is no dot loader inside the response surface.
- Terminal runs can show a safe thought-summary row from allowlisted metadata only. Hidden/raw thinking metadata is stripped from frontend events.
- Large string fields inside redacted tool results are compacted before provider continuation. The compact result keeps path/status/error signals and a compaction marker, while small tool results are unchanged.
- The default Loomi persona prompt now explicitly asks for result-first, brief answers and forbids exposing hidden chain-of-thought.

Validation:

```bash
bun test --cwd web src/runtime/incrementalTypewriter.test.ts src/runtime/realExecutionAdapter.test.ts src/components/RunRail.polish.test.ts src/components/ChatCanvas.states.test.ts
go test ./internal/runtime -run 'TestCompactToolResult|TestGatewayCompactsLargeContinuationToolResult|TestGatewayBuildsContinuationContextFromToolResultEvents' -count=1
go test ./internal/productdata -run TestBuiltInPersonaDefaultPromptStaysConcise -count=1
```

Boundaries:

- No new tools, provider routes, Docker/Redis/Firecracker, or multi-agent runtime were added.
- Tool result compaction applies only to provider continuation input, not to stored run events.
- No token, provider key, raw hidden thinking, or raw provider payload is rendered or documented.
