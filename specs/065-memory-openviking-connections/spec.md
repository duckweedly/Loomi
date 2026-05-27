# Feature Spec: M65 Memory OpenViking Connections

## Goal

Bring Loomi closer to the OpenViking memory provider shape by routing `memory.connections` for `viking://...` URIs through OpenViking directory listing.

## User Story

As an agent using OpenViking-backed memory, I can ask for connections around a `viking://...` memory URI and receive safe child resource summaries instead of falling back to local memory search.

## Functional Requirements

- When OpenViking is active and configured, `memory.connections` with a `viking://...` `entry_id` must call OpenViking `/api/v1/fs/ls`.
- The result must preserve Loomi's safe tool envelope: provider, operation, target id, bounded items, count, and redaction flag.
- Items must include opaque child URI, safe title, node type, relation, and redaction flag.
- Non-OpenViking providers and non-`viking://` IDs must keep the existing behavior.

## Non-Goals

- No raw provider payloads in tool results.
- No recursive tree traversal.
- No OpenViking write or delete behavior changes.
- No Settings UI change.

## Success Criteria

- Runtime tests prove OpenViking connections use `/api/v1/fs/ls`.
- Existing Nowledge connections and local memory connections remain intact.
