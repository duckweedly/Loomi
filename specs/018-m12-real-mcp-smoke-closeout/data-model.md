# Data Model: M12 Real MCP Smoke Closeout

## M12.5 Smoke Run

Represents one local model-gateway run used as closeout evidence.

Fields:

- `run_id`: Existing run identifier.
- `thread_id`: Existing thread identifier.
- `message_id`: Existing user message identifier.
- `provider_id`: Existing provider route, using a deterministic local test provider.
- `tool_call_id`: Stable MCP tool-call id used across request, approval, execution, and continuation.
- `event_sequence`: Persisted run events in order.

Validation:

- Must include discovery, approval-required, approved, executing, succeeded, continuation, and final states.
- Must not contain raw fixture secrets, env values, private paths, raw stderr, or raw command details.

## Local MCP Fixture

Represents a local stdio process used by the smoke.

Fields:

- `server_slug`: Safe local MCP slug.
- `display_name`: Safe display name.
- `command`: Test subprocess command.
- `args`: Test subprocess arguments.
- `env`: Fixture mode flags only.
- `tools_call_count`: Evidence counter for approved execution.

Validation:

- Must emit MCP `Content-Length` frames.
- Must answer `tools/list` during discovery.
- Must answer exactly one `tools/call` after approval.
- Must fail the smoke if `tools/call` happens before approval or more than once.

## Closeout Evidence

Represents the docs-site record of M12.5.

Fields:

- `smoke_scope`: What was validated.
- `validation_commands`: Exact commands run.
- `browser_smoke_status`: Passed or skipped with reason.
- `known_limitations`: Explicit non-goals and environment notes.

Validation:

- Must mention that broader MCP platform capabilities remain out of scope.
- Must build with the docs site.
