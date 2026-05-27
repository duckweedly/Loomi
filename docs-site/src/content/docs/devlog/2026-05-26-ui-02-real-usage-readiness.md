---
title: 2026-05-26 UI-02 Real Usage Readiness
description: Real-use UI readiness notes, validation targets, and known blockers.
---

## Scope

UI-02 closes the demo-looking gaps in the existing UI-01 shell:

- Real API provider unavailable state points to Provider Settings without keeping the old runtime/context header in the chat canvas.
- Work mode no longer shows a disabled fake Work in Folder button.
- WorkPlanView behaves like a user task panel and does not invent plan steps without plan metadata.
- RunRail and ToolCallCard show human-first tool labels and keep raw runtime ids/event names out of the primary UI.
- RunRail and ToolCallCard share a safe preview formatter that redacts paths, `.env`, Authorization/cookie/token-like values, stdout/stderr, and raw body fields before rendering.
- Tool call history uses compact action rows by default, completed batches grouped behind one expandable summary, and detailed request/result fields only shown after expansion.
- Approval-blocked runs show a waiting-for-confirmation notice, Approve/Deny tool actions, and Stop.
- Sidebar removes the duplicate search field and bottom new/search action cluster while keeping thread rows and actions visible.
- The titlebar compose button creates a new thread for the current Chat/Work mode only when the sidebar is collapsed.
- Thread action menus use compact native-feeling typography, hover/active states, and light/dark surfaces instead of oversized floating text.
- Thread rename no longer uses a browser prompt; the selected row switches to an inline input with confirm/cancel controls.
- Mode switching creates a first thread for the target mode when none exists, so Chat/Work tabs do not appear stuck on the previous mode.
- Thread rows expose an explicit rename/delete menu.
- Composer renders working text entry, Stop for active runs, attachment picker, pasted-image chips, available model selector, and send button.
- Copy, retry, and regenerate moved to icon-only assistant message actions instead of living in the composer toolbar.
- Attachment chips are draft-only UI metadata for now; runtime file/PDF/image parsing is deferred to a later file-ingestion slice.
- Chat messages render basic safe Markdown for headings, lists, bold text, inline code, links, paragraphs, fenced code, and pipe tables.
- Composer placeholder copy differs between Chat and Work.
- Settings > Providers uses a management-list layout with provider search, All/Enabled/Local/Cloud filters, local provider cards, and an Add provider dialog for OpenAI-compatible provider save flow.
- Settings > Tools renders tool cards as readable title/code/description/badge groups and localizes built-in tool names, descriptions, and safe status badges for the Chinese UI.
- Sidebar workspace branding uses the supplied Loomi wordmark asset in light/dark variants, removes the placeholder `L` glyph and current-status label, and keeps the compact settings entry aligned with the card.
- The Progress floating rail now defaults to a concise current-run recent-activity feed with localized human labels, hides pipeline/model-stream noise, and prevents long run/message/persona/tool ids from wrapping into unreadable columns.
- Chat transcript now filters visible run UI against the latest user turn so old approval/tool cards cannot appear under a newer completed message.
- Real API message send preflights provider availability and active-run state before creating a durable user message, preventing orphan messages from binding to stale approvals.
- Real API list responses now validate their shape before mapping, so a transient null/invalid response becomes a clear API error instead of a raw `Cannot read properties of null` UI crash or a false "provider missing" state.
- Desktop dev now defaults the renderer API base URL to `http://127.0.0.1:18080` when no explicit `VITE_LOOMI_API_BASE_URL` is supplied, preventing `/v1` calls from accidentally hitting the Vite renderer server on `5180`.
- The shell visual language now follows Loomi's leaf identity: warmer cream surfaces, soft green/teal accents, rounded tactile controls, clearer selected states, and matching treatments across sidebar, composer, settings, provider cards, modal, and right-side run panels.
- Dark mode uses a neutral graphite base with restrained leaf accents, so the theme keeps the leaf identity without turning the whole product green.
- The second visual pass replaces thin glass/dashboard treatments with an island-style component system: two-pixel soft borders, raised button shadows, organic card radii, tactile segmented controls, and non-generic sidebar thread rows.
- Interaction polish now uses shared motion tokens for hover, press, focus, menu/modal entry, card/list entry, and active run pulses, with reduced-motion overrides for users who disable animation.
- The app shell now uses two visual layers instead of structural divider lines: the sidebar reads as the base layer, while chat, settings, and right-side panels float above it as compact rounded cards.
- Chat and right-side panels avoid nested card frames inside those outer panels; internal progress, preview, and composer regions render as content surfaces instead of stacked card shells.
- The chat shell removes the extra full-panel inset frame: sidebar, main content, and right drawer now occupy the outer shell directly without their own large-card border, radius, margin, or shadow.
- App identity assets now use the supplied gecko icon family across Dock, favicon/web manifest, Electron window icons, and titlebar mark; the Dock master is inset on a 1024 canvas so it matches neighboring macOS app icon scale instead of rendering oversized.
- The first brand refinement pass softens the shell into a gecko/leaf-led desktop workspace: less hard green outlining, unified titlebar/main/right-panel surfaces, gentler thread selection, branded message avatars, raised composer focus, and settings-to-mode switching that returns to the workspace instead of leaving stale settings content visible.
- Dark mode now uses restrained inset highlights instead of raised rectangular drop shadows for the sidebar footer, settings button, segmented controls, composer, main surface, right drawer, and dense cards.
- The visual system now targets a pastel compact green direction: cute through color and soft rounded states, but with 1px borders, minimal shadows, muted selected states, and calmer dark-mode surfaces.
- Secondary icon affordances such as titlebar toggles and thread row menus render as icon-only controls without persistent circular wrappers.
- The sidebar footer returns to a single compact Settings icon entry, removing the experimental day/night toggle treatment from the sidebar.
- Chat Markdown rendering now keeps fenced code blocks separate from inline code and tones down inline code styling so long prose does not turn into green chip blocks.
- Live run state now still accepts a late final assistant-content event after `run.completed`, so completed messages hydrate in the current view instead of only becoming well-rendered after the next thread refresh.
- Assistant drafts no longer render partial streaming Markdown as final prose; active runs show a restrained progress line and only render structured Markdown after final content is available.
- Work mode no longer injects the expanded WorkPlanView card into the main chat transcript; raw plan/progress projection stays out of the conversation surface until it has a compact, user-facing design.

