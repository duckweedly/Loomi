---
title: M5.5 Settings Placeholder Architecture
description: Temporary settings surface boundaries for session-local controls, read-only runtime state, and safe placeholders.
---

M5.5 adds an in-app Settings surface to the existing Loomi web shell. The surface is a desktop-style two-column view with Loomi-owned categories, grouped cards, and clear working versus placeholder status.

## Boundary

Settings is frontend state for this milestone. Working controls are limited to current-session local behavior:

- default workspace mode for future local thread creation
- mock runtime scenario for future mock sends
- current-session interface language, defaulting to Chinese with an English switch in Settings

Runtime, data source, stream, selected thread/run, and provider capability rows are read-only status views. They expose only frontend-safe state already available to the shell.

Providers also has a current-session draft form for OpenAI-compatible gateway notes: Base URL, model ID, and masked API key presence. The key string is not retained in state; the UI keeps only whether a key was entered.

Provider secrets stay outside durable settings in M5.5. The settings UI must never render API keys, Authorization headers, raw provider payloads, raw provider errors, or secret-bearing URL fragments.

## State flow

`useWorkspaceShellState` owns the Settings surface state: open/closed, selected category, session-local default workspace mode, current language, and provider draft settings. `useWorkspaceState` still owns thread/message/run behavior and receives the default workspace mode when creating future local threads.

The mock runtime scenario control reuses the existing `selectRuntimeScript` path. Changing it does not mutate an active run; it only affects the next mock send.

## Categories

The catalog groups Settings into Workspace, Agent Core, and Management sections. General is the primary working category. Providers is mixed: redacted capability status plus session-local draft inputs. About is mixed status/preview, and the remaining future platform categories are mock placeholders.

## Placeholder areas

Future platform areas such as durable provider management, connectors, plugins, skills, MCP, notebook, memory, activity recording, context, safety, tools, routes, about metadata, and advanced diagnostics are represented as mock or preview panels only. Placeholder controls are disabled and do not execute tools, call providers or connectors, write files, write backend state, or persist values as real settings.
