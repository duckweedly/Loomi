# Plan: M61 Memory External Provider Write Adapters

1. Add external write/edit/delete helper functions beside the read adapters.
2. Wire `memory.write`, `memory.edit`, and `memory.forget` to external providers when the selected provider and URI match.
3. Preserve local proposal/tombstone behavior for local memory.
4. Extend local `httptest` adapter tests to cover write/edit/delete.
5. Update architecture/API/runbook/devlog documentation.