## Non-goals

No M38/activity recorder work, no backend/runtime/provider/tool execution changes, no DB change, no new tool capability, no directory picker, and no provider fallback behavior change.

## Validation Targets

Closeout for this slice requires:

```bash
go test ./...
bun test --cwd web
bun run --cwd web build
bun run --cwd docs-site build
git diff --check
```

Browser smoke must verify:

- The old Mock/Real/runtime header is not visible in the chat canvas.
- Chat and Work composers accept input.
- Composer shows available model selection, file attachment, and pasted-image affordances without voice controls.
- Completed assistant messages show Copy and Regenerate below the answer; failed drafts show Retry below the failed assistant draft.
- Thread rows can open rename/delete actions.
- Basic Markdown and pipe tables render in assistant messages.
- Sidebar does not show the duplicate search field or bottom new/search action cluster.
- Titlebar compose creates a new current-mode thread only in the collapsed sidebar state.
- Electron titlebar icons align with the native window control centerline.
- Tool events are human-readable.
- Completed tool batches are collapsed by default and do not dominate the chat transcript.
- Approval blocked state shows Approve/Deny/Stop.
- Settings Providers and Tools open.
- Provider search/filter controls render, Add provider opens, and the type menu can be opened without console errors.
- Settings > Tools cards do not overlap badges with descriptions, and Chinese locale does not expose the built-in tool descriptions as raw English-only copy.
- Sidebar workspace branding shows the Loomi wordmark without the placeholder `L` glyph or current-status label.
- Leaf-style shell polish is visible in both light and dark modes without reducing chat/work readability.
- The shell no longer shows hard divider lines between sidebar/content or titlebar/content; the main and right panels read as rounded cards above the sidebar layer.
- Chat and right-side panels do not introduce card-inside-card frames for progress, preview, or inner content groups.
- Main chat content and right drawer fill the app shell without a visible nested full-panel frame on the left, top, or right edges.
- Dock, favicon, web manifest, Electron window, and titlebar icons show the gecko identity with Dock visual padding intact.
- Switching Chat/Work from Settings exits Settings and shows the corresponding workspace content.
- Dark mode interactions do not leave blocky rectangular shadows under the sidebar settings entry or common floating controls.
- Controls, cards, thread rows, composer, and menus avoid bulky toy-like depth while keeping a green Loomi identity.
- Titlebar toggles and thread row menus do not show persistent outer button shells; only hover/focus feedback is visible.
- The sidebar footer only shows a compact Settings icon button, without a separate theme switch or animated sky treatment.
- Long assistant prose with fenced `text` blocks remains readable, without `text` language markers or multiline content becoming inline code chips.
- A run that receives `run.completed` before the final assistant message still updates the visible completed bubble without waiting for the next user send or thread reload.
- Streaming assistant output does not expose malformed partial Markdown, blockquotes, code fences, or tables in the transcript while the model is still generating.
- Progress rail shows current-run recent activity without exposing raw ids/event names as primary text, and right-panel menu labels follow the active locale.
- Console error count is 0.
- Screenshot path is recorded.

## Known Blocker

Work folder selection remains intentionally blocked until a safe directory selection API and permission model exists. UI-02 exposes that limitation instead of pretending the control works.
