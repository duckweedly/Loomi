---
title: M18.5 Local Provider Autodetect
description: Detection-only boundary for local Claude Code and Codex provider capability.
---

M18.5 adds a conservative local provider autodetect slice. It helps Settings answer one question: does the local machine appear to have Claude Code or Codex auth material that could support a future local provider?

This slice does not use those credentials for model calls.

## Boundary

The detector lives in the backend provider/runtime boundary and returns safe capability summaries only:

- provider id
- display name
- provider kind
- auth mode
- status
- model candidate labels
- categorical source
- redaction flag
- stable safe message

It does not return file paths, raw auth values, Authorization headers, token field values, provider request bodies, or account metadata.

## Detector inputs

Detector tests pass explicit fixture inputs:

- `HomeDir`
- `CodexHome`
- `ClaudeConfigDir`
- environment key/value map

Production endpoint construction may use process environment and user home for read-only shape detection, but tests do not depend on real HOME.

## Claude Code detection

The first slice recognizes these local fixture shapes:

- `.claude.json` with `primaryApiKey` presence
- `.claude/settings.json` with `env.ANTHROPIC_AUTH_TOKEN`
- `.claude/settings.json` with safe `ANTHROPIC_MODEL`
- `.claude/.credentials.json` with OAuth token presence

`apiKeyHelper` is explicitly unsupported. Loomi records a stable unsupported status and does not execute it.

## Codex detection

The first slice recognizes:

- `CODEX_API_KEY` or `OPENAI_API_KEY` in injected env
- `CODEX_HOME/auth.json`
- `HOME/.codex/auth.json`
- auth-file API key presence
- `auth_mode` plus `tokens.access_token` OAuth token presence

`CODEX_API_KEY` env wins over auth-file detection. Loomi does not refresh OAuth tokens and does not write auth files.

## Settings behavior

Settings > Providers shows local provider autodetect behind an explicit Detect action. Opening the workspace or Settings page does not automatically scan local provider config shape.

When detection succeeds, the section shows Local Claude Code and Local Codex in read-only cards that say:

- detected or not detected
- explicit opt-in is required before use
- no secrets are shown

Detected local providers are not appended to configured model providers and do not change the model-gateway route.

If the endpoint is unavailable or fails, Settings shows the endpoint error only and does not render fallback "not detected" provider cards.

## Non-goals

M18.5 does not install or execute Claude Code, Codex, rtk, or any CLI. It does not read keychain data, refresh OAuth tokens, call external provider endpoints, enable providers, persist tokens, or start model calls.
