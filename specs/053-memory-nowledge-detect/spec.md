# Feature Spec: M53 Nowledge Local Detect

## Goal

Add a Settings > Memory action that detects a local Nowledge instance and fills its base URL.

## Requirements

- Add `GET /v1/memory/provider/nowledge/detect`.
- Probe `http://127.0.0.1:14242/health` with a short timeout.
- Return only detected state, base URL, and safe message.
- Add a Settings > Memory detect button for Nowledge configuration.

## Non-goals

- No remote network scan.
- No API key discovery.
- No automatic provider activation beyond filling the base URL.
