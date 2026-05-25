# Quickstart: RunContext Pipeline Foundation

## Purpose

Validate the M9 foundation slice after implementation: the worker prepares durable RunContext, resolves MVP tools, invokes runtime, finalizes, and exposes a safe stage trace in Timeline/debug views.

## Backend Validation

Run:

```bash
go test ./internal/productdata ./internal/runtime ./internal/httpapi ./cmd/...
```

Expected coverage:

- RunContext loader restores run/thread/messages/job metadata/provider route/enabled tools.
- Missing required context fails before runtime/provider invocation.
- Worker path uses RunContext for queued model-gateway and local simulated runs.
- Tool-result continuation still succeeds through the existing M7 path.
- Pipeline stages record started/completed/failed events without duplicate terminal writes.

## Web Validation

Run related runtime/timeline tests, including any new or updated tests under:

```text
web/src/runtime/realExecutionAdapter.test.ts
web/src/runtime/runtimeEventGroups.test.ts
web/src/components/RunTimeline.runtime.test.ts
web/src/components/RunRail.runtime.test.ts
```

Then run:

```bash
bun run --cwd web build
```

Expected coverage:

- Stage events map to stable frontend runtime event types.
- Live and replayed traces show the same ordered stages.
- Failed stage summaries are visible and redacted.

## Docs Validation

Run:

```bash
bun run --cwd docs-site build
```

Expected docs updates:

- `docs-site/src/content/docs/architecture/worker-job-pipeline.md`
- `docs-site/src/content/docs/api/worker-job-pipeline.md`
- `docs-site/src/content/docs/runbooks/local-m9.md`
- `docs-site/src/content/docs/roadmap/current-status.md`
- `docs-site/src/content/docs/spec-kit/workflow.md`
- `docs-site/src/content/docs/devlog/2026-05-25-m9-run-context-pipeline-foundation.md`

## Browser Smoke

1. Start the local API/worker and web app in real API mode.
2. Create or open a thread with a user message.
3. Start a run using the current provider/model route.
4. Open Timeline/debug details for the run.
5. Confirm the trace includes, in order:
   - context prepared
   - tools resolved
   - runtime invoked
   - finalized
6. Refresh/reconnect and confirm history replay shows the same trace.
7. For a controlled missing-context or provider-route failure, confirm the failed stage is visible and redacted.

## Non-Goals to Verify

- No new worker queue implementation.
- No Redis or external queue service.
- No Persona/Skill, MCP, Memory/RAG, Sandbox, Desktop Runtime, multi-agent, shell/filesystem/browser automation tools.
- No raw provider payloads, credentials, raw tool results, file contents, or shell output in stage metadata.

## Validation Results (2026-05-25)

- PASS: `go test ./internal/productdata ./internal/runtime ./internal/httpapi ./cmd/...`
- PASS: `bun test ./web/src/realApiClient.test.ts ./web/src/runtime/realExecutionAdapter.test.ts ./web/src/runtime/runtimeEventGroups.test.ts ./web/src/components/RunTimeline.runtime.test.ts ./web/src/components/RunRail.runtime.test.ts`
- PASS: `bun run --cwd web build`
- PASS: `bun run --cwd docs-site build`
- PASS: Browser smoke created a real run and confirmed Timeline/debug trace displayed `prepare_context`, `resolve_tools`, `invoke_runtime`, and `finalize` with redacted metadata. Screenshot: `/tmp/loomi-m9-smoke-stages.png`

Note: `.specify/scripts/bash/check-prerequisites.sh --json --require-tasks --include-tasks` was run after implementation and reported the current branch is `main`, not a Spec Kit feature branch. The feature pointer still targets `specs/014-run-context-pipeline-foundation/plan.md`.
