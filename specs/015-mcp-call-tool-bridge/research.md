# Research: M13 MCP Call Tool Bridge

## Decision: Built-in Local MCP Pair First

The first MCP slice supports only `local.echo`. This proves the approval/run-event/worker/catalog shape without adding external process lifecycle, sockets, or third-party server configuration.

## Decision: Approval Required

All current Loomi tools begin blocked on approval. M13 preserves that rule because MCP calls can represent external side effects in later slices.

## Decision: Existing ToolCall Storage

Arguments and results stay in existing tool-call projections. Full MCP session state and server registry are future work.

## Decision: Secret-Looking Messages Rejected

The local echo tool rejects secret-looking strings before execution. Existing metadata redaction still applies during persistence.
