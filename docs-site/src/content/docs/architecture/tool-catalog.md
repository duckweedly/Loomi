---
title: Tool Catalog Visibility
description: M11 read-only visibility for Loomi runtime and workspace tools.
---

M11 adds a read-only tool catalog so the active code-agent surface is visible from both API and Settings. It does not change tool execution, approval, worker resume, or provider behavior.

## Boundary

The catalog is static in-process metadata derived from the current allowlisted tools:

- `runtime.get_current_time`
- `workspace.glob`
- `workspace.grep`
- `workspace.read_file`
- `workspace.write_file`
- `workspace.edit`
- `workspace.exec_command`

Each entry exposes name, label, group, capability, approval policy, safety class, risk level, side effect, enabled state, and description.

## Flow

1. Backend runtime owns the deterministic catalog in `internal/runtime`.
2. HTTP exposes `GET /v1/tools/catalog`.
3. The real frontend API client maps snake_case catalog fields into domain fields.
4. Mock mode returns the same tool names for local UI development.
5. Settings > Tools renders the catalog as read-only cards.

## Safety

The catalog never includes provider credentials, raw provider payloads, file content, command output, command examples, or editable permission state. Approval remains required for every listed tool.

## Non-Goals

M11 does not add permission editing, auto-approval, tool execution buttons, MCP, browser automation, sandbox sessions, or multi-agent delegation.
