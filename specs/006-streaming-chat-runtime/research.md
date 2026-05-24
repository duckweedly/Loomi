# Research: Streaming Chat Runtime

## Decision: Render assistant draft as a transient chat bubble

**Rationale**: The current runtime already tracks `assistantDraft`, but users cannot see partial model output as a natural chat response. Rendering the selected run's draft as a transient assistant bubble gives immediate feedback, supports pending/streaming/failed/stopped/recovering states, and lets successful finalization become a normal assistant message exactly once.

**Alternatives considered**:

- Render deltas only in the timeline. Rejected because the primary user experience is the conversation, not the debug rail.
- Append a new message for every delta. Rejected because it creates duplicated or noisy chat history and complicates finalization.
- Wait for final output before rendering. Rejected because it does not satisfy the streaming Chat Canvas goal.

## Decision: Preserve prior assistant responses when regenerating

**Rationale**: Clarification selected the behavior that Regenerate preserves the previous assistant response and adds a new assistant attempt. This matches Loomi's observability principle because users and developers can explain each attempt through messages, runs, and events without silently replacing history.

**Alternatives considered**:

- Replace the previous assistant response in place. Rejected because it loses execution history and makes timeline/debug harder to reason about.
- Show only the latest response by default with prior attempts hidden. Rejected for this slice because it adds attempt-history UI complexity before the basic streaming runtime is stable.

## Decision: Group timeline events by frontend semantic purpose

**Rationale**: M5/M6 will add model, provider, usage, queue, worker, retry, cancel, and stream events. Grouping in the frontend by Run lifecycle, Model stream, Worker/job, and Error allows richer event vocabulary without waiting for a backend schema redesign, while still letting the same selected run drive Chat Canvas and Timeline state.

**Alternatives considered**:

- Keep a single chronological list only. Rejected because richer event volume would obscure failures and model progress.
- Require backend to pre-group events. Rejected because frontend can group from event type/category now and backend event vocabulary is still evolving.
- Create separate panels for every event family. Rejected because it fragments the debug view and is premature for the current slice.

## Decision: Derive capability status from current frontend runtime facts

**Rationale**: Users need to distinguish mock, local simulated, real model, backend unavailable, model setup missing, provider unavailable, stream disconnected, and run recovering. The frontend already knows mode, stream state, run state, and adapter capability; deriving one user-readable status prevents misleading copy such as implying the model is thinking when the backend is unreachable.

**Alternatives considered**:

- Show raw adapter capability only. Rejected because `available`/`unavailable` is not enough to diagnose provider setup, stream disconnect, or recovery.
- Hide mode status in debug-only UI. Rejected because staged development requires visible capability honesty in the main chat experience.
- Automatically fallback from real API to mock mode. Rejected because it violates the project's no-hidden-fallback boundary.

## Decision: Treat composer actions as guarded run transitions

**Rationale**: Send, stop, retry, regenerate, and continue all affect the selected thread's current or latest run. Modeling them as guarded actions keeps duplicate sends blocked while a run is active, preserves failed user input for retry, and ensures regenerate creates a new attempt only after a completed assistant response exists.

**Alternatives considered**:

- Leave retry/regenerate as visual-only buttons for later. Rejected because Composer usability is part of this feature's success criteria.
- Allow sending new messages during active runs. Rejected because the current scope assumes one active/pending run per selected thread.
- Clear composer input on every failure. Rejected because failed-send recovery is explicitly required.

## Decision: Reuse existing thread/message/run/event domain concepts

**Rationale**: The spec intentionally builds on the current runtime adapter and assistant draft state. Reusing Thread, Message, Run, Run Event, and Assistant Draft keeps the slice aligned with M3/M4 and avoids pulling forward backend persistence changes. New state should be presentation-level unless implementation proves a durable model gap.

**Alternatives considered**:

- Introduce a separate Conversation Turn model immediately. Rejected because existing message/run/event relationships can represent the current feature.
- Add backend schema changes before UI work. Rejected because the requested work is frontend readiness ahead of richer backend events.
- Store draft/recovery state in local storage. Rejected because offline/cross-session draft persistence is outside this feature's scope.
