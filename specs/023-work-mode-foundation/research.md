# Research: M16 Work Mode Foundation

## Decision: Derive Work Plan View from existing messages and run events

**Rationale**: The current roadmap already has thread/message/run/event foundations. Projecting a read-only work surface from these structures produces the thin vertical slice without inventing a new task system.

**Alternatives considered**: A new task table or worker queue was rejected because M16 only needs display/projection and the constitution says core flow should precede platform complexity.

## Decision: Artifact references are safe metadata only

**Rationale**: The requested first version only needs title, type, source run/thread, summary, and timestamps. Metadata projection can redact secret-looking values before UI render.

**Alternatives considered**: File previews, shell/browser tools, and artifact execution were rejected as explicit non-goals and would require permission/sandbox design.

## Decision: Put Work Plan View in the main Work thread area

**Rationale**: The user accepts main area or right panel. The main area is already selected-thread scoped and can show clear loading/error/empty states beside message history.

**Alternatives considered**: Right drawer-only rendering was rejected because the existing drawer is optional/collapsible and would make the first Work mode slice easier to miss.

## Decision: No new backend API for M16

**Rationale**: Existing event metadata and mock/real run event mapping are enough for a minimal projection. A documented UI contract is sufficient.

**Alternatives considered**: A `/work-plan` endpoint was rejected until the frontend projection proves a concrete server-side gap.
