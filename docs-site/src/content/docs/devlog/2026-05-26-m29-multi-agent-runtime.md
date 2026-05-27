---
title: 2026-05-26 M29 Multi-agent Runtime Foundation
description: Implementation notes and validation for M29 agent coordination tools.
---

## Completed

- Added Spec Kit feature `specs/037-multi-agent-runtime-foundation/`.
- Added builtin `agent.spawn`, `agent.list`, and `agent.complete` catalog identity, default persona allowlist, Work-mode filtering, and safe tool-call metadata grouping.
- Added productdata agent task records, in-memory service methods, PostgreSQL `agent_tasks` migration, and PostgresRepository spawn/list/complete methods.
- Added `AgentToolExecutor` for coordination-only task creation, bounded list, and safe completion.
- Routed agent tools through ToolBroker, worker approved-tool resume, provider continuation, and HTTP smoke coverage.
- Added PG/in-memory alignment coverage for spawn/list/complete and cross-thread no-leak behavior.
- Updated Settings > Tools, RunRail labels, mock catalog, seeded run data, and runtime scripts for visible agent lifecycle metadata.

## Validation

Focused validation during implementation:

```bash
go test ./internal/productdata -run 'TestValidateAgentToolCallArguments|TestMemoryServiceAgentTaskLifecycle|TestToolCatalogIncludesAgentRuntimeTools|TestWorkModeScopedToolsOnlyEnabledForWorkModeRunContext'
go test ./internal/runtime -run 'TestAgentSpawnListAndComplete|TestAgentRejectsUnsafeInputsAndScope'
go test ./internal/runtime -run 'TestToolBrokerExecutesAgentSpawnThroughOneEntrypoint|TestWorkerExecutesApprovedAgentSpawnAndContinuesModel'
go test ./internal/runtime -run 'TestGatewayRejectsAgentToolInChatMode|TestWorkerDoesNotSpawnAgentTaskAfterStopOrDenied|TestAgentRejectsUnsafeInputsAndScope'
go test ./internal/httpapi -run 'TestM29Agent'
bun test web/src/components/SettingsView.tools.test.tsx web/src/components/RunRail.runtime.test.ts web/src/runtime/runtimeScripts.test.ts web/src/runtime/mockExecutionAdapter.test.ts
```

Full closeout commands should also run before marking M29 complete:

```bash
go test ./...
bun test --cwd web
bun run --cwd web build
bun run --cwd docs-site build
git diff --check
```

## Non-goals

No autonomous child model runs, cross-thread delegation, external worker pool, process spawning, filesystem access, network calls, shell execution, long-term multi-agent memory, marketplace packaging, or background swarm orchestration were added.
