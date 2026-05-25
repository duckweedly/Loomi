# Implementation Plan: Tool Result Model Continuation

**Branch**: `codex/012-tool-result-model-continuation` | **Date**: 2026-05-25 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/012-tool-result-model-continuation/spec.md`

## Summary

Add the minimal continuation slice after approval-gated `runtime.get_current_time` succeeds: project the redacted tool result from run events, construct one gateway-neutral synthetic tool result item for the second provider call, stream the final assistant answer through the existing run-event/SSE path, and finalize exactly one assistant message. Denied and tool-failed paths remain terminal without a second model call.

## Technical Context

**Language/Version**: Go backend/worker/runtime, TypeScript React frontend, Starlight docs.

**Primary Dependencies**: Existing `internal/productdata`, `internal/runtime` provider gateway/worker pipeline, existing web runtime adapters and Timeline/ToolCallCard.

**Storage**: Existing PostgreSQL messages, runs, run_events, and tool_calls projection. No new durable `messages.role = tool` by default.

**Testing**: Go unit/integration tests, web TypeScript tests, backend SSE/event ordering tests, local browser smoke after implementation.

**Target Platform**: Local desktop-feeling web app backed by local API/worker.

**Project Type**: Web application plus local API/worker runtime.

**Performance Goals**: The second provider request starts within 2 seconds after `tool_call_succeeded` is persisted in local smoke when the provider is responsive.

**Constraints**: One active run per thread. One tool call per run. One continuation provider call per run. Redacted result only. No shell/filesystem/arbitrary network/MCP/browser automation/multi-agent/RAG. Do not change Window A approve/deny API unless implementation exposes a blocking contract gap.

**Scale/Scope**: Local M7 continuation slice for one approved `runtime.get_current_time` call and one final assistant answer.

## Constitution Check

- **I. Mechanism Parity, Original Expression**: PASS. Uses Loomi's run/event/tool-call language.
- **II. Runnable Vertical Slices**: PASS. Success path is demonstrable end to end from tool success to final assistant answer.
- **III. Core Flow Before Platform Complexity**: PASS. Adds only the next core flow step, not platform tools or multi-agent loops.
- **IV. Observable Agent Execution**: PASS. Continuation is represented through ordered run events, timeline grouping, and final message state.
- **V. Safety, Permissions, and Data Boundaries**: PASS. Tool results are redacted before persistence and provider continuation.

## Project Structure

### Documentation (this feature)

```text
specs/012-tool-result-model-continuation/
├── spec.md
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   ├── provider-continuation.md
│   ├── run-event-sse.md
│   └── frontend-runtime.md
├── checklists/
│   └── requirements.md
└── tasks.md
```

### Source Code (implementation target)

```text
internal/
├── productdata/
│   ├── models.go                 # event/status constants if phase metadata needs typed helpers
│   ├── repository.go             # read current run messages and tool result projection
│   └── service.go                # continuation projection and final assistant persistence use cases
└── runtime/
    ├── gateway.go                # gateway-neutral continuation request/response contract
    ├── providers.go              # provider-specific synthetic tool result serialization
    ├── runner.go                 # initial model phase and continuation handoff
    ├── queued_runner.go          # post-tool-success resume path
    ├── tools.go                  # redacted result shape remains allowlisted
    └── worker.go                 # run terminal behavior around continuation

web/src/
├── apiClient.ts / realApiClient.ts          # event metadata mapping if needed
├── domain.ts                                # model phase/draft metadata if needed
├── runtime/realExecutionAdapter.ts          # assistantDraft phase handling
└── components/
    ├── RunTimeline.tsx
    ├── RunRail.tsx
    └── ToolCallCard.tsx

docs-site/src/content/docs/
├── architecture/tool-result-continuation.md
├── api/tool-call-approval.md
├── runbooks/local-m7.md
├── devlog/2026-05-25-tool-result-continuation.md
└── roadmap/current-status.md
```

**Structure Decision**: Reuse `internal/runtime` for provider continuation and worker flow, `internal/productdata` for projections/final message persistence, `web/src/runtime` for replay/draft reconstruction, and docs-site for implementation documentation. No new service, queue, tool family, or provider package is planned.

## Phase 0: Research Decisions

- Build continuation context from persisted conversation messages and current run-event projection.
- Use an in-memory synthetic tool result item for provider adapters.
- Keep persistent message roles unchanged for MVP.
- Gateway owns a provider-neutral continuation request.
- MVP allows exactly one tool call and one continuation call per run.
- If continuation requests another tool, fail safely with a redacted unsupported-loop event.
- Denied and tool-failed paths do not call the model again.
- Use ordinary model delta/final events for the second phase, with phase metadata when needed.
- Redact before persistence and before provider continuation.

See [research.md](./research.md).

## Phase 1: Design Artifacts

- [data-model.md](./data-model.md) defines Continuation Context, Synthetic Tool Message, Tool Result Projection, Model Phase, and Assistant Draft.
- [contracts/provider-continuation.md](./contracts/provider-continuation.md) defines the gateway/provider continuation contract.
- [contracts/run-event-sse.md](./contracts/run-event-sse.md) defines event ordering and replay expectations.
- [contracts/frontend-runtime.md](./contracts/frontend-runtime.md) defines assistantDraft and Timeline behavior.
- [quickstart.md](./quickstart.md) defines local validation scenarios.

## Window A Dependencies

- Approve/deny endpoints are available and idempotent.
- Approved `runtime.get_current_time` execution records `tool_call_executing` then one terminal `tool_call_succeeded` or `tool_call_failed`.
- `tool_call_succeeded` metadata contains a redacted result summary usable by UI and model continuation.
- Deny path records `tool_call_denied` and prevents tool execution.
- Window A does not need to implement model continuation.

If Window A lands different result field names, 012 should adapt at the projection boundary rather than changing approve/deny endpoint semantics.

## Complexity Tracking

No constitution violations. The synthetic provider tool result item is a minimal adapter boundary, not a durable schema expansion. One-loop limit avoids multi-agent or general agent-loop complexity.
