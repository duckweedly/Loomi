# Quickstart: M15 Chat Real Integrated Smoke Closeout

## Backend Smoke

```bash
LOOMI_M15_REAL_CHAT_SMOKE=1 go test ./internal/httpapi -run TestM15ChatRealIntegratedSmoke -count=1 -v
```

Expected:

- deterministic provider first requests one discovered persona-allowed MCP tool
- run blocks on approval
- HTTP approve resumes worker execution
- worker executes one MCP `tools/call`
- redacted result enters one provider continuation
- final assistant message is persisted
- replayed events include memory, MCP, approval, execution, continuation, and completion milestones
- configured sensitive canaries are absent from shareable surfaces

## Required Validation Before Closeout

```bash
go test ./...
bun test --cwd web
bun run --cwd web build
bun run --cwd docs-site build
git diff --check
```

## Browser Smoke

If local API and web shell startup are available, open Chat in real API mode and verify the completed run timeline shows the same memory, MCP approval/execution, continuation, and completion states. If unavailable, record the blocker and use the backend smoke output as equivalent evidence.
