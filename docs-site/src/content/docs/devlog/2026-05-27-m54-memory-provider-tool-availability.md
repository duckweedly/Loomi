---
title: M54 Memory Provider Tool Availability
description: Provider-aware memory tool exposure for Nowledge/OpenViking parity.
---

## Completed

- Added provider-aware memory tool catalog availability.
- Disabled `memory.edit` for Nowledge and included safe provider metadata/reason codes.
- Filtered unsupported memory tools out of prepared RunContext so workers do not advertise unavailable actions.
- Kept local, semantic, and OpenViking on the full current memory tool set.

## Validation

- `go test ./internal/productdata -run 'TestToolCatalogHidesNowledgeUnsupportedMemoryEdit|TestNowledgeRunContextFiltersUnsupportedMemoryEdit|TestToolCatalogIncludesMemoryRuntimeTools' -count=1`
- `bun run --cwd web build`
- `bun run --cwd docs-site build`
- `git diff --check`
