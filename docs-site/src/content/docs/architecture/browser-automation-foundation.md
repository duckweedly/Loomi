---
title: M27 Browser Automation Foundation
description: Approval-gated public HTTP browser-session tool architecture.
---

M27 adds the first browser automation slice as builtin tools: `browser.open`, `browser.snapshot`, `browser.click_link`, `browser.screenshot`, `browser.type`, and `browser.press`.

The tools reuse the existing `ToolCatalog -> RunContext -> approval -> ToolBroker -> worker continuation` path. The runtime is an HTTP-backed, run-scoped page session model. It does not introduce Chrome profile control, cookies, JavaScript rendering, binary screenshots, downloads, crawler behavior, artifact capture, or authenticated browsing.

## Boundaries

Browser tools are:

- builtin
- Work mode only
- approval required
- medium risk
- public HTTP(S) only
- stateful within one run-scoped browser session

Chat mode filters them out with workspace, sandbox, LSP, and web tools. Production execution rejects credentialed URLs, non-HTTP schemes, localhost, loopback, link-local, private, multicast, and unspecified hosts before dialing. Redirect and clicked link targets pass the same validation before response bodies are read.

## Execution

`BrowserToolExecutor` keeps bounded in-memory sessions keyed by a generated `session_id`. `browser.open` fetches one public HTTP(S) page and stores only safe derived state: final URL, title, text excerpt, bounded links, byte count, truncation, and content type. Raw HTML is not stored in the session result.

`browser.snapshot` returns the current safe page state for an existing session. `browser.click_link` resolves one approved link index from that session, validates the target, navigates once, and updates the safe session state.

`browser.screenshot` returns a bounded text screenshot summary derived from the safe page state. It is not a PNG/JPEG capture and does not render JavaScript.

`browser.type` records bounded text against an input target discovered from the current safe HTML snapshot. `browser.press` records a bounded key press such as `Enter`, `Escape`, `Tab`, or arrow keys. These operations update the run-scoped browser session state only; they do not execute JavaScript or submit credentials to authenticated sites.

Gateway continuation now allows supported browser tools after a previous tool terminal result, while retaining duplicate tool-call and bounded loop limits.

## Visibility

Settings > Tools shows browser tools as browser-scoped, approval-required, medium risk, public HTTP only, and stateful. RunRail labels browser lifecycle rows separately from workspace, sandbox, LSP, web fetch, MCP, and runtime tools.
