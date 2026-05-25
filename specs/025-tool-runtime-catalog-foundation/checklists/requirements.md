# Requirements Checklist: M18 Tool Runtime + Tool Catalog Foundation

- [x] Scope excludes workspace, shell, sandbox, browser, web, artifact runtime, plugin marketplace, remote MCP/OAuth, autodetect, multi-agent, and worker queue rewrite.
- [x] Catalog fields cover name, display, description, source, group, schema hash, risk, approval policy, enabled state, execution state, and safe metadata.
- [x] Broker checks approval, tool name, schema hash, scope, persona allowlist, enabled state, and execution state.
- [x] Existing M7/M12 lifecycle events remain the projection boundary.
- [x] API and UI are read-only and safe.
- [x] Validation commands are listed.
