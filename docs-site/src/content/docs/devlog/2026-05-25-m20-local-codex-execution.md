---
title: M20 Local Codex Execution Bridge
description: Enabled Local Codex now routes through the model gateway provider path.
---

Completed:

- Added `local_codex` runtime provider implementation behind explicit detect/enable.
- Chose auth.json direct execution over CLI execution.
- Returned enabled Local Codex as `status=available` and `execution_state=supported` when the bridge can execute.
- Registered Local Codex with the existing Gateway instead of creating a separate chat path.
- Covered fixture Chat HTTP smoke: create thread/message/run, process worker, persist assistant message, and record gateway events.
- Added Local Codex-specific Chat warning copy for unsupported and unavailable states.
- Kept token/key/path canaries out of API responses, run events, assistant metadata, and frontend mapping tests.

Known limitation:

- M20 is a fixture-backed Chat bridge candidate. Real local Codex Chat depends on the user's local auth token being usable against the configured compatible endpoint and must be verified with a manual smoke on each machine. The bridge fails explicitly instead of fabricating output when auth or endpoint execution is unavailable.
