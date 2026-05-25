# Contract: Future MCP Execution Boundary

## Purpose

Document the boundary that must exist before any discovered MCP tool can execute. M11 foundation does not implement execution.

## Non-Executable State

For this slice, every MCP ToolSpec candidate has:

```json
{
  "source": "mcp",
  "execution_enabled": false,
  "approval_policy": "future_always_required",
  "auto_execute": false
}
```

## Future Execution Requirements

Before any MCP tool invocation is allowed, a later spec must define and implement:

- M7-style user approval before execution
- persisted tool-call projection scoped to run/thread/user
- audit events for request, approval, denial, execution start, success, failure, and cancellation
- redacted argument summaries and hashes
- redacted result/error summaries
- worker ownership and cancellation guards
- retry/idempotency semantics
- explicit safety class for the MCP tool
- no automatic execution from model output, persona allowlist, discovery metadata, or server description

## Required Rejections

The execution boundary must reject:

- unapproved MCP calls
- calls from disabled candidates
- calls where server config changed since approval without a new approval
- calls that would expose raw env, args, stderr, tokens, credentials, secret paths, file contents, shell output, or browser/desktop captured state
- calls that try to bypass M7 approval or write outside the audited tool-call lifecycle
