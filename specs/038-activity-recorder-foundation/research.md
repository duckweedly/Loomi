# Research: M30 Activity Recorder Foundation

## Decision: Start with explicit summary events, not automatic desktop capture

**Rationale**: The constitution requires opt-in and deletion/cleanup paths for activity recording. A safe foundation can prove the data contract, redaction, user controls, and UI without introducing OS-level capture, screenshots, keystroke logging, or clipboard access.

**Alternatives considered**:

- Automatic desktop capture: rejected for the first slice because it needs platform permissions, stronger deletion semantics, and higher privacy review.
- Browser profile capture: rejected because authenticated browsing/cookies/raw HTML are explicitly outside current browser automation boundaries.
- Reusing memory entries: rejected because activity summaries have different opt-in, cleanup, and visibility semantics from user-approved memory.

## Decision: Store redacted bounded summaries in productdata memory service

**Rationale**: M21-M29 foundation slices already use productdata memory-backed records for local runtime state. This keeps M30 testable without migrations while preserving a clean seam for later durable storage.

**Alternatives considered**:

- PostgreSQL migration now: deferred until the data model stabilizes.
- Run events only: rejected because Activity Recorder needs user-visible status, list, and cleanup independent of a single run.

## Decision: Settings Activity Recorder becomes the first real control surface

**Rationale**: The Settings IA already includes Activity Recorder as a placeholder. Moving it to a real panel provides visible opt-in, list, empty/error states, and cleanup controls without inventing a new navigation surface.

**Alternatives considered**:

- RunRail-only display: rejected because recorder state is user setting/control, not one run lifecycle.
- Hidden API-only foundation: rejected because activity capture must be user-visible.
