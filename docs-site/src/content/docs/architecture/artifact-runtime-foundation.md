---
title: M28 Artifact Runtime Foundation
description: Approval-gated non-executable text artifact tool architecture.
---

M28 adds the first artifact runtime slice as three builtin tools: `artifact.create_text`, `artifact.read`, and `artifact.list`.

The tools reuse the existing `ToolCatalog -> RunContext -> approval -> ToolBroker -> worker continuation` path. The runtime is storage-only and non-executable. It does not render, download, run, compile, export to the filesystem, open a browser, call shell, or call network.

## Boundaries

Artifact tools are:

- builtin
- Work mode only
- approval required
- medium risk
- non-executable
- bounded UTF-8 text only

Chat mode filters them out with workspace, sandbox, LSP, web, and browser tools. `artifact.create_text` stores one bounded text artifact through `ArtifactService`, backed by both in-memory service and PostgreSQL `artifacts` in the real API path. `artifact.read` returns one bounded excerpt. `artifact.list` returns bounded safe summaries without raw content.

## Execution

`ArtifactToolExecutor` depends on `productdata.ArtifactService`. Worker approved-tool resume injects the configured productdata service or repository, then records tool success and continues the provider with a safe result summary.

Run events persist title/type/size/excerpt/truncation/source ids only. They do not include raw unbounded content, executable payloads, downloads, local paths, credentials, or raw provider payloads.

## Visibility

Settings > Tools shows artifact tools as artifact-scoped, approval-required, medium risk, and non-executable. RunRail labels artifact lifecycle rows separately from workspace, sandbox, LSP, web fetch, browser, MCP, and runtime tools.
