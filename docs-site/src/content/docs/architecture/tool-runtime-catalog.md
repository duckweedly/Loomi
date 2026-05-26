---
title: M18 Tool Runtime Catalog
description: Tool catalog, broker, approval, and execution boundaries for builtin and MCP tools.
---

M18 makes tools first-class runtime objects without adding new powerful tools. Later slices extend the same catalog and broker boundary with discovery, workspace, bounded command, LSP, web, browser, artifact, coordination, and Work todo tools.

## Catalog

The catalog is computed from builtin tool definitions and safe MCP discovery events. Each entry exposes only safe fields: name, display name, description, source, group, input schema hash, risk level, approval policy, enabled state, execution state, and safe metadata.

For display/API catalog summaries, MCP entries use the latest successful discovery event for the same namespaced tool and report `non_executable` when executor availability cannot be proven from productdata alone. For actual execution, the worker builds the broker catalog from the prepared RunContext for that run, so the schema hash is the current discovery projection rather than an older global catalog row.

M18 does not persist user policy overrides. Settings > Tools is read-only.

The discovery thin slice adds `tool.load_tools` and `skill.load_skill` as safe builtin tools. They are auto-approved read-only helpers:

- `tool.load_tools` returns safe descriptions and metadata for tools already enabled in the current run.
- `skill.load_skill` returns installed skill manifest summaries by name, but does not return full `SKILL.md` instruction bodies.
- The current slice does not dynamically inject new provider schemas after the call; provider schemas still come from the run's enabled-tool snapshot.

## Broker

Approved tool resume jobs call the broker before any concrete executor. The broker checks scoped thread/run/tool-call identity, approval status, execution status, catalog membership, enabled state, persona allowed tool resolution, and MCP candidate schema hash.

Only after those checks does the broker dispatch to the concrete executor. Builtin executors currently include current time, discovery, workspace read/mutation, bounded command, LSP, web, browser, artifact, coordination task, and `todo.write`; MCP tools dispatch through the local stdio MCP executor. Provider and worker code should not call concrete executors directly.

## RunContext

RunContext tool resolution now uses the catalog plus persona allowlist plus MCP discovery metadata. Builtin tools appear when allowed by the persona. Discovery tools are available in Chat and Work when persona-allowed. Work-scoped tools, including workspace, bounded command, LSP, web fetch, browser, artifact, coordination task, and `todo.write`, are filtered out unless the thread is in Work mode. `web.search` remains available to Chat when configured and allowed. MCP tools appear only when namespaced, discovered, schema-hashed, and persona-allowed.

## Events

M18 keeps the M7/M12 lifecycle:

`tool_call_requested -> tool_call_approval_required -> tool_call_approved -> tool_call_executing -> tool_call_succeeded/failed`

Tool event metadata adds safe source/group/schema context. Raw args, raw result, MCP command/env/stderr, provider traces, and credential material remain outside events and API responses.

`todo.write` is approval-gated like other tools. After worker execution, Loomi appends a durable `work.todo.updated` event with normalized `todo_items`, `updated_by=provider`, and redaction metadata. This lets WorkPlanView replay the model-maintained plan from run events instead of relying only on runtime-derived tool lifecycle snapshots.

## Deferred

True dynamic tool-schema injection after `tool.load_tools`, full skill instruction-body loading, persistent shell/PTTY, container/syscall sandboxing, authenticated browser automation, binary artifact execution, remote MCP/OAuth, plugin marketplace, autonomous child model runs, and worker queue rewrite remain out of the current tool runtime.
