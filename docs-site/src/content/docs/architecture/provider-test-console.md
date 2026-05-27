---
title: Provider Test Console
description: M6.5 provider readiness UX for local real model testing.
---

M6.5 turns Settings > Providers into a Provider Test Console for the real local testing path. The console shows backend-configured providers, lets the user trigger a safe connection check, and keeps browser-local draft fields separate from real model execution.

## Source of truth

Configured providers come from the backend provider readiness/configuration surface. They affect real model calls in `real_api` and model gateway flows.

The local draft section is browser-session-only. Base URL, model, and API key draft fields are notes for the current browser session. They are not saved, are not sent as backend provider configuration, and do not change real model calls.

## Provider row contract

Each configured provider row should expose the safest available subset of:

- provider id
- family
- model
- base URL
- readiness status
- sanitized message

Provider messages are display data, not instructions. API keys, bearer tokens, and secret-like values must not be rendered.

## Test connection

The UI presents Test connection as a per-provider row action. If the backend only supports aggregate readiness refresh, the frontend maps the aggregate result back to provider rows without introducing new secret storage or provider-management APIs.

The visible states are:

- not checked
- checking
- configured
- reachable
- completion-ok
- completion-failed
- failed
- not configured

M72 separates static configuration from live completion health. `GET /v1/model-providers` only reports local configuration readiness and must not call the provider. `POST /v1/model-providers/check` runs a minimal completion smoke and may return `check_code` values such as `completion-failed-503`. Settings shows the check code on the provider card so an upstream 503 is not mistaken for an available model.

## Chat and Composer relationship

Provider readiness gates real provider-dependent modes only:

| Capability | Provider readiness required |
| --- | --- |
| mock | no |
| real_api | yes |
| model_gateway | yes |

When a real mode has no available provider, Chat/Composer shows provider unavailable guidance and an Open Settings > Providers CTA. It must not imply the model is generating.

## Out of scope

M6.5 does not add final provider secret storage, persistent API key management, approval, tool calls, or tool execution protocol.
