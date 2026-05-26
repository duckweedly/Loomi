---
title: Local M29 Multi-agent Runtime Validation
description: Commands for validating the M29 multi-agent runtime foundation locally.
---

## Focused Validation

```bash
go test ./internal/productdata -run 'TestValidateAgentToolCallArguments|TestMemoryServiceAgentTaskLifecycle|TestToolCatalogIncludesAgentRuntimeTools|TestWorkModeScopedToolsOnlyEnabledForWorkModeRunContext'
LOOMI_TEST_DATABASE_URL="$DATABASE_URL" go test ./internal/productdata -run TestPostgresArtifactsAndAgentTasksUseThreadScope -count=1 -v
go test ./internal/runtime -run 'TestAgent|TestToolBrokerExecutesAgentSpawnThroughOneEntrypoint|TestWorkerExecutesApprovedAgentSpawnAndContinuesModel|TestWorkerDoesNotSpawnAgentTaskAfterStopOrDenied|TestGatewayRejectsAgentToolInChatMode'
go test ./internal/httpapi -run 'TestM29Agent'
bun test --cwd web SettingsView.tools.test.tsx RunRail.runtime.test.ts runtimeScripts.test.ts mockExecutionAdapter.test.ts
```

## Full Closeout

```bash
go test ./...
bun test --cwd web
bun run --cwd web build
bun run --cwd docs-site build
git diff --check
```

## Manual Smoke

1. Start the API and web app.
2. Open Settings > Tools and confirm `agent.spawn`, `agent.list`, and `agent.complete` render as builtin, agent-scoped, approval-required, medium risk, coordination-only, no autonomous execution, and backed by real API storage when PostgreSQL is configured.
3. Open RunRail and confirm agent lifecycle rows are visible.
4. Confirm the UI shows task id, role, goal, status, result summary, and `autonomous_execution=false` without raw provider payloads, credentials, local paths, child execution logs, or external process ids.

M29 does not support autonomous child runs, cross-thread delegation, external worker pools, process spawning, filesystem access, network calls, shell execution, long-term multi-agent memory, marketplace packaging, or background swarm orchestration.
