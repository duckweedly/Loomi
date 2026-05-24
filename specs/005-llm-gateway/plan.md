# Implementation Plan: M5 LLM Gateway

**Branch**: `005-llm-gateway` | **Date**: 2026-05-23 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `specs/005-llm-gateway/spec.md`

## Summary

M5 connects Loomi's existing thread/message and M4 run/event/SSE foundation to real model providers through a backend model gateway. The slice adds local-only provider configuration for Anthropic, OpenAI, Gemini, and OpenAI-compatible custom providers; normalizes provider streams into Loomi run events; persists completed assistant messages; and keeps tool/function-call output as visible non-executed boundary events.

## Technical Context

**Language/Version**: Go 1.23 for API/runtime gateway; TypeScript/React/Vite in `web/`; Bun for frontend/docs commands.

**Primary Dependencies**: Existing Go stdlib HTTP stack and `pgx/v5`; existing React/Vite frontend runtime adapter; existing M3 thread/message API; existing M4 run/event/SSE API. No provider SDK dependency is required for the first M5 slice.

**Storage**: PostgreSQL via existing migrations. M5 needs migration `000004` to allow assistant messages and `model_gateway` run source while preserving M4 run/event tables.

**Testing**: `go test ./...`; `bun test ./web/src/*.test.ts ./web/src/components/*.test.ts ./web/src/runtime/*.test.ts`; `bun run --cwd web build`; API/SSE smoke with local providers; browser smoke for streaming Chat Canvas and provider-unavailable states; `bun run --cwd docs-site build` when docs are updated.

**Target Platform**: Local macOS/Darwin development, local Go API at `127.0.0.1:8080`, web renderer and Electron-compatible frontend shell, PostgreSQL local container.

**Project Type**: Web application with Go API/backend runtime and React frontend.

**Performance Goals**: First visible assistant output within 5 seconds for 95% of standard local validation prompts; stop reaches terminal state within 2 seconds in 95% of local stop attempts; provider capability checks complete quickly enough for development feedback without blocking page load.

**Constraints**: Provider credentials are local backend configuration only and never exposed to frontend, run events, or user-facing logs. Model request context is limited to current user message plus necessary recent messages from the same thread. Real API mode must not silently fall back to mock output. Actual tool execution, worker/job queue, desktop runtime, multi-agent behavior, memory/RAG, and settings UI remain out of scope.

**Scale/Scope**: Local-development M5 vertical slice; one active run per thread; provider families are Anthropic, OpenAI, Gemini, and OpenAI-compatible custom; validation covers one custom provider at `https://apikey.tgjqr.com/v1` with model `gpt-5.5` through local secret configuration.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **I. Mechanism Parity, Original Expression**: PASS. The gateway uses Loomi's own run/event/provider language and does not copy provider UI expression or another product's brand surface.
- **II. Runnable Vertical Slices**: PASS. The plan produces an end-to-end slice: submit a message, create a model-gateway run, stream normalized events, persist final assistant message, and display progress/failure states.
- **III. Core Flow Before Platform Complexity**: PASS. M5 follows the roadmap after M4 run/event/SSE and explicitly keeps worker queues, desktop runtime, tools, RAG, memory, and multi-agent behavior deferred.
- **IV. Observable Agent Execution**: PASS. Provider streams, failures, refusals, rate limits, stops, and tool-boundary requests are represented through persisted run events and visible timeline/debug states.
- **V. Safety, Permissions, and Data Boundaries**: PASS. Credentials stay server-side, provider errors are redacted, request context is limited to the current thread, and tool-like model output is not executed.
- **Technical Constraints**: PASS. The plan reuses Go stdlib/pgx, existing API/runtime boundaries, and existing frontend adapter seams rather than adding broad framework or provider SDK abstractions.
- **Development Workflow**: PASS. The spec is clarified, planning artifacts are generated, and Phase 2 task generation can proceed after this plan.
- **Documentation Definition of Done**: PASS. Implementation must update docs-site API, architecture, runbook, devlog, and spec-kit status pages and validate `bun run --cwd docs-site build`.

## Project Structure

### Documentation (this feature)

```text
specs/005-llm-gateway/
├── spec.md
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   ├── http-m5.openapi.yaml
│   ├── provider-event-mapping.md
│   ├── frontend-runtime.md
│   └── migration-cli.md
└── tasks.md
```

### Source Code (repository root)

