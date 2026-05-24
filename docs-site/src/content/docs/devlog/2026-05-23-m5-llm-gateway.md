---
title: 2026-05-23 M5 LLM Gateway Devlog
description: Implementation notes, validation results, limitations, and next steps for M5.
---

## Completed scope

M5 now connects the existing thread/message and run/event/SSE foundation to Anthropic, OpenAI, Gemini, and OpenAI-compatible custom providers through a backend model gateway.

Implemented slice:

- migration support for assistant messages and `model_gateway` runs
- local provider configuration and redacted provider capability API
- backend HTTP streaming providers without provider SDK dependencies
- current-thread request context through the durable trigger message
- provider-normalized `model_output_delta`, `model_output_completed`, redacted failure, and `tool_call_blocked` events
- final assistant message persistence on successful completion
- no hidden mock fallback in real API model-gateway send flow
- frontend assistant draft rendering for streaming model output
- RunRail and Chat Canvas display for provider failures and tool-boundary events

`006-streaming-chat-runtime` remains separate worktree work and was not mixed into this implementation.

## Safety notes

Provider credentials remain backend-local. API responses and run events exclude API keys, Authorization headers, raw provider request payloads, raw provider error bodies, and custom provider URL path/query/userinfo fragments.

Provider tool/function-call output is visible as `tool_call_blocked`; M5 does not execute tools or external actions.

## Validation log

Validated during implementation:

```bash
go test ./internal/productdata ./internal/runtime
go test ./internal/httpapi
go test ./internal/productdata ./internal/runtime
go test ./...
bun test ./web/src/runtime/realExecutionAdapter.test.ts
bun test ./web/src/runtime/realExecutionAdapter.test.ts ./web/src/state.runtime.test.ts ./web/src/realApiClient.test.ts
bun test ./web/src/*.test.ts ./web/src/components/*.test.ts ./web/src/runtime/*.test.ts
bun run --cwd web build
bun run --cwd docs-site build
```

Latest recorded results:

- `go test ./...` passed.
- frontend Bun tests passed with 94 tests and 213 expectations after the mock Stop browser regression, stopped-draft UI, late-event guard, and duplicate stream-event dedupe fixes.
- `bun run --cwd web build` passed.
- `bun run --cwd docs-site build` passed.
- mock-mode browser smoke passed for the seeded running thread Stop path with no new console errors.
- late provider events after a terminal frontend run are ignored, preventing stopped runs from being revived by delayed deltas.
- isolated `006-streaming-chat-runtime` frontend validation passed in its separate worktree with 97 tests and a successful build.

Still pending as local/manual smoke:

- US1 API/SSE quickstart sections 5-8 with a live local database and configured provider.
- US2 failure/stop quickstart sections 9-10 with provider-specific failure scenarios.
- US3 tool-boundary quickstart section 10 against a live provider/tool-call response.
- Real API browser smoke for model gateway with live provider credentials.

## Known local blockers

Real provider/API/SSE browser smoke from `specs/005-llm-gateway/quickstart.md` still requires local database startup and real provider credentials. Until those are supplied locally, automated Go and frontend tests are the primary validation evidence.

## Remaining follow-up

- Run docs build after these docs updates.
- Run or record a real-provider smoke when local credentials are available.
- Keep actual tool execution, worker/job queues, settings UI, RAG, memory, desktop runtime, and multi-agent behavior deferred to later specs.
