# Contract: Frontend Persona Summary

## Minimal UI Surface

The implementation may choose one of these MVP UI paths:

1. A minimal persona selector on the existing run/thread creation surface.
2. A read-only resolved persona display when default persona inheritance is the only implemented interaction.

The chosen path must still support browser smoke proving a run used a persona.

## Safe Summary Fields

Timeline/debug may display:

- persona name
- persona version
- description
- resolved source: run, thread, or default
- model route label
- reasoning mode
- budget summary
- allowed tool names or count

Timeline/debug must not display:

- raw system prompt
- provider credentials
- raw provider request/response bodies
- raw tool result payloads
- file contents
- shell output
- browser/desktop captured state
- hidden local state

## Event and Replay Expectations

Persona summary may appear on `prepare_context` completed metadata or an equivalent safe run-context debug row.

Rules:

- Live SSE and history replay must produce the same visible persona summary.
- A run without explicit selection should show that the default persona was resolved.
- A run with thread/run selection should show the selected persona name/version.
- Missing persona summary must not crash Timeline/debug; the UI may omit the row or show a redacted unavailable state.

## Browser Smoke Acceptance

1. Open the local web app in real API mode.
2. Select a persona or confirm the default persona display.
3. Create a run.
4. Open Timeline/debug.
5. Confirm name/version/model route/reasoning/budget/tool summary is visible.
6. Confirm raw persona prompt text is absent.
