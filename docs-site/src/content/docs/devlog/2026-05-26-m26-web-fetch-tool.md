---
title: 2026-05-26 M26 Web Fetch Tool Foundation
description: Implementation notes and validation for M26 web.fetch.
---

## Completed

- Added Spec Kit feature `specs/034-web-fetch-tool-foundation/`.
- Added builtin `web.fetch` catalog identity, default persona allowlist, Chat/Work availability, and safe tool-call metadata grouping.
- Added `WebToolExecutor` with HTTP(S)-only URL validation, credential rejection, private/local network denial, DNS checks, redirect validation, timeout/byte bounds, text-like content handling, title extraction, and bounded excerpts.
- Routed `web.fetch` through ToolBroker, worker approved-tool resume, provider continuation, and HTTP smoke coverage.
- Updated Settings > Tools, RunRail labels, mock catalog, and runtime scripts for visible web lifecycle metadata.

## 2026-05-26 Web Search Add-on

- Added builtin `web.search` with Brave Search and Tavily provider execution.
- `web.search` and `web.fetch` are read-only, auto-approved, and available to Chat and Work runs when the persona allows them.
- Provider requests now include a `web_search` function schema for OpenAI-compatible/local Codex paths and map it back to Loomi's internal `web.search` tool name.
- Added a simple Settings > Web Search menu for entering Tavily/Brave keys; Settings > Tools stays read-only.
- Web search keys and the local custom OpenAI-compatible provider are persisted through product data after migration `000012`, with key presence exposed as booleans only.
- Bumped the built-in default persona to `2026-05-26.1` so existing local databases can sync a persona version that allows `web.search`.
- The API worker passes `LOOMI_TAVILY_API_KEY` and `LOOMI_BRAVE_SEARCH_API_KEY` into the WebToolExecutor.
- Search result summaries include bounded `title`, `url`, and `snippet` fields only; API keys, raw provider responses, headers, cookies, and raw page content are excluded.
- Fixed OpenAI-compatible streaming tool-call parsing so split `tool_calls[].function.arguments` chunks are accumulated before recording `web.search`; this prevents false `tool_call_rejected` failures for real search prompts.
- Continuation requests stop advertising tools once the bounded tool-loop limit has been reached, giving the model a final chance to answer from gathered search results instead of failing only because it requested another search.

## 2026-05-26 Chat Tool Policy Alignment

- Provider requests now include a system tool policy: do not call tools for greetings or stable knowledge, use `web_search` for current information, and use `web_fetch` for public URLs or source follow-up.
- OpenAI-compatible and Local Codex request builders pass that policy into model instructions instead of relying on model guesswork.
- Chat RunContext keeps both public web tools while continuing to filter workspace, sandbox, LSP, browser, artifact, todo, and agent tools.
- OpenAI-compatible, Anthropic, and Gemini provider tool-call parsing now normalize common aliases such as `web.fetch` and `fetch` back to the stable internal `web.fetch` tool name before gateway validation.
- Auto-approved public web tools now execute inside the same leased worker turn and immediately continue the provider response, instead of enqueueing a second active job for the same run.
- Rejected tool calls now record safe tool name, argument summary, and redacted error metadata so real browser/API smoke can identify the failing boundary without leaking provider payloads or secrets.

## Validation

Focused validation during implementation:

```bash
go test ./internal/productdata
go test ./internal/runtime
go test ./internal/config
go test ./internal/httpapi -run 'TestM26WebFetch|TestM25LSPReadonlyApproveExecuteFinalSmoke'
bun test --cwd web SettingsView.tools RunRail.runtime runtimeScripts
```

Full closeout commands should also run before marking M26 complete:

```bash
go test ./...
bun test --cwd web
bun run --cwd web build
bun run --cwd docs-site build
git diff --check
```

## Non-goals

No JavaScript rendering, cookies, authenticated fetch, deep crawler, social media search, activity recorder, plugin marketplace, channels, heartbeat, or multi-agent orchestration were added.
