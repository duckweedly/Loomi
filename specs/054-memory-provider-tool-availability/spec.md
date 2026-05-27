# Feature Spec: M54 Memory Provider Tool Availability

## Goal

Align Loomi memory tool exposure with the selected memory provider so Settings > Tools and run contexts do not advertise provider-unsupported memory actions.

## Requirements

- Memory catalog entries must include safe provider availability metadata.
- When memory is disabled or the selected provider is unconfigured, memory tools are disabled in the catalog and omitted from run contexts.
- When the selected provider is Nowledge, `memory.edit` is disabled/omitted because Nowledge exposes the semantic memory subset without edit semantics.
- Local, semantic, and OpenViking providers keep the full current memory tool set.
- Disabled catalog entries must use safe reason codes only; no provider secrets, URLs, traces, or local paths.

## Non-Goals

- Do not implement external Nowledge/OpenViking adapters.
- Do not add process control, install, restart, or bulk-delete actions.
- Do not change persona allowlists or approval semantics.
