---
title: M5.5 Settings Placeholder Architecture
description: Temporary settings surface boundaries for session-local controls, real management panels, and hidden placeholders.
---

M5.5 adds an in-app Settings surface to the existing Loomi web shell. The surface is a desktop-style two-column view with Loomi-owned categories, grouped cards, and only real user-facing controls.

## Boundary

Settings is frontend state for this milestone. Working controls are limited to current-session local behavior:

- default workspace mode for future local thread creation
- current-session interface language, defaulting to Chinese with an English switch in Settings
- current-session light/dark appearance

Runtime, data source, stream, selected thread/run, and provider capability diagnostics are not part of General settings. Runtime state belongs in Chat, run details, or future debug/timeline surfaces where it explains execution instead of pretending to be a user preference.

Providers uses a management-list layout for redacted model providers: search, All/Enabled/Local/Cloud filters, read-only local provider cards, safe connection-test actions, and an Add provider dialog. The dialog submits the current OpenAI-compatible Base URL, model, and API key to the local API save path; the key string is cleared after save and is never echoed back into the interface.

Provider secrets stay outside durable settings in M5.5. The settings UI must never render API keys, Authorization headers, raw provider payloads, raw provider errors, or secret-bearing URL fragments.

## State flow

`useWorkspaceShellState` owns the Settings surface state: open/closed, selected category, session-local default workspace mode, current theme, current language, and provider draft settings. `useWorkspaceState` still owns thread/message/run behavior and receives the default workspace mode when creating future local threads.

## Categories

The catalog groups Settings into Workspace, Agent Core, and Management sections. General is the primary working category. Providers is mixed: redacted capability status plus local provider save inputs. Skill is read-only: it lists Loomi personas and installed local `SKILL.md` manifests from Codex, Claude Code, project, and plugin roots. About is mixed status/preview. Placeholder-only categories are not exposed in the navigation.

## Placeholder areas

Plugin management is intentionally not exposed while it has no real backend slice. Settings should not show generic mock panels; each visible category needs a real read/control surface.

Activity Recorder and Context are intentionally not exposed as standalone Settings navigation items while they are placeholder-only. Activity Recorder returns to Settings only when the opt-in, audit, redaction, list, and cleanup controls are implemented as a real surface. RunContext remains an execution trace/debug concern rather than a user-facing settings page.

Safety, Routes, and Advanced are also intentionally not exposed while they are placeholder-only. Existing safety signals live in the concrete surfaces that use them: tool approval in Chat, tool policy in Settings > Tools, memory redaction/audit in Settings > Memory, provider state in Settings > Providers, and MCP state in Settings > MCP.
