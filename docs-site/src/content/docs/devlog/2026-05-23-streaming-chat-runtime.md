---
title: 2026-05-23 Streaming Chat Runtime Devlog
description: 006 frontend streaming chat runtime implementation, validation results, limitations, and next steps.
---

## Completed work

006 completes the frontend streaming chat runtime UX slice:

- Chat Canvas renders a first-class assistant draft bubble for pending, streaming, completed, failed, stopped, and recovering run states.
- Runtime event mapping preserves ordered assistant deltas, blocks stale terminal promotions, and keeps token/provider metadata out of assistant message text.
- RunRail and RunTimeline group events into Run lifecycle, Model stream, Worker/job, and Error, with error-like events taking precedence.
- Capability status distinguishes mock, local simulated, real model, backend unavailable, model setup missing, provider unavailable, stream disconnected, and run recovering states.
- Composer supports guarded send/continue, stop, retry, and regenerate; regenerated responses preserve the previous assistant response and create a new attempt.
- Thread/message surfaces keep no-thread, empty-thread, loading, error, history, terminal run, and recovering states synchronized across Chat Canvas, ThreadSidebar, Composer, and Timeline.

## Validation results

- `bun test ./web/src/**/*.test.ts "web/vite.config.test.ts"` — PASS, 131 tests.
- `bun run --cwd web build` — PASS.
- `bun run --cwd docs-site build` — PASS, 29 pages built.
- Browser smoke with `bun run --cwd web dev --host 127.0.0.1 --port 5179` — PASS in mock mode. Verified Chat mode selection, user send, completed assistant bubble, Regenerate preserving prior response and completing a new attempt, Run details grouped timeline, Mock capability copy, and no browser console warnings/errors.

## Fixes found during smoke

Browser smoke exposed that mock Regenerate could create a pending run that never advanced. The root cause was twofold: retry/regenerate used a local pending run instead of starting a subscribable runtime run, and mock API run IDs could collide with the adapter's internal store IDs. The fix added `mockApiClient.startRun`, made retry/regenerate use the API start-run seam when available, kept success script terminal status on `run.completed`, and added a mock API regression test for subscribable retry/regenerate runs.

## Known limitations

- Mock scripts remain deterministic and short; they validate state semantics rather than realistic token timing.
- Real API mode still reflects M4 local simulated execution unless a later backend slice provides real model events.
- Retry/regenerate start a new runtime attempt for the selected thread; persistent attempt lineage for real API remains limited by current M4 contracts.
- Work mode remains outside this runtime execution slice.

## Next steps

- M5 should map real LLM gateway deltas/final/error/usage events into the same runtime event and assistant draft semantics.
- Later backend work can persist regenerate/retry attempt lineage once the run/message contract supports it directly.
