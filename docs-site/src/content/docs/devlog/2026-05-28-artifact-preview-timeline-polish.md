---
title: 2026-05-28 Artifact Preview Timeline Polish
description: Makes generated documents first-class preview resources in the chat timeline.
---

Changed:

- Added a frontend `PreviewArtifact` projection that extracts document/resource metadata from tool-call result summaries and run events.
- Render completed artifact/document tools as compact document resource cards instead of inline request/result panels.
- Extract assistant-provided Markdown document payloads from `md` fenced blocks or accidental inline `md#...` output into the same compact document card path.
- Connected document resource cards to the existing right Preview drawer through `openArtifact`.
- Replaced the Preview placeholder with a resource preview state that renders Markdown/text excerpts when available.
- Normalized accidental `md#...` heading prefixes before Markdown rendering.
- Normalized ordinary `markdown#...` answer payloads as inline Markdown prose instead of extracting them as generated document cards.
- Unwrapped whole-answer `md`/`markdown` fenced blocks into rendered Markdown prose, while dropping empty fenced blocks that produced blank code cards.
- Repaired dense README-style Markdown without spaces after heading markers, including compact Chinese section titles, blockquote summaries, and ordered setup steps.
- Extracted assistant inline Markdown file payloads such as `` `markdown #...` `` into the artifact card path when the surrounding prose explicitly says it is Markdown file/document content, so these payloads do not render as oversized inline code chips.
- Applied the same artifact extraction to run transcript assistant blocks replayed from model delta events, keeping live/restored timelines consistent with persisted assistant messages.
- Added a recovery fallback for loose `markdown #...` payloads when stream restoration loses visible code-fence boundaries but the surrounding prose still identifies the content as a Markdown file/document.
- Extended dense Markdown repair coverage for common generated-document headings such as overview, goals, content, installation, build, directory structure, examples, changelog, and license sections.
- Hid restored completed assistant drafts when a persisted assistant final already exists after the latest user turn, avoiding duplicate final blocks with different whitespace from the same run.
- Suppressed assistant-text transcript replay when the final assistant message is already persisted for the current turn, while keeping tool transcript blocks available.
- Folded duplicate persisted assistant finals within the same user turn when the only difference is whitespace loss from stream restoration.
- Extended `artifact.create_text` result summaries with safe `artifacts[]` resource refs containing `key`, `filename`, `mime_type`, and `display`.
- Added assistant `artifact:<key>` link extraction so final replies can cite generated documents without dumping the document body into chat.
- Kept a small waiting-for-model transcript state after terminal tool events and before continuation text arrives.
- Added a chat-level workspace authorization prompt when the latest user request targets Downloads or local files but the selected workspace is missing or points somewhere else.
- Paused real sends before starting a model run for those workspace-authorization cases, so Loomi persists the user request and waits for folder selection instead of producing a generic "no local file tool" reply.
- Refined the authorization card in the late unified theme layer with compact sizing, readable dark-mode CTA contrast, mobile wrapping, and suppression of stale "paste the file list" fallback replies.
- Promoted the current chat to Work mode after desktop folder authorization, and also before follow-up sends like "现在呢" when the conversation is continuing a just-authorized workspace request.
- Added the same authorization pause/card path for macOS Documents/文稿 directory requests so Loomi asks for the right folder instead of reading the currently selected Downloads workspace.
- Persisted visible partial assistant output when a model run is stopped before `model_output_completed`, preserving interrupted replies in message history after refresh or the next user turn.
- Normalized `artifact.*` internal tool names to OpenAI-compatible function names before provider requests, preventing upstream HTTP 400 failures when Work mode includes artifact tools.
- Scoped runtime warning state to the selected thread so a stale completed run cannot leak a "missing final reply" banner into an empty new conversation.
- Reworded the rare missing-final warning as a user-facing retry hint instead of exposing internal `assistant final message` persistence terminology.
- Moved the rare missing-final warning into the chat transcript and gated it on a real latest user turn, so empty new conversations never show an internal run banner.
- Rendered the rare no-final-content state as a normal assistant transcript row instead of a floating system banner, and promoted restored assistant transcript text from completed runs as the visible final answer.
- Collapsed single completed non-artifact tool events into the same compact activity summary used by multi-tool turns, while keeping completed artifact tools as preview resource cards.
- Split theme handling into a user preference (`system`/`light`/`dark`) and the resolved display theme, so fresh sessions follow the OS appearance unless the user pins a manual choice in Settings.
- Added a Settings theme segment for `System`, `Light`, and `Dark`, matching the desktop expectation that restart should not force a light screen at night.
- Reworked grouped tool activity summaries from raw tool-name lists into intent/state summaries such as searched/read/inspected counts, keeping the chat timeline compact until details are expanded.
- Added disabled states for unavailable Composer `+` menu actions so the compact context menu no longer presents inert clickable rows.
- Split Settings `Appearance` out of `General`, moving language/theme controls into their own page and leaving General focused on workspace defaults.
- Refined Settings navigation toward the Craft-style mechanism of icon, title, one-line description, and subtle state text instead of status-heavy rows.
- Reduced Settings status noise by hiding routine `working`/`available` badges in navigation, page headers, and rows; only mixed/read-only/preview states remain visibly labeled.
- Reworked Provider and Web Search settings internals from oversized management cards/forms into compact scan-first rows with inline meta, quiet badges, and right-aligned actions.
- Narrowed the default open sidebar and suppressed sidebar loading rows during background thread refreshes, reducing visual flicker when switching or archiving conversations.
- Reworked MCP settings from a bulky local-admin form into compact editable rows and quieter server actions, following Craft's object-row mechanism without copying its expression layer.
- Flattened routine General and Appearance settings into the same compact row grammar, reducing card weight around language, theme, and workspace defaults.
- Tightened the default open sidebar to 264px with a compact 236-320px resize range, and kept thread refreshes out of the blocking loading state once content exists so switching or deleting conversations does not flash through empty/loading frames.
- Removed the duplicate post-archive refresh path for selected-thread deletion; the settled selection now owns the refresh.
- Refined sidebar thread rows to a 34px desktop list rhythm: row-level containment, no hidden title column reserved for the menu button, hover/menu-open action reveal, and menu-open row state.
- Removed the duplicate top approval notice so permission requests appear once as the chat-flow tool confirmation card, matching the turn-local approval pattern.
- Repaired collapsed pipe-table summaries such as `|类型|数量||---|---:||...` before Markdown rendering, so dense directory classifications render as real tables instead of giant headings.
- Made the main titlebar's long-thread-title clamp mobile-safe, preserving ellipsis without collapsing the title to zero width on narrow windows.
- Persisted user-adjusted sidebar width within the compact 236-320px bounds, so manual resizing survives reloads without widening the default shell.
- Clamped sidebar thread action menus to the visible viewport, matching the collision-aware menu behavior studied from Craft so low rows never leave only a clipped menu edge visible.
- Tightened the default open sidebar again to 248px with a compact 224-300px resize range, so the desktop shell opens with less left-side weight.
- Added per-thread timeline snapshots and loaded target thread content before applying a selection/delete transition, reducing old-content and empty-state flashes when switching or archiving conversations.
- Separated routine Settings section headers from grouped row surfaces, removed first-row double dividers, and narrowed routine settings content width toward Craft's compact settings rhythm.
- Removed category-level status badges from Settings navigation and page headers, leaving status detail inside the relevant rows or page-specific controls.
- Tightened the default sidebar again to 236px with a 216-288px resize band, and skipped the redundant selected-thread refresh after a ready thread snapshot is applied so switching/deleting conversations avoids a second visual bounce.
- Tightened the default sidebar again to 224px with a 208-272px resize band, keeping the open shell lighter while preserving user-resized widths.
- Replaced the decorative yellow wave between user turns with a quiet spacing gap, moving the chat timeline closer to the Craft-style turn rhythm without copying its expression layer.
- Made assistant message copy/retry/regenerate actions reveal on hover or keyboard focus, matching the quieter Craft-style row action mechanism while keeping the controls accessible.
- Replayed `assistant.message.completed`, `message.model_output_completed`, and `model.final` event content inside the run transcript, so final answers remain visible after tool events instead of being hidden by the tool activity block.
- Downgraded the rare no-final-content fallback from a fake assistant bubble into a neutral inline run notice, avoiding confusing "Loomi said no reply" transcript entries.
- Tightened the default sidebar to 216px with a 196-256px resize band, matching the current compact desktop shell target.
- Added Craft-style local sidebar scanning: search, recency groups, compact thread metadata, and stable background-refresh rows so switching/deleting conversations avoids loading or empty-state flashes.
- Refined assistant code blocks into compact content surfaces with an inline 32px header, reduced code padding, quieter copy controls, and contained horizontal scrolling instead of the previous oversized floating-header panel.
- Refined Settings Tools and Skills catalog rows from badge-heavy cards into compact scan rows: title/code, quiet operational metadata, and only the essential safety chips.
- Narrowed the Settings navigator column and softened its group/description treatment, with a small-screen two-column navigator so the Settings shell no longer feels like a second oversized sidebar.
- Tightened the default left sidebar to 204px with a 188-244px resize band, and skipped identical thread snapshot re-application so switching or deleting conversations avoids a redundant redraw.
- Kept selected sidebar thread rows visually quiet by revealing the action menu only on hover, keyboard focus, or while the menu is open; selection now only communicates current context.
- Refined workspace authorization prompts into a narrower inline confirmation card with an explicit text-and-arrow action, so folder permission requests read as a small next step instead of a wide alert.
- Removed the duplicate Theme row from General settings and made Appearance the single compact display-preference page, with system/light/dark selection showing the currently resolved theme without adding another card layer.
- Reworked About settings from a placeholder-status card into the same quiet routine row grammar, removing read-only/mock badges and stale placeholder copy from basic app information.
- Reworked Web Search settings into the same routine row grammar, keeping Tavily/Brave key entry and tool readiness visible without a wide card shell or code-chip status block.
- Reworked MCP settings controls to match the same compact row grammar: narrower content width, neutral save action, and low-noise per-server action buttons instead of provider-style primary buttons.
- Simplified the inner Settings navigator into a directory-style rail: slimmer column, quiet back row, icon + label entries, and accessible descriptions moved out of the clipped visible row.
- Narrowed the default open sidebar to 136px with a 128-172px resize band, invalidated older wider stored defaults, memoized sidebar grouping, and disabled leftover row entrance animation so switching or deleting conversations no longer produces a visible flash.
- Forced a final message reconciliation pass when the run event stream reaches a terminal event, even if the stream effect is being cleaned up, so completed replies appear immediately instead of only after refresh or thread switch.
- Scoped persisted-assistant suppression to the current run or current turn, so older replies in the same thread no longer hide the new run's thinking/drafting placeholder before the final response arrives.
- Added an optimistic send snapshot that inserts the user's new turn and a pending assistant draft immediately, then replaces it with the backend run once the real response arrives.
- Reworked Memory provider choices from three card buttons into a single row-style selector with inline status and quiet controls, keeping the service state visible without a separate heavy card stack.
- Reworked the Memory management surface itself into the same quiet row grammar: search/filter, manual add, pending proposals, saved memories, and history now avoid stacked cards and use thin separators plus subdued actions.
- Reworked Runtime Memory snapshots into the same row grammar, so impression/overview panels no longer sit as a second card stack under Memory management.
- Tightened the default sidebar again to 168px with a 156-204px resize band, clears older wide default widths, and changes thread switch/archive transitions so uncached conversations swap only after the target snapshot is ready.
- Fixed selected-thread deletion so the next thread snapshot is still fetched after the optimistic row selection, avoiding a stale timeline or blank refresh flash.
- Excluded deferred workspace-authorization turns from the no-final-content notice, so folder permission requests stop on the authorization card instead of showing a misleading missing-reply state.
- Tightened the default sidebar to 156px with a 148-196px resize band, treats the previous 168px default as migratable, and creates new conversations only after their first snapshot is ready to avoid a blank-frame flash.
- Made running and failed non-approval tools default to compact activity rows, reserving phase strips and heavier panels for approval-required tools or explicit expansion.
- Refined Provider setup internals again: search/filter, empty state, capability list, and add-provider dialog now use compact rows and a small sheet instead of stacked management cards.
- Tightened the Settings detail rhythm again: smaller page headers, narrower routine content, a slimmer MCP form, and responsive Tools/MCP metadata rows that avoid right-side overflow.
- Tightened the default sidebar to 148px with a 140-188px resize band, treating the previous 156px default as migratable so existing sessions adopt the lighter shell.
- Tuned routine Settings rows again: 580px content width, tighter row/control gaps, narrower right-side controls, and smaller segmented controls to better match the Craft-style settings mechanism without adding another card layer.
- Reworked chat user turns toward Craft's interaction mechanism: user messages now sit as right-aligned compact bubbles without the left avatar/meta rail, and the message scroll region uses a top/bottom fade mask to reduce hard clipping near the composer.
- Promoted bare generated Markdown documents into Loomi artifact cards when the assistant returns a dense `markdown#...` document-shaped payload, leaving ordinary short markdown answers in the chat transcript. This keeps file-like output out of the conversational text flow and removes the empty markdown shell after extraction.
- Kept the optimistic "thinking" turn visible for a short minimum interval before fast final results replace it, so quick runs no longer flash directly from user input to a completed answer.
- Let fast completed assistant messages typewrite once when no live stream deltas were observed, preserving Craft-like conversation pacing even when the backend only returns the final persisted message.
- Changed active chat scrolling to bottom-follow mode: streaming updates only keep the view pinned when the user is already near the bottom, and manual scroll-away pauses auto-follow to avoid fighting the user's scroll.

