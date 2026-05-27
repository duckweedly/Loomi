---
title: M92 Agent Tool Selection And Failure Recovery
description: Provider prompt strategy, load_tools query lookup, compaction safety, and actionable terminal failures.
---

M92 addresses the real-use failure pattern where the model had tools available but skipped them, a tool chain ended with unreadable `[redacted]` context, or a terminal state only surfaced as a generic run failure.

Implemented changes:

- Added explicit Work-mode provider guidance for directory, content, modification, and shell/validation tool choice.
- Made `tool.load_tools` provider-facing schema query-only and optional, while validation accepts `query`, `queries`, empty `names`, or no query for bounded catalog listing.
- Compacts continuation tool results without collapsing benign terminal summaries to `[redacted]`; sensitive lines remain redacted.
- Normalizes tool request and execution failures into user-actionable categories: provider, validation, permission, workspace binding, and bounded timeout/limit.

Regression coverage added:

- Provider prompt includes the tool selection strategy.
- `tool.load_tools` accepts query-only and empty-query catalog lookups.
- Tool result compaction preserves readable summaries and redacts secrets.
- Existing terminal-run guards continue to reject late events through product-data and frontend replay boundaries.

Validation notes:

- `go test ./internal/productdata -run 'TestValidateDiscoveryToolCalls|TestRunEventKeepsAssistantFinalContentWithBenignTokenWords' -count=1` passed.
- `go test ./...` passed after the sandbox process branch was completed.
- M93 follow-up validation is documented as memory-backed in-process restore, not API-restart durable persistence.

Known gaps versus Arkloop:

- Loomi has bounded continuation and event persistence, but not Arkloop-style durable rollout item orchestration for every agent step.
- Terminal lifecycle is visible through run/tool events, and sandbox process records can be rebuilt from the memory-backed repository inside the current API process. Loomi still does not have productdata/Postgres-backed process recovery across API process restarts.
- Tool choice guidance is prompt/schema-level; Loomi still does not have a planner that can enforce tool order independently of the provider.
