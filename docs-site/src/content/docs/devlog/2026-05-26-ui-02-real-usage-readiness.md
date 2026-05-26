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
- Tool call history follows the ArkLoop-style pattern: compact action rows by default, completed batches grouped behind one expandable summary, and detailed request/result fields only shown after expansion.
- Approval-blocked runs show a waiting-for-confirmation notice, Approve/Deny tool actions, and Stop.
- Sidebar removes the duplicate search field and bottom new/search action cluster while keeping thread rows and actions visible.
- The titlebar compose button creates a new thread for the current Chat/Work mode only when the sidebar is collapsed.
- Thread action menus use compact native-feeling typography, hover/active states, and light/dark surfaces instead of oversized floating text.
- Thread rename no longer uses a browser prompt; the selected row switches to an inline input with confirm/cancel controls.
- Mode switching creates a first thread for the target mode when none exists, so Chat/Work tabs do not appear stuck on the previous mode.
- Thread rows expose an explicit rename/delete menu.
- Composer renders only the working text entry, run actions, and send button; unimplemented attachment, persona/provider selector, Work folder, and voice controls are removed.
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
- Composer does not show unimplemented Work folder, attachment, persona/provider selector, or voice controls.
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
- Progress rail shows current-run recent activity without exposing raw ids/event names as primary text, and right-panel menu labels follow the active locale.
- Console error count is 0.
- Screenshot path is recorded.

## Known Blocker

Work folder selection remains intentionally blocked until a safe directory selection API and permission model exists. UI-02 exposes that limitation instead of pretending the control works.
