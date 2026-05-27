---
title: Local Web Tools Validation
description: Commands for validating the web.fetch and web.search runtime slices locally.
---

## Focused Validation

```bash
go test ./internal/productdata -run 'TestToolCatalogIncludesWebFetchTool|TestToolCatalogIncludesWebSearchTool|TestWebSearchIsAvailableInChatRunContext|TestValidateWebFetchToolCallArguments|TestValidateWebSearchToolCallArguments'
go test ./internal/runtime -run 'TestWebFetch|TestWebSearch|TestToolBrokerExecutesWebFetchThroughOneEntrypoint|TestWorkerExecutesApprovedWebFetchAndContinuesModel|TestWorkerExecutesApprovedWebSearchAndContinuesModel|TestHTTPProviderSendsEnabledWebSearchToolSchema'
go test ./internal/httpapi -run 'TestM26WebFetchAutoExecuteFinalSmoke'
bun test --cwd web SettingsView.tools RunRail.runtime runtimeScripts
```

## Search Provider Environment

```bash
export LOOMI_TAVILY_API_KEY=tvly-...
export LOOMI_BRAVE_SEARCH_API_KEY=...
```

At least one key is required for `web.search`. If both are present, the runtime defaults to Tavily unless the model asks for `provider: "brave"`.

## Full Closeout

```bash
go test ./...
bun test --cwd web
bun run --cwd web build
bun run --cwd docs-site build
git diff --check
```

## Manual Smoke

1. Start the API and web app.
2. Open Settings > Web Search and save either Tavily or Brave Search key. One key is enough; two keys enable both providers. Saved keys are not echoed back.
3. Open Settings > Tools and confirm `web.fetch` and `web.search` remain in the read-only catalog. Web search configuration lives in the Web Search menu.
4. Ask Chat for a current/news/search question and confirm the provider requests `web.search`, it auto-approves as a read-only search, execution succeeds, and the final answer uses returned result snippets.
4. Ask Chat to analyze a public URL and confirm the provider can request `web.fetch` without a manual approval click.
5. Confirm the UI shows URL/status/truncation/search metadata without raw response bodies, provider API keys, cookies, credentials, Set-Cookie values, or local paths.

Web tools do not support JavaScript rendering, cookies, authenticated fetch, deep crawling, social media search, artifact runtime, channels, desktop activity recording, or multi-agent orchestration.
