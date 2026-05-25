# Research: RunContext Pipeline Foundation

## Decision: Load RunContext from durable product data inside the worker path

**Rationale**: M9 Step 71 requires worker execution to avoid API process memory. The current M6/M8 worker queue already stores run/job metadata durably, and M7 continuation can rebuild provider context from messages and run events. Loading context at the worker boundary preserves that direction without changing the queue.

**Alternatives considered**:

- Pass rich request memory through job metadata: rejected because it keeps execution tied to API request lifetime and risks persisting unsafe payloads.
- Add a standalone context service or cache: rejected as unnecessary platform complexity for the first M9 slice.

## Decision: Keep RunContext MVP fields narrow

**Rationale**: The user requested run, thread, messages, job metadata, provider/model route, and enabled MVP tools. Persona/Skill, Memory, MCP, Sandbox, Desktop Runtime, and broad tools are later roadmap items and should not appear as fake placeholders.

**Alternatives considered**:

- Include empty Persona/Memory fields now: rejected because placeholders imply unsupported behavior and make the boundary look broader than it is.
- Persist a new context snapshot table: rejected unless implementation finds a real replay/ownership need; the current source of truth already exists.

## Decision: Use a linear stage list instead of a workflow engine

**Rationale**: M9 Step 72 needs a foundation where new middleware/stages do not require a large AgentLoop rewrite. A small ordered list of stage functions is enough for `prepare_context`, `resolve_tools`, `invoke_runtime`, and `finalize`.

**Alternatives considered**:

- Full middleware chain with arbitrary branching: rejected as broad abstraction.
- Keep direct calls in the queued runner: rejected because adding a stage would continue to edit the main execution body.

## Decision: Persist stage trace through existing run events

**Rationale**: Existing run events, SSE, Timeline, and history replay are Loomi's observable execution contract. Stage trace should use that path instead of a parallel debug store.

**Alternatives considered**:

- Store stage trace only in logs: rejected because users cannot replay logs in Timeline/debug panel.
- Add a new debug endpoint/table: rejected because existing events are sufficient for the MVP.

## Decision: Treat stage metadata as safe summaries only

**Rationale**: Context and runtime stages can touch provider routes, messages, tool summaries, and errors. Persisted metadata must explain execution without exposing secrets, raw provider payloads, raw tool results, file contents, shell output, or hidden local state.

**Alternatives considered**:

- Persist complete context snapshots for debugging: rejected due to privacy and secret-leak risk.
- Redact only at UI time: rejected because SSE/history and database storage must already be safe.

## Decision: Preserve M7 continuation through the existing runtime boundary

**Rationale**: `specs/012-tool-result-model-continuation` already defines provider-neutral continuation and single-tool loop limits. M9 should call that behavior from `invoke_runtime`, not duplicate it.

**Alternatives considered**:

- Rebuild continuation inside the pipeline stage: rejected as duplication and risk to M7 behavior.
- Delay continuation support until a later M9 slice: rejected because current runs must not regress.
