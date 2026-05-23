# Quickstart: Frontend Agent Runtime Skeleton

## Goal

Validate that M3.5 provides a frontend-only Agent runtime skeleton before backend run/event/SSE exists.

## Prerequisites

- Repository dependencies installed.
- Use mock mode by leaving `VITE_LOOMI_API_BASE_URL` unset for the web renderer.
- Use real API mode only to verify backend capability unavailable behavior.

## Automated validation

From the repository root:

```bash
bun test ./web/src/**/*.test.ts "web/vite.config.test.ts"
bun run --cwd web build
bun run --cwd docs-site build
```

Expected results:

- All web tests pass.
- Web build completes without TypeScript errors.
- Docs build completes when documentation is updated.

## Browser smoke: state skeleton

1. Start the web renderer or desktop dev shell.
2. Open Chat mode.
3. Verify a new empty thread shows a clear empty-thread state rather than a blank canvas.
4. Force or simulate loading state and verify the main area shows loading.
5. Force or simulate error state and verify the main area shows a concise Chinese failure state and retry action.
6. Configure real API mode without runtime support and attempt runtime execution.
7. Verify the main area shows backend capability unavailable instead of mock execution.

## Browser smoke: mock success script

1. Open Chat mode in mock mode.
2. Select or create a Chat thread.
3. Send a user message.
4. Verify the user message appears immediately.
5. Verify Run Timeline shows these milestones in order:
   - run created
   - context loading
   - assistant thinking
   - assistant drafting
   - assistant message completed
   - run completed
6. Verify Agent state motion changes during execution and ends in done state.
7. Verify exactly one final assistant reply appears in Chat Canvas.

## Browser smoke: mock failure script

1. Select the mock failure scenario through the test control or configured mock script selector.
2. Send a user message.
3. Verify the user message appears immediately.
4. Verify Run Timeline ends with failed event.
5. Verify Chat Canvas shows failed state.
6. Verify no successful assistant reply is appended.

## Browser smoke: stop run

1. Start a mock runtime script.
2. Stop it before completion.
3. Verify Timeline shows stopped.
4. Verify Chat Canvas shows stopped/failed state.
5. Verify Agent state motion shows error/stopped semantics.
6. Verify no later script events change the stopped run to completed.

## Stale event smoke

1. Start a mock run in one Chat thread.
2. Switch to another thread before the script completes.
3. Verify later events from the old run do not change the visible Chat Canvas, Timeline, or Agent state motion for the newly selected thread.

## Documentation validation

Implementation must update `docs-site/src/content/docs/` with:

- architecture page explaining frontend runtime state/adapters;
- runbook page explaining mock success/failure/stopped smoke;
- devlog entry with validation results and known limitations.

Then run:

```bash
bun run --cwd docs-site build
```
