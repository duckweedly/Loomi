# Quickstart: Streaming Chat Runtime

## Goal

Validate that the frontend presents realistic LLM streaming chat behavior while preserving existing mock/real runtime boundaries and staged backend capability honesty.

## Prerequisites

- Repository dependencies installed.
- Use mock mode by leaving `VITE_LOOMI_API_BASE_URL` unset for the web renderer.
- Use real API mode with a ready local API when validating real run/event/SSE behavior.
- Run from the repository root unless a command uses `--cwd`.

## Automated validation

```bash
bun test ./web/src/**/*.test.ts "web/vite.config.test.ts"
bun run --cwd web build
bun run --cwd docs-site build
```

Expected results:

- All frontend tests pass.
- Web build completes without TypeScript errors.
- Docs build completes after documentation updates.

## Browser smoke: streaming assistant bubble

1. Start the web renderer with `bun run --cwd web dev`.
2. Open Chat mode in mock mode.
3. Select or create a Chat thread.
4. Send a non-empty user message.
5. Verify the user message appears immediately.
6. Verify a pending assistant bubble appears before final content.
7. Verify partial assistant output grows in the same bubble.
8. Verify successful completion produces one final assistant response without duplicate text.

## Browser smoke: failed and stopped drafts

1. Select a failure-capable mock scenario or simulate a failed run.
2. Send a user message.
3. Verify partial draft content, if any, remains visible after failure.
4. Verify the failed state offers a retry path.
5. Start another active run.
6. Stop the run before completion.
7. Verify partial draft content remains visible with stopped status.
8. Verify later completion events do not convert the stopped draft into a completed message.

## Browser smoke: regenerate and retry

1. Complete a successful assistant response.
2. Trigger Regenerate.
3. Verify the previous assistant response remains visible.
4. Verify a new assistant attempt appears and is associated with a new run.
5. Trigger or simulate a failed run.
6. Verify Retry starts a new attempt without clearing recoverable user input or failure context.

## Browser smoke: timeline event grouping

1. Replay or simulate a run containing lifecycle, model, worker/job, usage, retry/cancel, and error events.
2. Verify lifecycle events appear in Run lifecycle.
3. Verify model deltas, final output, and token usage appear in Model stream.
4. Verify queue, worker claimed, and retrying events appear in Worker/job.
5. Verify provider, stream, backend, and run failure events appear in Error and are visually distinct.

## Browser smoke: backend capability status

1. Verify mock mode is labeled as mock or deterministic local behavior.
2. Configure real API mode with local simulated execution and verify it is not labeled as real model output.
3. Simulate backend unavailable and verify the UI does not show model-thinking copy.
4. Simulate model setup missing or provider unavailable and verify the status is distinct from backend unreachable.
5. Simulate stream disconnect during an active run and verify stream-disconnected or recovery status remains visible until reconciliation.

## Browser smoke: thread and message states

1. Load the app with no selected thread and verify no stale messages appear.
2. Select an empty thread and verify the empty state invites conversation.
3. Simulate message loading and verify loading/skeleton state appears.
4. Simulate message load failure and verify retry is available without clearing selected thread context.
5. Select a thread with persisted assistant message and run events and verify Chat Canvas and Timeline agree on the latest run outcome.

## Documentation validation

Implementation must update `docs-site/src/content/docs/` with the most relevant pages for changed runtime behavior:

- `architecture/frontend-agent-runtime.md` for assistant draft bubble, runtime state derivation, timeline grouping, and composer actions.
- `architecture/run-event-sse.md` and `api/run-event-sse.md` if event payload interpretation or grouping assumptions change.
- `runbooks/frontend-runtime-smoke.md` for smoke scenarios above.
- `spec-kit/workflow.md` for links or status of this feature spec if the project workflow index tracks active features.
- `devlog/2026-05-23-streaming-chat-runtime.md` for validation results, known limitations, and next steps.

Then run:

```bash
bun run --cwd docs-site build
```