Boundaries:

- This is a frontend projection over existing safe result summaries; it does not expose raw unbounded tool output.
- Approval-required tools still render their confirmation controls.
- Non-artifact tools keep the existing compact/expandable tool-card path.

Validation:

```bash
bun test web/src/runtime/artifactPreview.test.ts web/src/runtime/markdownNormalize.test.ts web/src/components/ToolCallCard.test.tsx web/src/components/RightToolDrawer.preview.test.tsx web/src/useWorkspaceShellState.test.ts
bun test web/src/runtime/messageArtifactPreview.test.ts web/src/components/ChatCanvas.states.test.ts web/src/components/RightToolDrawer.preview.test.tsx
bun test --cwd web src/runtime/markdownNormalize.test.ts src/runtime/messageArtifactPreview.test.ts src/components/ChatCanvas.states.test.ts
bun test --cwd web src/components/ChatCanvas.states.test.ts src/runtime/messageArtifactPreview.test.ts src/runtime/markdownNormalize.test.ts
bun test --cwd web src/components/ChatCanvas.states.test.ts src/runtime/messageArtifactPreview.test.ts src/runtime/markdownNormalize.test.ts
bun test --cwd web src/realApiClient.test.ts src/components/ChatCanvas.states.test.ts src/runtime/messageArtifactPreview.test.ts src/runtime/markdownNormalize.test.ts
bun test --cwd web src/components/ChatCanvas.states.test.ts src/realApiClient.test.ts src/runtime/messageArtifactPreview.test.ts src/runtime/markdownNormalize.test.ts
bun test --cwd web src/state.runtime.test.ts -t "desktop workspace authorization promotes"
bun test --cwd web src/state.runtime.test.ts -t "continues a workspace authorization flow"
go test ./internal/runtime -run TestGatewayPersistsPartialAssistantMessageWhenStoppedAfterVisibleOutput -count=1
go test ./internal/runtime -run TestHTTPProviderSerializesArtifactToolNamesForOpenAI -count=1
bun test --cwd web src/components/ChatCanvas.states.test.ts -t "does not leak a stale completed run warning|flags a completed real API run"
bun test --cwd web src/useWorkspaceShellState.test.ts src/components/SettingsView.runtime.test.tsx src/components/SettingsView.mcp.test.tsx src/components/SettingsView.skills.test.tsx src/components/SettingsView.tools.test.tsx
bun test --cwd web src/components/ChatCanvas.states.test.ts src/components/ToolCallCard.test.tsx
bun test --cwd web src/components/ChatCanvas.states.test.ts src/components/Composer.test.ts
bun test --cwd web src/components/SettingsView.layout.test.tsx src/components/SettingsView.runtime.test.tsx src/components/ChatCanvas.states.test.ts
bun test --cwd web src/runtime/markdownNormalize.test.ts src/components/ChatCanvas.states.test.ts
bun test --cwd web src/components/ChatCanvas.states.test.ts -t "restored assistant transcript text|final content as a normal assistant recovery row"
bun test --cwd web src/components/ChatCanvas.states.test.ts -t "single completed tool|completed artifact tool|waiting-for-model"
bun test --cwd web src/components/SettingsView.layout.test.tsx src/components/SettingsView.runtime.test.tsx
bun test --cwd web src/components/SettingsView.layout.test.tsx -t "provider and web-search pages use compact settings rows"
bun test --cwd web src/components/ThreadSidebar.actions.test.ts -t "keeps existing thread rows stable"
bun test --cwd web src/useWorkspaceShellState.test.ts -t "starts with a narrower open sidebar"
bun test --cwd web src/components/SettingsView.tools.test.tsx -t "renders web search as a dedicated settings menu"
bun test --cwd web src/components/SettingsView.layout.test.tsx -t "mcp settings use compact"
bun test --cwd web src/components/SettingsView.layout.test.tsx -t "routine settings sections use compact"
bun test --cwd web src/useWorkspaceShellState.test.ts src/state.runtime.test.ts -t "narrower open sidebar|flashing through stale refreshes|without issuing a duplicate refresh"
bun test --cwd web src/useWorkspaceShellState.test.ts src/state.test.ts -t "narrower open sidebar|shouldSendWorkspaceRefreshIntoLoading"
bun test --cwd web src/components/ChatCanvas.states.test.ts -t "active approval requests"
bun test --cwd web src/components/ThreadSidebar.actions.test.ts -t "visually stable|sibling row menu"
bun test --cwd web src/App.controls.test.ts -t "long thread titles"
bun test --cwd web src/useWorkspaceShellState.test.ts -t "persists user-adjusted sidebar width|starts with a narrower open sidebar"
bun test --cwd web src/components/ThreadSidebar.actions.test.ts -t "visible viewport"
bun test --cwd web src/useWorkspaceShellState.test.ts src/state.runtime.test.ts src/state.test.ts
bun test --cwd web src/components/SettingsView.layout.test.tsx -t "routine settings sections"
bun test --cwd web src/components/SettingsView.layout.test.tsx -t "category status"
bun test --cwd web src/useWorkspaceShellState.test.ts src/state.runtime.test.ts -t "compact open sidebar|follow-up selected-thread refresh"
bun test --cwd web src/useWorkspaceShellState.test.ts src/state.runtime.test.ts
bun test --cwd web src/components/ThreadSidebar.actions.test.ts src/components/ThreadSidebar.layout.test.ts
bun test --cwd web src/useWorkspaceShellState.test.ts src/state.runtime.test.ts
bun test --cwd web src/components/ChatCanvas.states.test.ts
bun test --cwd web src/components/ChatCanvas.states.test.ts
bun test --cwd web src/components/ChatCanvas.states.test.ts -t "renders final event content"
bun test --cwd web src/components/ChatCanvas.states.test.ts -t "without final content"
bun test --cwd web src/components/ThreadSidebar.actions.test.ts src/components/ThreadSidebar.layout.test.ts src/useWorkspaceShellState.test.ts src/state.runtime.test.ts
bun test --cwd web src/components/ChatCanvas.states.test.ts src/runtime/markdownNormalize.test.ts
bun test --cwd web src/components/SettingsView.tools.test.tsx src/components/SettingsView.skills.test.tsx src/components/SettingsView.layout.test.tsx
bun test --cwd web src/components/SettingsView.layout.test.tsx
bun test --cwd web src/useWorkspaceShellState.test.ts src/state.test.ts src/state.runtime.test.ts
bun test --cwd web src/components/ThreadSidebar.actions.test.ts src/components/ThreadSidebar.layout.test.ts
bun test --cwd web src/components/ChatCanvas.states.test.ts -t "workspace authorization card"
bun test --cwd web src/components/SettingsView.layout.test.tsx
bun test --cwd web src/components/SettingsView.layout.test.tsx src/components/SettingsView.runtime.test.tsx
bun test --cwd web src/components/SettingsView.layout.test.tsx -t "provider and web-search pages use compact settings rows"
bun test --cwd web src/components/SettingsView.layout.test.tsx -t "mcp settings use compact"
bun test --cwd web src/components/SettingsView.layout.test.tsx -t "settings navigator stays narrow"
bun test --cwd web src/components/SettingsView.layout.test.tsx -t "memory provider settings use row-style"
bun test --cwd web src/components/SettingsView.layout.test.tsx -t "runtime memory snapshot uses the same row grammar"
bun test --cwd web src/components/MemoryPanel.test.tsx -t "unified workspace styles keep memory management row-based"
bun test --cwd web src/useWorkspaceShellState.test.ts src/state.runtime.test.ts src/state.test.ts
bun test --cwd web src/useWorkspaceShellState.test.ts src/state.runtime.test.ts -t "narrow open sidebar|archives selected threads"
bun test --cwd web src/components/SettingsView.layout.test.tsx -t "provider setup empty states"
bun run --cwd web build
go test ./internal/runtime ./internal/productdata
bun test --cwd web src/runtime/markdownNormalize.test.ts src/runtime/messageArtifactPreview.test.ts src/components/ChatCanvas.states.test.ts
bun test --cwd web src/state.runtime.test.ts src/components/ChatCanvas.states.test.ts
bun run build
bun test --cwd web src/components/ChatCanvas.states.test.ts src/state.runtime.test.ts
bun run --cwd web build
bun run build
```
