# Contract: M7 Documentation Update Plan

Implementation of M7 must update docs-site in the same work session.

## Required docs-site targets

| Path | Required Content |
|------|------------------|
| `docs-site/src/content/docs/architecture/tool-call-approval.md` | Tool-call lifecycle, approval boundary, run_event audit model, minimal Tool Call projection, worker block/resume, non-goals |
| `docs-site/src/content/docs/api/tool-call-approval.md` | Approve/deny endpoints, tool-call projection shape, tool event payloads, idempotency, redaction, errors |
| `docs-site/src/content/docs/runbooks/local-m7.md` | Local setup, fake/provider tool-call smoke, approve/deny smoke, cancellation smoke, validation commands |
| `docs-site/src/content/docs/devlog/2026-05-24-m7-tool-call-approval.md` | Completed work, validation results, known limitations, deferred multi-step loop/tool categories |
| `docs-site/src/content/docs/roadmap/current-status.md` | Current status updated to include M7 and next boundaries |

## Cross-links

- Link M7 architecture page to existing LLM gateway and worker job pipeline pages.
- Link M7 API page to run/event SSE and LLM gateway API pages.
- Link runbook to local M5/M6 runbooks where provider and worker setup are reused.

## Required documented safety boundaries

- No shell/terminal tool.
- No file system read/write tool.
- No arbitrary network request tool.
- No MCP integration.
- No browser automation.
- No multi-agent execution.
- No long-term memory/RAG.
- No approval bypass for approval-required tools.
- No raw provider payloads, API keys, Authorization headers, secrets, shell output, file contents, or arbitrary URL contents in events/results/UI.
- Model-generated tool arguments are untrusted and require schema validation.

## Validation

When docs are changed during implementation, run:

```bash
bun run --cwd docs-site build
```

Record exact validation results or exact skipped-command reasons in the M7 devlog.
