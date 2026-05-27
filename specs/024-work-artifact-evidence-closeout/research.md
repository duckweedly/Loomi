# Research: M17 Work Artifact Evidence Closeout

## Decision: Use `loomi-seed` as the local evidence path

**Rationale**: Existing public API can create threads/messages and start runs, but it does not expose a safe arbitrary event metadata write endpoint. Adding such an endpoint would be broader than M17. The seed command already exists for local development and tests, so an explicit `LOOMI_SEED_SCENARIO=m17-work-artifact` path can write deterministic evidence through the existing product service boundary.

**Alternatives considered**: A production POST events endpoint was rejected because it would become a general write surface. Mock-only smoke was rejected because M17 requires replayable evidence beyond M16 mock data.

## Decision: Keep artifact evidence metadata-only

**Rationale**: M17 is a closeout slice for evidence display, not artifact execution. Artifact cards should expose id, title, type, source thread/run, summary, timestamps, and a redaction marker only.

**Alternatives considered**: File previews, command execution, browser actions, downloads, and tool replay were rejected as explicit non-goals.

## Decision: Redaction marker lives on the projection

**Rationale**: The UI needs to say that unsafe metadata was removed or redacted without exposing the raw values. Adding `redactionApplied` to the projected artifact reference keeps the raw event payload out of the card.

**Alternatives considered**: Showing redacted key/value dumps was rejected because it can still imply action semantics or leak unsupported fields.
