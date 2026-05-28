---
title: M92 Agent Tool Selection And Failure Recovery
description: Provider prompt strategy, load_tools query lookup, compaction safety, and actionable terminal failures.
---

M92 addresses the real-use failure pattern where the model had tools available but skipped them, a tool chain ended with unreadable `[redacted]` context, or a terminal state only surfaced as a generic run failure.

Implemented changes:

- Added explicit Work-mode provider guidance for directory, content, modification, and shell/validation tool choice.
- Added a runtime guard for the worst real-use tool planning failures: directory inventory cannot start with grep/glob/read, and repeated same-argument read/list/grep calls fail safely with `tool_planner_guardrail`.
- Made `tool.load_tools` provider-facing schema query-only and optional, while validation accepts `query`, `queries`, empty `names`, or no query for bounded catalog listing.
- Compacts continuation tool results without collapsing benign terminal summaries to `[redacted]`; sensitive lines remain redacted.
- Normalizes tool request and execution failures into user-actionable categories: provider, validation, permission, workspace binding, and bounded timeout/limit.
- Normalizes raw structured provider completion payloads before final assistant persistence, so benign project names and URLs are not replaced by `[redacted]` or shown as raw tool protocol JSON.

Regression coverage added:

- Provider prompt includes the tool selection strategy.
- Runtime guard rejects directory-inventory grep starts while allowing `workspace.tree_summary`.
- Runtime guard rejects repeated same-argument workspace read requests.
- `tool.load_tools` accepts query-only and empty-query catalog lookups.
- Tool result compaction preserves readable summaries and redacts secrets.
- Final assistant message tests cover structured provider payload extraction and terminal-run late model/tool rejection.

Validation notes:

- `go test ./internal/productdata -run 'TestValidateDiscoveryToolCalls|TestRunEventKeepsAssistantFinalContentWithBenignTokenWords' -count=1` passed.
- `go test ./...` passed after the sandbox process branch was completed.
- M93 follow-up validation has since moved from memory-backed restore to productdata/Postgres-backed durable process summaries.

Known gaps versus Arkloop:

- Loomi has bounded continuation and event persistence, but not Arkloop-style durable rollout item orchestration for every agent step.
- Terminal lifecycle is visible through run/tool events, and sandbox process summaries can be rebuilt from productdata/Postgres after API restart. Loomi still does not reattach or restart OS processes after API restart.
- Tool choice is now prompt/schema plus a small deterministic guard. Loomi still does not have a general planner that can infer arbitrary multi-step strategy, reorder tool calls, or recover every bad provider choice.
