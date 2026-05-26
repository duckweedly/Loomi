# Tasks: Web Search Providers

- [X] Add failing productdata tests for `web.search` catalog, argument validation, and Chat RunContext availability.
- [X] Add failing runtime tests for Tavily/Brave execution and safe summaries.
- [X] Add failing provider serialization test for the `web_search` function schema.
- [X] Implement `web.search` constants, catalog metadata, persona allowlist, and Chat filtering exception.
- [X] Implement Tavily/Brave search execution with bounded result summaries and no key/raw-body leakage.
- [X] Pass `LOOMI_TAVILY_API_KEY` and `LOOMI_BRAVE_SEARCH_API_KEY` from config into the API worker.
- [X] Add worker continuation coverage for approved `web.search`.
- [X] Update web mock catalog and RunRail safe preview labels.
- [X] Add simple Settings > Web Search key-entry menu; keep Settings > Tools read-only.
- [X] Persist web search keys and custom provider configs across API restart; bump built-in persona version for the `web.search` allowlist.
- [X] Update docs-site API/runbook/devlog/current-status.
- [X] Run full closeout validation.
