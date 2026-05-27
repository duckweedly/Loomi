---
title: 2026-05-26 M43 Memory Agent Tools
description: Agent-facing memory tools routed through ToolCatalog, ToolBroker, worker continuation, and safe summaries.
---

M43 adds the second memory parity slice after provider configuration: agent-facing `memory.*` tools.

Implemented:

- `memory.search`, `memory.read`, `memory.write`, `memory.forget`, and `memory.status` constants, catalog entries, validation, and built-in persona allowlist.
- Provider schemas and provider/internal tool-name mapping for all five memory tools.
- `MemoryToolExecutor` routed through `DefaultToolExecutor`, ToolBroker approval checks, worker resume jobs, and model continuation.
- Safe result summaries for search/read/status/write-proposal/forget without raw memory content or credential-like data.
- Settings > Tools type/copy/mock catalog support for the `memory` group.
- Spec Kit artifacts under `specs/043-memory-agent-tools/`.

Validation:

```bash
go test ./internal/productdata ./internal/runtime
bun test --cwd web src/components/SettingsView.tools.test.tsx src/components/SettingsView.runtime.test.tsx src/memory.test.ts
```

Deferred: automatic post-run distillation, semantic/vector retrieval execution, external memory adapters, background memory workers, and multi-agent long-term memory automation.
