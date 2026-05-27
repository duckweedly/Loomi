# Contract: Settings Provider UI

## Surface

Settings > Providers shows a "Local provider autodetect" section below configured provider checks.

## Required rows

- Local Claude Code
- Local Codex

Each row shows:

- display name
- detected/not detected status
- auth mode label
- safe model candidate labels
- copy that explicit opt-in is required before use
- copy that no secrets are shown

## Forbidden UI behavior

- Do not show token/key/path content.
- Do not auto-select or enable local providers.
- Do not trigger model-gateway calls.
- Do not expose a Connect/Use button in this slice.

## Backend unavailable state

If local detection cannot load, show a stable unavailable/error message and keep configured provider UI usable.
