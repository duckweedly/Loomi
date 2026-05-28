---
title: UI Animal Island Component Source
description: Notes for using animal-island-ui directly and ac-site-template theme tokens.
---

Loomi's cozy UI direction now uses `animal-island-ui` as a direct component source instead of only recreating a similar style:

- `animal-island-ui` is installed in `web/` and its stylesheet is imported at the app entry.
- Composer's primary send action uses `Button` directly from `animal-island-ui`.
- Composer's model picker uses `Select` directly from `animal-island-ui`, removing the native `<select>` from the main input bar.
- Memory filters use `Select` and `Switch` directly from `animal-island-ui` for scope/source/deleted controls.
- Tool approval actions use `Button` directly from `animal-island-ui`, and the old Lobe tag dependency was removed from tool cards.
- The AC theme layer now loads last and gives the main composer a visible thick-border command bar, stronger select surface, and raised send action.
- Sidebar mode buttons, the thread create affordance, and the settings footer action now use `Button` directly from `animal-island-ui`.
- Sidebar project/scheduled entries, thread rows, and thread action affordances now use `Button` directly from `animal-island-ui`.
- Titlebar chrome controls now use `Button` directly from `animal-island-ui`.
- Streaming assistant drafts now render Markdown as soon as content arrives; `animal-island-ui` `Typewriter` is only used when a completed answer appears without prior visible streaming, and completed/streamed run keys are remembered in session storage so returning to the page does not replay the same answer.
- Streaming unfinished fenced code now opens the code block container immediately, including common cases where the language marker and first code token arrive in the same chunk, so SQL/JS/Python/Bash snippets do not flash as inline text before becoming a block.
- Chat history now uses the `animal-island-ui` `Divider` `wave-yellow` variant between conversation turns, starting before the second user message in a thread.
- Markdown code blocks now keep Loomi's renderer but follow a Codex-like minimal container: thin neutral border, generous code padding, quiet copy affordance, and light/dark variants that adapt through Loomi tokens instead of fixed package colors.
- Navigation chrome now has AC-style thick borders, raised active states, and a stronger titlebar so the skin is visible beyond the composer.
- Settings navigation, provider filter tabs, provider add/test actions, provider empty state, and sidebar retry now use the same raised animal/ac surfaces.
- Dark settings headers no longer keep the old blue/purple block; the settings workspace uses the shared AC day/night surface tokens.
- The visible animal treatment is now on the controls themselves, not on nested wrapper cards: sidebar mode, retry state, chat shell, composer, settings header, and provider empty state keep flat surrounding surfaces.
- The dark app shell now uses one consistent blue-purple surface with a blue-violet accent, so the sidebar, titlebar, main canvas, settings, and composer no longer drift across unrelated palettes.
- Backend failure screens now avoid repeated full-width error pills: the center state carries the failure, provider guidance hides during backend failure, and the composer collapses to a compact toolbar.
- Dark sidebar and main panel now share the same surface color, the chat/work mode control is a single segmented control, and backend failure hides the composer instead of leaving an empty bottom rectangle.
- A dedicated `87-ac-site-theme.css` layer maps day/night tokens learned from `yunxinz/ac-site-template` into Loomi's existing theme variables.
- The theme mapping keeps Loomi's own app shell and product copy, while borrowing the day/night token structure and tactile AC surface hierarchy.

Validation:

- `bun test --cwd web ./src/animalIslandUi.test.ts ./src/components/Composer.test.ts`
- `bun test --cwd web ./src/components/ChatCanvas.states.test.ts`
- `bun run --cwd web build`
