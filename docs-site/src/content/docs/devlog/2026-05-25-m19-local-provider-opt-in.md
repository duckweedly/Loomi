---
title: M19 Local Provider Opt-in Bridge
description: Session-local Local Codex opt-in route candidate and unsupported execution guard.
---

## Completed

- Added explicit local provider enable/disable API routes.
- Kept detected-only local providers out of `GET /v1/model-providers`.
- Added session-local Local Codex route candidate metadata: local provider, session-local, credential redacted, execution unsupported.
- Updated Settings > Providers to show Enable for this session and Disable for this session.
- Kept Chat blocked when Local Codex is enabled but execution is unsupported.
- Preserved existing OpenAI-compatible provider save/check path.

## Chat status

M19 cannot actually Chat through Local Codex. The remaining blocker is a dedicated Local Codex execution bridge with fixture-backed provider tests and full request/response/event redaction.

## Safety notes

No CLI execution, keychain access, OAuth refresh, external validation, token persistence, sandbox/browser/filesystem/shell/workspace tools, or secret-bearing response fields were added.