```text
cmd/
├── loomi-api/
│   ├── main.go
│   └── main_test.go
└── loomi-seed/
    ├── main.go
    └── main_test.go

internal/
├── config/
│   ├── config.go
│   └── config_test.go
├── httpapi/
│   ├── runtime.go
│   ├── runtime_test.go
│   ├── server.go
│   └── errors.go
├── productdata/
│   ├── models.go
│   ├── repository.go
│   ├── repository_test.go
│   ├── service.go
│   └── service_test.go
└── runtime/
    ├── runner.go
    ├── simulator.go
    ├── stream.go
    ├── gateway.go
    ├── providers.go
    └── *_test.go

migrations/
├── 000004_m5_llm_gateway.up.sql
└── 000004_m5_llm_gateway.down.sql

web/src/
├── apiClient.ts
├── realApiClient.ts
├── runtime/
│   ├── executionAdapter.ts
│   ├── realExecutionAdapter.ts
│   └── *.test.ts
├── components/
│   ├── ChatCanvas.tsx
│   ├── RunTimeline.tsx
│   ├── RunRail.tsx
│   └── *.test.ts
└── state.ts

docs-site/src/content/docs/
├── architecture/llm-gateway.md
├── api/llm-gateway.md
├── runbooks/local-m5.md
├── spec-kit/workflow.md
└── devlog/2026-05-23-m5-llm-gateway.md
```

**Structure Decision**: M5 adds backend gateway code under `internal/runtime/` because provider execution is runtime behavior, not frontend state. `internal/httpapi/runtime.go` remains the HTTP boundary for run/provider endpoints. `internal/productdata/` owns durable message/run/event persistence changes. Frontend work stays inside the existing runtime adapter, Chat Canvas, and timeline components so real model output reuses M3.5/M4 UI state instead of adding a parallel model UI.

## Phase 0: Research Summary

Research is recorded in [research.md](./research.md). Key decisions:

- Use a server-side model gateway and normalize provider streams into Loomi run events.
- Use Go standard HTTP primitives first instead of provider SDKs or a unified LLM library.
- Support Anthropic, OpenAI, Gemini, and OpenAI-compatible custom providers.
- Treat custom providers as OpenAI-compatible HTTP endpoints with local base URL, key, and model configuration.
- Manage provider configuration outside the product UI for M5.
- Limit request context to the current thread.
- Add migration `000004` for assistant messages and `model_gateway` run source.
- Keep tool-like provider output as visible non-executed boundary events.
- Redact provider errors into visible failure states.

## Phase 1: Design Summary

Design artifacts are generated:

- [data-model.md](./data-model.md) defines Message, Run, Run Event, Provider Configuration, Request Context, Gateway Stream State, and Tool Boundary Event.
- [contracts/http-m5.openapi.yaml](./contracts/http-m5.openapi.yaml) defines provider capability and model-gateway run API expectations layered on M4 endpoints.
- [contracts/provider-event-mapping.md](./contracts/provider-event-mapping.md) defines Anthropic/OpenAI/Gemini/custom stream normalization into Loomi events.
- [contracts/frontend-runtime.md](./contracts/frontend-runtime.md) defines frontend runtime adapter behavior for model gateway states.
- [contracts/migration-cli.md](./contracts/migration-cli.md) defines M5 migration/readiness/rollback expectations.
- [quickstart.md](./quickstart.md) defines local provider setup, API/SSE smoke, frontend smoke, validation commands, and rollback checks.

## Post-Design Constitution Check

- **Runnable Vertical Slice**: PASS. Quickstart validates migrations, provider capability, model-gateway run creation, SSE event streaming, final assistant message persistence, stop, failure classes, and frontend smoke.
- **Core Flow Before Platform Complexity**: PASS. The design does not introduce workers, job queues, tool execution, settings UI, plugins, multi-agent behavior, RAG, memory, or desktop runtime.
- **Observable Agent Execution**: PASS. Provider stream deltas, completion, failure, rate limit, timeout, refusal, and tool-boundary events are durable run events.
- **Safety/Data Boundaries**: PASS. Provider credentials remain backend-local, errors are redacted, request context is current-thread only, and tool calls are not executed.
- **Documentation**: PASS. Docs targets and validation commands are identified.

## Complexity Tracking

No constitution violations. No additional provider SDK, worker framework, queue, settings UI, plugin runtime, external tool execution layer, or broad multi-provider abstraction is justified for M5 beyond the focused provider gateway boundary.
