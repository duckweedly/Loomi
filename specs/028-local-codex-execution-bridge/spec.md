# Feature Specification: Local Codex Execution Bridge

**Feature Branch**: `028-local-codex-execution-bridge`

**Created**: 2026-05-25

**Status**: Complete candidate

**Input**: User description: "M20 Local Codex Execution Bridge: after explicit detect and enable, make Local Codex a real executable provider for Chat through the existing model gateway, run events, SSE, worker, and timeline path, without secret leakage or automatic local auth scanning."

## User Scenarios & Testing

### User Story 1 - Enable Local Codex as an executable provider (Priority: P1)

A user detects Local Codex in Settings, explicitly enables it for the session, and sees `local_codex` returned as available and supported in configured providers.

**Why this priority**: Chat cannot use Local Codex until the provider capability is honest and executable.

**Independent Test**: A backend fixture with temporary `CODEX_HOME` detects Local Codex, enables it, and verifies `GET /v1/model-providers` returns `local_codex` with `status=available` and `execution_state=supported`.

**Acceptance Scenarios**:

1. **Given** Local Codex is detected as available, **When** the user explicitly enables it, **Then** Loomi stores only session-local redacted enablement and returns a supported provider capability.
2. **Given** Local Codex is detected but disabled, **When** model providers are listed, **Then** `local_codex` is absent.
3. **Given** Local Codex auth is unavailable or no executable bridge can be built, **When** model providers are listed after enable attempt, **Then** Chat remains blocked with a clear unavailable status.

### User Story 2 - Send Chat through the existing run pipeline (Priority: P1)

A user sends a Chat message with enabled Local Codex and receives an assistant reply generated through the existing model gateway, worker, run events, and SSE timeline path.

**Why this priority**: M20 must not create a separate chat path; observability and future tool integration depend on the existing run pipeline.

**Independent Test**: A Chat HTTP smoke creates a thread, message, and model gateway run for `local_codex`, processes the worker, and verifies final assistant message plus `model_request_started`, `model_output_delta`, and `run_completed` events.

**Acceptance Scenarios**:

1. **Given** `local_codex` is available and supported, **When** Chat starts a model gateway run, **Then** the run is accepted and processed by the existing worker/gateway route.
2. **Given** the Local Codex execution bridge receives a provider failure, **When** the run is processed, **Then** the run records a provider failure instead of fabricating assistant output.
3. **Given** a normal Local Codex response streams deltas, **When** the run finishes, **Then** RunTimeline and RunRail can show model request, output, and completion events.

### User Story 3 - Preserve safety and frontend blocking semantics (Priority: P1)

A user never sees tokens, private auth paths, or authorization material in API responses, run events, assistant metadata, frontend state, docs, or logs; unsupported or unavailable Local Codex states still block sending.

**Why this priority**: Local desktop auth material is sensitive and must not leak while making the route executable.

**Independent Test**: Go and web tests inject canary `access_token`, `refresh_token`, auth path, API key, and Authorization values, then assert they do not appear in responses, run events, assistant metadata, mapped frontend state, or captured test logs.

**Acceptance Scenarios**:

1. **Given** Local Codex auth contains token canaries, **When** detection, enable, list, run, and event APIs respond, **Then** no secret or auth path appears.
2. **Given** Local Codex is enabled but unavailable, **When** Chat renders, **Then** Composer shows "Local Codex 登录态不可用，请重新检测或配置 OpenAI-compatible provider" and sending is disabled.
3. **Given** Local Codex is enabled but execution is unsupported, **When** Chat renders, **Then** Composer shows "Local Codex 已启用，但暂不支持执行" and sending is disabled.
4. **Given** Local Codex is available and supported, **When** Chat renders, **Then** no provider unavailable warning blocks Composer.

### Edge Cases

