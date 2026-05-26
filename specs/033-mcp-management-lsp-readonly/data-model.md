# Data Model: M25 MCP Management + LSP Read-only Foundation

## MCPServerStatus

Safe read-only status for a local stdio MCP server.

Fields:

- `server_safe_id`: stable safe id such as `mcp:local-search`
- `server_slug`: safe slug
- `display_name`: safe display name
- `transport`: `stdio`
- `enabled`: boolean
- `config_source`: `local`
- `discovery_status`: `not_configured`, `disabled`, `succeeded`, `failed`, or `rejected`
- `candidate_count`: bounded integer
- `candidate_names`: safe MCP tool names
- `execution_mode`: `approval_gated` when executable through current runtime, otherwise `disabled` or `unavailable`
- `redacted_error_code`: optional redacted error code
- `last_discovered_at`: optional timestamp

Validation:

- No command, args, env values, absolute private paths, raw payloads, or secrets.
- Candidate names must satisfy existing MCP tool-name validation.

## LSPToolCatalogEntry

Builtin read-only code intelligence tool.

Names:

- `lsp.diagnostics`
- `lsp.symbols`
- `lsp.references`

Shared catalog metadata:

- `source`: `builtin`
- `group`: `lsp`
- `risk_level`: `low`
- `approval_policy`: `always_required`
- `execution_state`: `executable`
- `safe_metadata.scope`: `lsp`
- `safe_metadata.read_only`: `true`

## LSPRequestSummary

Safe argument preview.

Fields:

- `path`: workspace-relative source file path or workspace-relative search root
- `query`: optional bounded symbol/reference query
- `line`: optional 1-based line for position-based tools
- `column`: optional 1-based column for position-based tools
- `include_declaration`: optional boolean for references
- `language`: optional safe language hint
- `limit`: bounded integer
- `redaction_applied`: boolean

## LSPResultSummary

Safe bounded result.

Fields:

- `tool`: LSP tool name
- `scope`: `lsp`
- `operation`: `diagnostics`, `symbols`, or `references`
- `path`: workspace-relative path
- `items`: bounded list of diagnostics/symbol/reference summaries
- `count`: returned count
- `truncated`: boolean
- `redaction_applied`: boolean

Validation:

- All paths are workspace-relative.
- Result text is UTF-8 safe and bounded.
- Sensitive path and secret-looking values are redacted or rejected.
