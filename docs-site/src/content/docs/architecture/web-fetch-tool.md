---
title: M26 Web Fetch Tool Foundation
description: Approval-gated public HTTP(S) fetch tool architecture.
---

M26 adds the first web runtime slice as one builtin tool: `web.fetch`.

The tool reuses the existing `ToolCatalog -> RunContext -> approval -> ToolBroker -> worker continuation` path. It does not introduce a browser runtime, JavaScript rendering, cookies, crawler behavior, search provider integration, authenticated fetch, or artifact capture.

## Boundaries

`web.fetch` is:

- builtin
- Work mode only
- approval required
- read-only
- medium risk
- public HTTP(S) only

Chat mode filters it out with workspace, sandbox, and LSP tools. Production execution rejects credentialed URLs, non-HTTP schemes, localhost, loopback, link-local, private, multicast, and unspecified hosts before dialing. Redirect targets pass the same validation before response bodies are read.

## Execution

`WebToolExecutor` uses Go stdlib HTTP execution with bounded timeout, redirect count, and response bytes. Result summaries include safe request/final URL, status, content type, optional HTML title, text excerpt, bytes read, byte limit, truncation flag, and redaction flag.

Events store summaries, not full response bodies. They do not include cookies, authorization values, raw headers, Set-Cookie values, local paths, provider raw payloads, or secret-looking content.

## Visibility

Settings > Tools shows `web.fetch` as web-scoped, read-only, approval-required, and public HTTP only. RunRail labels web fetch lifecycle rows separately from workspace, sandbox, MCP, LSP, and runtime tools.
