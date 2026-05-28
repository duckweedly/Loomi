---
title: 2026-05-27 M94 Unified Workspace Theme
description: Consolidates Loomi's frontend visual override stack into a single terminal theme layer.
---

M94 consolidates the frontend style stack after multiple green, pastel, compact, AC-site, and commercial-product layers had accumulated as terminal overrides.

Changed:

- Kept base structural styles in their existing module files.
- Removed the old experimental terminal theme imports from `web/src/styles.css`, including the previous brand-refinement, compact, pastel, AC-site, and commercial-product layers.
- Added `web/src/styles/92-unified-workspace.css` as the single imported terminal visual layer.
- Re-centered the current Loomi expression around neutral light/dark workspace tokens, restrained blue accents, compact controls, quiet thread rows, flat panels, and consistent composer/menu/tool-card states.
- Tightened the dark theme after visual review: selected threads now use neutral contrast, message avatars no longer keep the blue framed treatment, and the composer uses a quieter dark surface with clearer placeholder and control states.
- Reworked the sidebar settings affordance from a detached icon into a full-width footer row with icon, label, chevron, and a subtle divider so it belongs to the sidebar navigation.
- Fixed sidebar overflow from long thread titles by applying the grid and ellipsis constraints to the inner `animal-island-ui` button content wrapper and hiding horizontal overflow at the workspace/settings boundaries.
- Refined sidebar texture after visual review: selected rows now use a single neutral row surface without nested borders or status rails, the row action is vertically centered, and the settings footer follows the Codex-style single pill control instead of an outlined nested card.
- Removed the detached row-action chip background so the thread overflow control reads as plain trailing dots rather than a second card layered on top of the selected row.
- Added outside-click and Escape dismissal for the thread overflow menu so it stays selected only while the pointer is inside the action or menu.
- Refined the composer run-state control: the primary send button now becomes the stop button during active runs, and the separate composer/runtime rail stop controls are no longer shown as extra blocks.
- Reworked the composer context entry into a single `+` menu for files/photos, folder/workspace, skills, connectors, and plugins, with a compact workspace status chip beside it instead of duplicated directory copy.
- Reduced the composer context menu density after visual review so it fits Loomi's compact shell instead of reading as an oversized imported product menu.
- Compressed completed tool-call rendering in the chat transcript into lightweight disclosure rows; phase strips now stay reserved for approval/running tool states where the user needs live context.
- Refined completed tool-call rows against Craft and Ark Loop references: completed tools now default to a single transparent activity row with icon, human action, compact preview, status, and disclosure; request/result/source details are only shown after expansion.
- Repaired dense Chinese Markdown output where streamed model text collapses `##` headings and `1.-` list markers onto the previous sentence; chat rendering now inserts safe structural breaks outside fenced code so long reports do not become one oversized heading.
- Added a bottom composer occlusion layer so transcript content no longer remains visible underneath the fixed input surface.
- Removed the redundant titlebar run-details entry; tool execution is now surfaced in the transcript and the remaining titlebar utility button opens the right-side tools menu.
- Changed theme startup to follow the system color scheme until the user manually switches light/dark, then stores that explicit choice for later restarts.
- Changed active-run transcript rendering to replay run events chronologically: assistant text segments, tool calls, and continuation text now appear in the same order they happened instead of aggregating current tools at the bottom of the chat.
- Moved thread overflow menus to viewport-positioned popovers so sidebar overflow constraints no longer clip rename/delete menus.
- Added a shared `LoomiFloatingMenu` primitive for compact menus. Thread row actions and the composer `+` context menu now share portal rendering, outside-click dismissal, Escape dismissal, item/separator semantics, and the same floating surface styling instead of each component hand-rolling its own menu behavior.
- Moved assistant Markdown normalization into `runtime/markdownNormalize`, with focused tests for streamed fence starts and dense Chinese report output. The chat canvas now consumes a single rendering boundary instead of keeping parser repairs inside the component.
- Updated style contract tests so they protect the unified import order instead of older patch layers.

Validation:

```bash
bun test web/src/animalIslandUi.test.ts web/src/components/ThreadSidebar.layout.test.ts web/src/components/SettingsView.layout.test.tsx web/src/components/RunRail.polish.test.ts web/src/components/ChatCanvas.states.test.ts web/src/components/Composer.test.ts
bun test web/src/components/LoomiMenu.test.ts web/src/components/Composer.test.ts web/src/components/ThreadSidebar.actions.test.ts web/src/components/ThreadSidebar.layout.test.ts
bun run build
```

Browser smoke:

- Opened `http://127.0.0.1:5181/`.
- Verified the main shell, sidebar, error state, composer disabled state, and Settings layout render without panel collapse.
- Full chat-flow smoke was blocked by the local API CORS response from `http://127.0.0.1:18080`, so the browser session showed `Failed to fetch` rather than live thread data.

Known follow-up:

- The old CSS files are no longer imported, but they were left in place to avoid a large deletion in the same change. A later cleanup can delete the unused historical theme files after review.
