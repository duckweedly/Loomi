---
title: M27 Browser Automation Foundation
description: Approval-gated public HTTP browser-session tool architecture.
---

M27 adds the first browser automation slice as three builtin tools: `browser.open`, `browser.snapshot`, and `browser.click_link`.

The tools reuse the existing `ToolCatalog -> RunContext -> approval -> ToolBroker -> worker continuation` path. The runtime is an HTTP-backed, run-scoped page session model. It does not introduce Chrome profile control, cookies, JavaScript rendering, forms, screenshots, downloads, crawler behavior, artifact capture, or authenticated browsing.

## Boundaries

Browser tools are:

- builtin
- Work mode only
- approval required
- read-only from Loomi's perspective
- medium risk
- public HTTP(S) only
- stateful within one run-scoped browser session

Chat mode filters them out with workspace, sandbox, LSP, and web tools. Production execution rejects credentialed URLs, non-HTTP schemes, localhost, loopback, link-local, private, multicast, and unspecified hosts before dialing. Redirect and clicked link targets pass the same validation before response bodies are read.

## Execution

`BrowserToolExecutor` keeps bounded in-memory sessions keyed by a generated `session_id`. `browser.open` fetches one public HTTP(S) page and stores only safe derived state: final URL, title, text excerpt, bounded links, byte count, truncation, and content type. Raw HTML is not stored in the session result.

`browser.snapshot` returns the current safe page state for an existing session. `browser.click_link` resolves one approved link index from that session, validates the target, navigates once, and updates the safe session state.

Gateway continuation now allows supported browser tools after a previous tool terminal result, while retaining duplicate tool-call and bounded loop limits.

## Visibility

Settings > Tools shows browser tools as browser-scoped, approval-required, medium risk, public HTTP only, and stateful. RunRail labels browser lifecycle rows separately from workspace, sandbox, LSP, web fetch, MCP, and runtime tools.
