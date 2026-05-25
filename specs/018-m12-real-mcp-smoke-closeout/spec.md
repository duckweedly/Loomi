# Feature Specification: M12 Real MCP Smoke Closeout

**Feature Branch**: `018-m12-real-mcp-smoke-closeout`

**Created**: 2026-05-25

**Status**: Draft

**Input**: User description: "补齐 M12 MCP approval-gated execution 的真实本地验证证据，作为 M12.5 closeout。只做真实 smoke/evidence closeout，不扩展 remote MCP / marketplace / plugin install / sandbox / shell/filesystem/browser automation / 多工具循环。"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Prove Local MCP Approval Execution (Priority: P1)

As a Loomi maintainer closing M12.5, I want one real local stdio MCP smoke that covers discovery, approval wait, approval decision, execution, redacted result, and provider continuation, so that M12 can be marked complete with executable evidence instead of only unit-level assertions.

**Why this priority**: M12 is a safety boundary. The closeout is not credible until the whole local path is exercised once.

**Independent Test**: Run the M12.5 backend smoke and verify the persisted event sequence shows discovery metadata with candidate schema hashes, an approval-required MCP tool call, a successful HTTP approve decision, one stdio `tools/call`, a redacted result, and one continuation/final message.

**Acceptance Scenarios**:

1. **Given** a real local stdio MCP fixture configured through local MCP server JSON, **When** discovery runs, **Then** Loomi records a discovered namespaced candidate with `candidate_schema_hashes`.
2. **Given** a selected persona allows that candidate, **When** the provider requests the MCP tool, **Then** Loomi records `tool_call_approval_required` and does not execute the fixture before approval.
3. **Given** the pending call is approved through the scoped approval API, **When** the worker resumes, **Then** it executes exactly one `tools/call`, records a redacted `tool_call_succeeded`, sends the redacted tool result to provider continuation, and completes with one final assistant message.

---

### User Story 2 - Document Closeout Evidence (Priority: P2)

As a future maintainer, I want docs-site pages to show the exact smoke scope, evidence, validation commands, and remaining non-goals, so that later MCP work does not accidentally treat M12.5 as a broader platform launch.

**Why this priority**: Documentation is part of done and prevents scope drift into remote MCP, plugin install, sandboxing, automation, or multi-tool loops.

**Independent Test**: Build the docs site and inspect the M12 runbook/devlog/status/workflow pages for the M12.5 evidence and explicit non-goals.

**Acceptance Scenarios**:

1. **Given** the smoke passes, **When** docs are updated, **Then** the runbook and devlog name the evidence chain and validation commands.
2. **Given** readers review roadmap/status and Spec Kit workflow, **When** they look for current MCP state, **Then** they see M12.5 as a closeout evidence slice, not a new platform capability.

### Edge Cases

- The browser smoke cannot run because the local environment lacks a live API/database/provider fixture; in that case backend/httpapi/runtime smoke must cover the same chain and the docs must state the browser limitation.
- The stdio fixture emits sensitive-looking output; the smoke must fail if persisted events, projection output, continuation messages, or docs examples contain raw secrets/private paths.
- The provider continuation asks for another tool; that remains out of scope and must not be validated as supported.
- The fixture receives more than one `tools/call`; the smoke must fail.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The closeout MUST add a real local stdio MCP fixture that uses `Content-Length` frames for both `tools/list` discovery and approved `tools/call` execution.
- **FR-002**: The fixture MUST support discovery before approval and exactly one execution after approval.
- **FR-003**: The smoke MUST configure the worker's real `StdioMCPToolExecutor` through `LOOMI_MCP_SERVERS_JSON`, not through a fake executor.
- **FR-004**: The smoke MUST prove discovery metadata includes candidate names and `candidate_schema_hashes`.
- **FR-005**: The smoke MUST prove a persona allowed-tool snapshot gates the provider-requested MCP call.
- **FR-006**: The smoke MUST prove the provider-requested MCP call blocks at `tool_call_approval_required` before execution.
- **FR-007**: The smoke MUST approve through the scoped HTTP approval path before worker execution.
- **FR-008**: The smoke MUST prove the worker records `tool_call_executing`, exactly one redacted `tool_call_succeeded`, provider continuation, and final completion.
- **FR-009**: The closeout MUST update docs-site runbook, devlog, roadmap current status, and Spec Kit workflow pages.
- **FR-010**: The closeout MUST keep remote MCP, HTTP/SSE/OAuth MCP, marketplace/plugin install, sandboxing, shell/filesystem/browser automation, and multi-tool loops out of scope.

### Key Entities *(include if feature involves data)*

- **M12.5 Smoke Run**: A local test run with discovery, approval, execution, continuation, final message, and redaction evidence.
- **Local MCP Fixture**: A stdio process fixture that emits MCP `Content-Length` frames for `tools/list` and `tools/call`.
- **Closeout Evidence**: Documentation record of command results, event sequence, browser limitation, and remaining non-goals.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A single backend smoke validates at least these event states in order: discovery succeeded, approval required, approved, executing, succeeded, continuation output, run completed.
- **SC-002**: The smoke fails if `tools/call` executes zero times or more than once.
- **SC-003**: The smoke fails if sensitive fixture strings appear in persisted result metadata or continuation request content.
- **SC-004**: Required validation commands complete or report exact blocking reasons: `go test ./...`, `bun test --cwd web`, `bun run --cwd web build`, `bun run --cwd docs-site build`, and `git diff --check`.

## Assumptions

- The current M12 implementation already provides approval projection, worker resume, stdio execution, and continuation primitives.
- Browser smoke is optional because it requires the local app stack and provider fixture to be runnable in the current environment.
- The closeout may add tests and documentation, but it must not add new user-facing MCP platform capabilities.
