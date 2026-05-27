# Feature Spec: M50 Memory Content View

## Goal

Let Settings > Memory open a safe content view from snapshot hits, matching the target "view memory" workflow without exposing raw memory rows.

## Requirements

- Add a safe content endpoint for `memory://{entry_id}` URIs.
- Support `overview` and `read` layers.
- Return title plus safe summary only; never raw stored content, content hash, proposal body, provider trace, local path, credential, or secret-like text.
- Let Settings > Memory snapshot hit chips open a modal with the safe content.
- Keep content reads scoped through the existing memory entry authorization boundary.

## Non-goals

- No external Nowledge/OpenViking content fetch.
- No editing approved memories from the modal.
- No bulk clear/delete-all action.
