---
title: M18 Tool Runtime Catalog
description: Tool catalog, broker, approval, and execution boundaries for builtin and MCP tools.
---

M18 makes tools first-class runtime objects without adding new powerful tools. The current executable surface remains `runtime.get_current_time` plus already-discovered local stdio MCP tools, but both now share one catalog and broker boundary.

## Catalog

The catalog is computed from builtin tool definitions and safe MCP discovery events. Each entry exposes only safe fields: name, display name, description, source, group, input schema hash, risk level, approval policy, enabled state, execution state, and safe metadata.

For display/API catalog summaries, MCP entries use the latest successful discovery event for the same namespaced tool and report `non_executable` when executor availability cannot be proven from productdata alone. For actual execution, the worker builds the broker catalog from the prepared RunContext for that run, so the schema hash is the current discovery projection rather than an older global catalog row.

M18 does not persist user policy overrides. Settings > Tools is read-only.

## Broker

Approved tool resume jobs call the broker before any concrete executor. The broker checks scoped thread/run/tool-call identity, approval status, execution status, catalog membership, enabled state, persona allowed tool resolution, and MCP candidate schema hash.

Only after those checks does the broker dispatch to the builtin current-time executor or the local stdio MCP executor. Provider and worker code should not call those executors directly.

## RunContext

RunContext tool resolution now uses the catalog plus persona allowlist plus MCP discovery metadata. Builtin tools appear when allowed by the persona. MCP tools appear only when namespaced, discovered, schema-hashed, and persona-allowed.

## Events

M18 keeps the M7/M12 lifecycle:

`tool_call_requested -> tool_call_approval_required -> tool_call_approved -> tool_call_executing -> tool_call_succeeded/failed`

Tool event metadata adds safe source/group/schema context. Raw args, raw result, MCP command/env/stderr, provider traces, and credential material remain outside events and API responses.

## Deferred

Workspace read tools, shell, sandbox execution, browser automation, web search/fetch, artifact runtime, plugin marketplace, remote MCP/OAuth, provider autodetect, multi-agent, and worker queue rewrite remain out of M18.
