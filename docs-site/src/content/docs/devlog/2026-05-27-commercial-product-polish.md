---
title: 2026-05-27 Commercial Product Polish
description: Frontend polish pass that moves Loomi from demo-like UI toward a commercial agent workspace.
---

## Scope

This pass tightens the existing web shell without changing backend behavior, routes, data models, or tool execution semantics.

- Adds a final product polish CSS layer for the web shell, settings, memory, chat composer, run rail, tool cards, and drawers.
- Removes the demo-card treatment from the mode switcher, selected thread row, chat transcript bubbles, shell frame, and composer; those surfaces no longer render as nested rounded cards inside a fake window.
- Reframes Settings navigation from engineering buckets toward product tasks: workspace, model/context, capabilities/safety.
- Moves Settings > Memory toward a command-center flow: review saved/proposed memory first, then runtime context, then provider connections.
- Updates Memory copy from a saved-list view to a memory console focused on review, search, and safe runtime summaries.
- Gives ToolCallCard state-specific visual treatment, so approval, running, failed, denied, and completed tools no longer read as the same successful card.
- Adds an accessible label to the icon-only composer submit button.
- Adapts useful AI Elements patterns without importing its shadcn/Tailwind stack: tool invocation phases, source chips for web search evidence, and a compact work-plan queue summary.

## Design Direction

The target is a commercial agent workspace:

- Run, tool, memory, provider, and worker states should feel observable and auditable.
- Green is reserved for healthy/completed states instead of washing the whole product.
- Warning, danger, and info states have distinct semantic colors.
- Panels use a consistent radius system: shell panels, cards, controls, and pills each have a clear role.
- The Memory surface prioritizes user trust: proposals, saved memories, runtime summaries, provider health, and audit history.
- AI Elements is used as a mechanism reference only. Loomi keeps its own visual language, runtime event model, redaction rules, and existing component stack.

## Non-goals

No new memory provider behavior, no new tool permissions, no backend API changes, no database migrations, and no route restructuring.

## Validation Targets

```bash
bun run --cwd web build
bun run --cwd docs-site build
```

Manual browser smoke should verify:

- The app shell reads as a single commercial desktop workspace rather than nested demo cards.
- The outer shell, Chat/Work switcher, selected thread, user/assistant messages, and composer do not render as nested filled cards or double-window chrome.
- Settings > Memory starts with the memory console and review queue before technical connection details.
- Runtime context and memory provider configuration remain reachable below the console.
- Tool approval cards stand out from completed tool cards.
- Failed/running/denied tool cards use distinct state styling.
- Tool cards expose request, approval, execution, and result phases without raw tool ids.
- Web search tool results show bounded source chips with safe titles and hosts.
- Work mode shows a task/queue summary before detailed steps, todos, artifacts, and recent progress.
- Composer submit remains icon-only but has an accessible name.
- Light and dark themes keep readable text, visible focus states, and consistent panel radii.
