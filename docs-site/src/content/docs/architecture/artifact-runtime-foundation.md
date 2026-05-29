---
title: M28 Artifact Runtime Foundation
description: Approval-gated text and visual artifact tool architecture.
---

M28 adds the first artifact runtime slice as four builtin tools: `artifact.create_text`, `artifact.create_visual`, `artifact.read`, and `artifact.list`.

The tools reuse the existing `ToolCatalog -> RunContext -> approval -> ToolBroker -> worker continuation` path. Text artifacts remain storage-only and non-executable. Visual artifacts are bounded SVG/HTML resources rendered only inside Loomi's sandboxed Preview frame. The runtime does not download, compile, export to the filesystem, open a browser, call shell, or call network.

## Boundaries

Artifact tools are:

- builtin
- Work mode only
- approval required
- medium risk
- non-executable for text artifacts
- renderable only through the sandboxed Preview frame for visual artifacts
- bounded UTF-8 text, SVG, or HTML only

Chat mode filters them out with workspace, sandbox, LSP, web, and browser tools. `artifact.create_text` stores one bounded text artifact through `ArtifactService`, backed by both in-memory service and PostgreSQL `artifacts` in the real API path. `artifact.create_visual` stores one bounded `visual` artifact and accepts only `image/svg+xml` or `text/html`. `artifact.read` returns one bounded excerpt and may include bounded content for the requested artifact. `artifact.list` returns bounded safe summaries without raw content.

## Execution

`ArtifactToolExecutor` depends on `productdata.ArtifactService`. Worker approved-tool resume injects the configured productdata service or repository, then records tool success and continues the provider with a safe result summary.

Run events persist title/type/size/excerpt/truncation/source ids only for text artifacts. Visual artifact result summaries may include bounded render content inside `artifacts[]` so the chat timeline and Preview drawer can rehydrate the visual. Events still do not include raw unbounded content, downloads, local paths, credentials, or raw provider payloads.

`artifact.create_text` result summaries now also expose a first-class `artifacts[]` resource reference. The first reference uses the persisted artifact id as `key`, carries safe presentation metadata (`title`, `filename`, `mime_type`, `display`), and keeps the legacy flat `artifact_id/title/text_excerpt` fields for existing consumers. The provider prompt tells Work-mode runs to create reports, articles, Markdown, and other saveable documents through `artifact.create_text`, then cite the returned resource as `[title](artifact:<key>)`.

`artifact.create_visual` result summaries use the same reference shape, set `operation = "create_visual"` and `artifact_type = "visual"`, and include the bounded SVG/HTML content on the first reference for rendering. The provider prompt tells Work-mode runs to use this tool for diagrams, charts, SVG drawings, and HTML mockups instead of dumping raw `<svg>` or fenced HTML/SVG into the final answer.

## Visibility

Settings > Tools shows artifact tools as artifact-scoped, approval-required, medium risk, and non-executable. RunRail labels artifact lifecycle rows separately from workspace, sandbox, LSP, web fetch, browser, MCP, and runtime tools.

The frontend also derives a lightweight `PreviewArtifact` projection from safe tool-call result summaries and run events. Completed artifact/document tools render as file resource cards in the chat timeline. Selecting a card opens the resizable right Preview drawer and renders Markdown/text excerpts or sandboxed SVG/HTML visual content when the summary contains preview content.

Assistant messages can also contain `artifact:<key>` Markdown links. The chat timeline strips the protocol link from prose, renders a compact card for the resource, and leaves the Preview drawer bounded to the safe artifact projection instead of displaying raw protocol text. If a provider still emits raw SVG text, Loomi extracts the SVG into a message-scoped visual artifact card so XML/CSS does not render as broken Markdown prose.

The chat transcript groups ordered model deltas, tool events, and continuation deltas into stable turns. When terminal tool events arrive before the next model delta, Loomi keeps a small waiting-for-model state in the transcript so the conversation does not visually collapse between tool completion and final generation.

This preview projection is intentionally display-only in the web runtime. It does not grant agent-side filesystem access, downloads, or raw unbounded tool output. In the desktop shell, the Preview drawer may write the bounded artifact body to an app-owned temporary file and ask the OS to open it with the default local application. Visual HTML/SVG runs in a sandboxed frame with no referrer and a restrictive CSP. Approval-required tools still use the normal tool-confirmation UI until the tool reaches a terminal state.
