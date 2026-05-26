# Implementation Plan: Web Search Providers

## Scope

Implement the smallest real search slice by extending the existing M26 web tool boundary rather than creating a new browser/crawler subsystem.

## Architecture

- `internal/productdata`: add `web.search` identity, argument validation, catalog metadata, default persona allowlist, and Chat RunContext allowance.
- `internal/runtime`: extend `WebToolExecutor` with Tavily/Brave HTTP clients, provider-safe result summaries, provider tool schema, and continuation support.
- `cmd/loomi-api`: pass search provider keys from config into the worker WebToolExecutor.
- `web`: render catalog and RunRail search lifecycle without raw metadata leakage.
- `docs-site`: document API, env vars, validation, and known limits.

## Safety

Search is public-network read-only but still approval-gated. API keys stay in env/config only; events and UI receive only safe summaries. Raw provider responses are parsed into bounded title/url/snippet items and discarded.

## Validation

Run focused Go tests for productdata/runtime/config, focused web tests for catalog/RunRail, then full Go/web/docs validation before completion.
