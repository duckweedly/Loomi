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

`artifact.create_text` result summaries now also expose a first-class `artifacts[]` resource reference. The first reference uses the persisted artifact id as `key`, carries safe presentation metadata (`title`, `filename`, `mime_type`, `display`), and keeps the legacy flat `artifact_id/title/text_excerpt` fields for existing consumers. The provider prompt tells Work-mode runs to create reports, articles, Markdown, and other saveable documents through `artifact.create_text`, then cite the returned resource as `[title](artifact:<key>)`.

## Visibility

Settings > Tools shows artifact tools as artifact-scoped, approval-required, medium risk, and non-executable. RunRail labels artifact lifecycle rows separately from workspace, sandbox, LSP, web fetch, browser, MCP, and runtime tools.

The frontend also derives a lightweight `PreviewArtifact` projection from safe tool-call result summaries and run events. Completed artifact/document tools render as compact document resource cards in the chat timeline. Selecting a card opens the right Preview drawer and renders Markdown/text excerpts when the safe summary contains preview content.

Assistant messages can also contain `artifact:<key>` Markdown links. The chat timeline strips the protocol link from prose, renders a compact document card for the resource, and leaves the Preview drawer bounded to the safe artifact projection instead of displaying raw protocol text.

The chat transcript groups ordered model deltas, tool events, and continuation deltas into stable turns. When terminal tool events arrive before the next model delta, Loomi keeps a small waiting-for-model state in the transcript so the conversation does not visually collapse between tool completion and final generation.

This preview projection is intentionally display-only. It does not grant filesystem access, download artifacts, execute HTML, or expose raw unbounded tool output. Approval-required tools still use the normal tool-confirmation UI until the tool reaches a terminal state.
