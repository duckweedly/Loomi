---
title: M84 Live Markdown Finalization
description: Closeout notes for promoting completed assistant content over collapsed stream drafts.
---

Fixed a desktop live-rendering regression where provider token deltas could be concatenated into a collapsed assistant draft and then displayed as the completed response.

What changed:

- Real API event mapping now treats `message.model_output_completed`, `assistant.message.completed`, and `model.final` as completed assistant content even when the API category is `message`.
- Runtime state now promotes any assistant final-content event to the assistant draft before later terminal run events arrive.
- The rendered assistant message uses the completed provider content instead of reconstructing Markdown from token fragments.

Validation:

- `bun test --cwd web ./src/realApiClient.test.ts ./src/state.runtime.test.ts`
