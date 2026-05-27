---
title: M29 Multi-agent Runtime Foundation
description: Approval-gated coordination-only agent task runtime architecture.
---

M29 adds the first multi-agent runtime slice as three builtin tools: `agent.spawn`, `agent.list`, and `agent.complete`.

The tools reuse the existing `ToolCatalog -> RunContext -> approval -> ToolBroker -> worker continuation` path. The runtime is coordination-only. It creates and updates bounded task records in the current thread; it does not start child model runs, spawn external processes, call networks, read or write files, execute shells, or delegate across threads.

## Boundaries

Agent tools are:

- builtin
- Work mode only
- approval required
- medium risk
- coordination-only
- no autonomous execution

Chat mode filters them out with workspace, sandbox, LSP, web, browser, and artifact tools. `agent.spawn` creates one task record with a supported role and bounded goal through `AgentTaskService`, backed by both in-memory service and PostgreSQL `agent_tasks` in the real API path. `agent.list` returns bounded summaries for the current thread. `agent.complete` marks one current-thread task complete with a bounded result summary.

## Execution

`AgentToolExecutor` depends on `productdata.AgentTaskService`. Worker approved-tool resume injects the configured productdata service or repository, then records tool success and continues the provider with a safe result summary.

Run events persist role, goal, status, task id, result summary, source thread, source run, and `autonomous_execution=false`. They do not include raw provider payloads, credentials, local paths, child execution logs, external process ids, or cross-thread delegation data.

## Visibility

Settings > Tools shows agent tools as agent-scoped, approval-required, medium risk, coordination-only, and no autonomous execution. RunRail labels agent lifecycle rows separately from workspace, sandbox, LSP, web fetch, browser, artifact, MCP, and runtime tools.
