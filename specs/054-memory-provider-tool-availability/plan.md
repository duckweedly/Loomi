# Plan: M54 Memory Provider Tool Availability

## Slice

Reuse the existing ToolCatalog and RunContext boundaries. Add a provider-aware filter in productdata so both Settings > Tools and worker-prepared run contexts reflect the selected memory provider.

## Design

- `ApplyMemoryToolAvailability` decorates memory catalog entries with active provider and provider list metadata.
- Unsupported tools become disabled catalog entries with `approval_policy=disabled` and a short `disabled_reason`.
- `FilterMemoryToolResolutionsForProvider` removes unsupported memory tools from run contexts after memory readiness is resolved.
- Nowledge removes only `memory.edit`; disabled/unconfigured memory removes all memory tools.

## Validation

- Unit tests cover Nowledge catalog disabling and RunContext filtering.
- Run productdata validation, web build, docs build, and diff whitespace check.
