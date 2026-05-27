---
title: M20 Local Codex Execution Bridge
description: Explicit Local Codex opt-in connected to the existing model gateway.
---

M20 turns enabled Local Codex from an unsupported route candidate into an executable local provider.

The execution bridge uses the existing runtime `Provider` interface. It does not add a second chat path. Chat still creates a normal model gateway run, the worker claims the queued job, the Gateway selects `local_codex`, and timeline/SSE projection receives the same `model_request_started`, `model_output_delta`, `model_output_completed`, `run_completed`, or provider failure events used by other model providers.

## Chosen Bridge

M20 uses the auth.json direct bridge.

The rejected alternative was invoking the installed `codex` CLI. CLI execution was not selected because this repo does not have a proven non-interactive Codex CLI contract that prevents prompt hangs, uncontrolled file writes, or stdout/stderr leakage of prompts and tokens.

## Safety Boundary

Local Codex remains opt-in only:

1. Settings > Providers runs explicit local provider detection.
2. The user explicitly enables Local Codex for the current session.
3. Enable validates that a credential snapshot and execution bridge are available.
4. The server registers a `local_codex` provider with the existing Gateway.
5. Later Chat sends use the existing worker/Gateway path.

`GET /v1/model-providers` does not scan local auth files. It only reads process-local enablement state.

## Redaction

The provider capability exposes only safe routing fields:

- `id`
- `family`
- `model`
- `status`
- `local_provider`
- `session_local`
- `credential_reference`
- `execution_state`

Access tokens, refresh tokens, API keys, Authorization headers, auth file paths, private home paths, and raw auth JSON must not enter API responses, run events, assistant metadata, frontend state, docs, or logs.

## Current Limitation

The automated proof is fixture-backed using temporary `CODEX_HOME` and a local OpenAI-compatible test server. Real local OAuth compatibility depends on whether the local Codex auth token can call the configured OpenAI-compatible endpoint without refresh.
