# Tasks: M7 Tool Approval Execution Closure

**Input**: Design documents from `specs/011-tool-approval-execution/`

**Prerequisites**: [plan.md](./plan.md), [spec.md](./spec.md), [research.md](./research.md), [data-model.md](./data-model.md), [quickstart.md](./quickstart.md), [contracts/](./contracts/)

**Tests**: Required because this closes a safety-critical approval and execution loop.

## Phase 1: Backend approval decisions

- [X] T001 Add failing productdata repository tests for approve idempotency, deny idempotency, conflicting decisions, unknown scope, wrong thread/run, and terminal conflicts in `internal/productdata/repository_test.go`.
- [X] T002 Implement atomic approve/deny transitions in `internal/productdata/repository.go`.
- [X] T003 Add failing productdata service tests for single `tool.call.approved` and `tool.call.denied` event writes in `internal/productdata/service_test.go`.
- [X] T004 Implement approval decision service methods in `internal/productdata/service.go`.
- [X] T005 Add failing HTTP tests for approve/deny route scoping, retries, unknown ids, terminal conflicts, and wrong-user boundaries in `internal/httpapi/runtime_test.go`.
- [X] T006 Implement approve/deny handlers in `internal/httpapi/runtime.go`.
- [X] T007 Register approve/deny routes in `internal/httpapi/server.go`.

## Phase 2: Worker execution closure

- [X] T008 Add failing runtime worker tests for approved current-time execution, single execution after repeated approve, and denial no-execution in `internal/runtime/worker_test.go`.
- [X] T009 Add failing runtime tool tests for redacted success and redacted failure summaries in `internal/runtime/tools_test.go`.
- [X] T010 Add failing service tests for `tool.call.executing`, `tool.call.succeeded`, and `tool.call.failed` terminal guards in `internal/productdata/service_test.go`.
- [X] T011 Implement worker resume/execution for approved `runtime.get_current_time` in `internal/runtime/worker.go` and minimal job wake hook in `internal/runtime/jobs.go` only if needed.
- [X] T012 Implement executing/succeeded/failed event writes and projection updates in `internal/productdata/service.go` and `internal/productdata/repository.go`.
- [X] T013 Add or update SSE replay tests for approval/execution/result events in `internal/runtime/stream_test.go` or `internal/httpapi/runtime_test.go`.

## Phase 3: Frontend actions and replay

- [X] T014 Add failing real API client tests for approve/deny methods in `web/src/realApiClient.test.ts`.
- [X] T015 Implement approve/deny client methods in `web/src/realApiClient.ts`.
- [X] T016 Add failing ToolCallCard interaction tests for approve loading, deny loading, disabled states, and error state in `web/src/components/ToolCallCard.test.tsx`.
- [X] T017 Wire ToolCallCard approve/deny controls to real actions in `web/src/components/ToolCallCard.tsx`.
- [X] T018 Add failing adapter tests for approved, succeeded, failed, and denied mappings in `web/src/runtime/realExecutionAdapter.test.ts` or `web/src/runtime/executionAdapter.test.ts`.
- [X] T019 Implement frontend event mappings in `web/src/runtime/realExecutionAdapter.ts` and related runtime state files.
- [X] T020 Update RunRail and RunTimeline rendering/tests for approval/execution/result/denied states in `web/src/components/RunRail.tsx` and `web/src/components/RunTimeline.tsx`.

## Phase 4: Documentation and validation

- [X] T021 Update `docs-site/src/content/docs/architecture/tool-call-approval.md`.
- [X] T022 Update `docs-site/src/content/docs/api/tool-call-approval.md`.
- [X] T023 Update `docs-site/src/content/docs/runbooks/local-m7.md`.
- [X] T024 Add `docs-site/src/content/docs/devlog/2026-05-25-m7-approval-execution.md`.
- [X] T025 Update `docs-site/src/content/docs/roadmap/current-status.md`.
- [X] T026 Run `go test ./internal/productdata ./internal/runtime ./internal/db ./internal/httpapi ./cmd/...`.
- [X] T027 Run frontend tests.
- [X] T028 Run `bun run --cwd web build`.
- [X] T029 Run `bun run --cwd docs-site build`.
- [X] T030 Perform browser smoke for approve -> executing/succeeded and deny -> clear terminal state.