- Auth material disappears between enable and execution: provider list becomes unavailable on refresh or execution records provider failure without exposing paths or tokens.
- Auth JSON is malformed: detection/enable fail as unavailable and never echo raw file content.
- Concurrent enable, disable, list, save, and check requests: provider state remains race-free and returns safe capabilities.
- OpenAI-compatible provider save/check continues to work while Local Codex is enabled.
- Unsupported local providers remain rejected by the HTTP run handler before a run starts.

## Requirements

### Functional Requirements

- **FR-001**: System MUST execute Local Codex only after explicit detect and enable actions.
- **FR-002**: System MUST NOT automatically scan local auth during startup, page mount, provider list, Chat send, or unrelated provider checks.
- **FR-003**: System MUST expose enabled executable Local Codex as `status=available` and `execution_state=supported` in `GET /v1/model-providers`.
- **FR-004**: System MUST omit detected-but-disabled Local Codex from `GET /v1/model-providers`.
- **FR-005**: System MUST keep enabled-but-unavailable Local Codex blocked with a clear unavailable status and user-facing message.
- **FR-006**: System MUST reject unsupported local provider runs directly in the HTTP start-run handler.
- **FR-007**: System MUST route supported `local_codex` runs through the existing runtime `Provider` interface, Gateway, worker, persisted run events, SSE, RunTimeline, and RunRail path.
- **FR-008**: System MUST record normal Local Codex executions as `model_request_started`, `model_output_delta`, and `run_completed`, or clear provider failure events.
- **FR-009**: System MUST NOT store or return access tokens, refresh tokens, API keys, Authorization headers, auth file paths, private home paths, or raw auth JSON in DB, run events, assistant metadata, docs, frontend state, logs, or API responses.
- **FR-010**: System MUST NOT modify `~/.codex/auth.json`, refresh OAuth, read keychain, install CLI, invoke interactive CLI, or introduce sandbox/browser/filesystem/shell/workspace tools.
- **FR-011**: System MUST keep `localProviderEnablements` and provider config access safe under concurrent enable, disable, list, save, and check operations.
- **FR-012**: System MUST preserve OpenAI-compatible provider save/check behavior.
- **FR-013**: Web MUST map Local Codex provider capability without secret/path fields and must use Local Codex-specific warning copy for unsupported and unavailable states.
- **FR-014**: Settings MUST show detect, enable, disable, available/supported, unavailable, and session-local redacted states accurately.

### Key Entities

- **LocalCodexCredentialSnapshot**: In-memory execution-time credential material derived from an explicit enable or execution boundary; never serialized.
- **LocalCodexProvider**: Runtime provider implementation that adapts Local Codex auth material into existing provider events.
- **LocalProviderEnablement**: Session-local redacted provider state used for capability listing and routing.
- **ProviderCapability**: Safe API/frontend view of provider readiness and execution support.
- **RunEvent**: Persisted observable execution events with redacted metadata only.

## Success Criteria

### Measurable Outcomes

- **SC-001**: Fixture-backed Go tests prove detect plus enable returns Local Codex available/supported and starts a model gateway run.
- **SC-002**: Chat HTTP smoke proves create thread/message/run, worker processing, final assistant message, and timeline events all work for `local_codex`.
- **SC-003**: Redaction tests prove token/key/path canaries are absent from API responses, run events, assistant metadata, and test logs.
- **SC-004**: Web tests prove available/supported Local Codex does not block Composer, while unsupported/unavailable states show Local Codex-specific disabled copy.
- **SC-005**: Race/concurrency tests cover enable, disable, list, save, and check without data races.
- **SC-006**: Existing OpenAI-compatible provider save/check tests pass.

## Assumptions

- M20 chooses an auth.json direct bridge rather than CLI execution unless research proves CLI is safer, more testable, and non-interactive.
- The first executable slice may be fixture-backed if real local OAuth endpoint compatibility cannot be proven without leaking or refreshing credentials.
- Local Codex enablement remains process-local/session-local; durable credential references are out of scope.
