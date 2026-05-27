# Feature Spec: M55 Memory Provider Config Modal

## Goal

Make Settings > Memory match the Arkloop-style configuration flow more closely: the page shows service toggles, provider choice, status, and a compact Configure action; provider-specific fields live in a modal.

## Requirements

- Keep memory enablement and post-run organization toggles on the main Memory surface.
- Keep provider selection on the main Memory surface.
- Move Nowledge/OpenViking provider details into a modal opened by Configure.
- Keep Nowledge local detection inside the Nowledge configuration modal.
- Do not expose provider secrets; keep key inputs write-only and status based on key presence.
- Do not add install/restart/process-control or bulk-delete behavior.

## Non-Goals

- Do not change provider persistence semantics.
- Do not execute external memory provider adapters.
- Do not copy external product names, branding, or private copy beyond provider names already configured in Loomi.
