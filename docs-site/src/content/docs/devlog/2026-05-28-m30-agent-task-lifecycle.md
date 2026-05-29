---
title: 2026-05-28 M30 Agent Task Lifecycle
description: Durable coordination-only lifecycle states for multi-agent tasks.
---

## Completed

- Added approval-gated builtin `agent.start` and `agent.fail` alongside `agent.spawn`, `agent.list`, and `agent.complete`.
- Extended `agent_tasks` from `spawned/completed` to `spawned/in_progress/completed/failed` with a migration that preserves the existing table.
- Added in-memory and PostgreSQL lifecycle methods for start, complete, and fail, with terminal-state guards so completed or failed tasks cannot be restarted or overwritten.
- Kept the runtime coordination-only: no autonomous child model runs, cross-thread delegation, external worker pools, process spawning, filesystem access, network calls, shell execution, or background swarm orchestration.

## Validation

```bash
go test ./internal/productdata ./internal/runtime -run 'TestValidateAgentToolCallArguments|TestMemoryServiceAgentTaskLifecycle|TestAgentSpawnListAndComplete|TestToolCatalogIncludesAgentRuntimeTools|TestToolBrokerExecutesAgentSpawnThroughOneEntrypoint' -count=1
go test ./internal/productdata ./internal/runtime -count=1
```

Docs build should be run before closeout:

```bash
bun run --cwd docs-site build
```
