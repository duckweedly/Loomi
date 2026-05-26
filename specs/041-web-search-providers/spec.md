# Feature Specification: Web Search Providers

**Status**: Candidate
**Date**: 2026-05-26

## User Goal

Chat should be able to answer current/news/search questions by using an approval-gated web search tool backed by Tavily or Brave Search, instead of saying it cannot access real-time information.

## Functional Requirements

- Add builtin `web.search` to the tool catalog with safe metadata for provider, query, limit, and timeout.
- Allow `web.search` in Chat and Work RunContext when the selected persona allowlist includes it.
- Keep `web.fetch` Work-only.
- Execute `web.search` through the existing ToolBroker, approval, worker resume, continuation, and run-event path.
- Support `LOOMI_TAVILY_API_KEY` and `LOOMI_BRAVE_SEARCH_API_KEY`.
- Return only bounded safe result summaries: provider, query, result count, and item title/url/snippet.
- Never expose provider API keys, request headers, raw provider bodies, cookies, raw page content, or local paths in events/UI/docs examples.
- Send provider tool schema as `web_search` and map it back to internal `web.search`.

## Non-goals

- No JavaScript rendering, browser profile, cookies, authenticated fetch, crawling, social media search, image search, marketplace/plugin install, activity recorder, or multi-agent expansion.

## Success Criteria

- Product data tests prove catalog, validation, and Chat RunContext availability.
- Runtime tests prove Tavily and Brave execution, no key leakage, provider tool schema, and worker continuation.
- Web tests prove Settings > Tools and RunRail render search state safely.
- Docs describe env vars, API shape, result summary, and limits.
