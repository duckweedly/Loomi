---
description: "M27 Browser Automation Foundation feature specification"
---

# Feature Specification: M27 Browser Automation Foundation

**Feature Branch**: `[035-browser-automation-foundation]`

**Created**: 2026-05-26

**Status**: Draft

**Input**: User description: "Continue Arkloop-level coverage after M26 web.fetch. Add the first browser automation runtime slice through the existing catalog/broker/approval boundary. Keep it Work-mode-only, approval-gated, observable, bounded, public-network only, and do not add logged-in browser profiles, cookies, JavaScript execution, screenshots, form submission, downloads, desktop activity recording, artifact runtime, or multi-agent orchestration yet."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Open and Snapshot a Public Page (Priority: P1)

As a Work mode user, I want the agent to open a public page in a bounded browser session and request a safe text/link snapshot, so Loomi can inspect page state beyond a one-off fetch.

**Why this priority**: M26 proved one public HTTP(S) read. Browser automation needs a stateful session boundary before click/navigation actions, JavaScript rendering, or profile-backed browsing.

**Independent Test**: A Work mode run requests `browser.open`, the user approves it, Loomi creates one bounded public page session and returns a safe snapshot with session id, URL, title, text excerpt, links, bytes, and truncation metadata.

**Acceptance Scenarios**:

1. **Given** a Work mode run and an allowed public URL, **When** `browser.open` is approved, **Then** Loomi stores a bounded browser session for that run and returns a safe snapshot.
2. **Given** an existing browser session, **When** `browser.snapshot` is approved, **Then** Loomi returns the current safe page state without issuing a new navigation.
3. **Given** a response is oversized, **When** the snapshot is produced, **Then** Loomi truncates text/links and marks truncation metadata.

---

### User Story 2 - Navigate by Approved Link Click (Priority: P2)

As a Work mode user, I want the agent to click a known safe link in the current browser session, so it can follow documentation pages with explicit approval and visible navigation history.

**Why this priority**: Link click proves browser automation semantics while keeping the first slice bounded and auditable.

**Independent Test**: Open a page with links, approve `browser.click_link` for one link index, and verify the session navigates to the selected public HTTP(S) target and returns the new safe snapshot.

**Acceptance Scenarios**:

1. **Given** a browser session with safe links, **When** `browser.click_link` is approved for a link index, **Then** Loomi navigates once and records previous URL, final URL, link text, status, and snapshot metadata.
2. **Given** a link points to a blocked scheme or private/local host, **When** `browser.click_link` validates the target, **Then** Loomi rejects it before network execution.
3. **Given** an unknown or expired session id, **When** `browser.snapshot` or `browser.click_link` runs, **Then** Loomi returns a safe error without creating a new session.

---

### User Story 3 - Keep Browser Automation Safe and Visible (Priority: P3)

As an operator, I want browser tools to be clearly separated from web fetch, sandbox, workspace, MCP, and LSP tools, so I can audit network/page access and know what automation did.

**Why this priority**: Browser automation is higher risk than fetch because it creates navigable state. Settings and RunRail must make that state and boundary visible.

**Independent Test**: Settings Tools and RunRail render browser tool scope/risk/approval lifecycle metadata without raw HTML, cookies, local paths, secrets, or hidden profile data.

**Acceptance Scenarios**:

1. **Given** the tool catalog is shown, **When** browser tools are present, **Then** they are marked builtin, browser-scoped, approval-required, executable, public HTTP only, and medium risk.
2. **Given** Chat mode asks for a browser tool, **When** the gateway validates tools, **Then** Loomi rejects it before approval or execution.
3. **Given** browser lifecycle events are replayed, **When** RunRail renders them, **Then** Loomi shows browser automation labels and safe URL/title/status/session metadata only.

### Edge Cases

- Missing URL, invalid URL, unsupported scheme, credentialed URL, private/local/link-local host, blocked redirects, unsupported content, oversized response, timeout, unknown session, out-of-range link index, duplicate tool-call ID, denied/stopped/terminal run, and Chat mode.
- This slice does not execute JavaScript, keep cookies, reuse user browser profiles, submit forms, type into inputs, download files, capture screenshots, read authenticated pages, run extensions, or control desktop browsers.
- Snapshot text and link lists must be UTF-8 safe and bounded.
- Events must not persist raw HTML, full page bodies, cookies, Authorization values, Set-Cookie values, local absolute paths, raw provider payloads, or secret-looking content.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Loomi MUST add builtin browser tool catalog entries for `browser.open`, `browser.snapshot`, and `browser.click_link`.
- **FR-002**: Browser tools MUST be available only for Work mode runs and only through the existing RunContext enabled-tool snapshot.
- **FR-003**: Browser tools MUST require approval and execute only through ToolBroker/worker resume.
- **FR-004**: `browser.open` MUST create a bounded run-scoped browser session from an absolute public HTTP(S) URL without credentials.
- **FR-005**: `browser.snapshot` MUST return the current safe state for an existing session without navigating.
- **FR-006**: `browser.click_link` MUST navigate exactly one safe link from an existing session by approved link index.
- **FR-007**: Browser tools MUST reject localhost, loopback, link-local, private, multicast, unspecified, and blocked private-network hosts before network execution.
- **FR-008**: Browser tools MUST validate redirects and clicked link targets before reading target response bodies.
- **FR-009**: Browser snapshots MUST include safe session id, URL, final URL, status code, content type, title, text excerpt, bounded link summaries, byte count, byte limit, and truncation flag.
- **FR-010**: Settings Tools and RunRail MUST distinguish browser automation from web fetch, workspace, sandbox, MCP, LSP, and runtime tools.
- **FR-011**: Normal events and UI MUST NOT persist or display full HTML/body content, cookies, credentials, local absolute paths, Authorization values, Set-Cookie values, raw provider payloads, or secret-looking content.
- **FR-012**: This feature MUST NOT add logged-in browser profile access, cookies, JavaScript rendering, screenshots, form submission, downloads, artifact runtime, channels, desktop activity recording, plugin marketplace, or multi-agent orchestration.

### Key Entities *(include if feature involves data)*

- **Browser Tool**: Builtin browser automation tool with `source=builtin`, `group=browser`, `risk_level=medium`, `approval_policy=always_required`, `execution_state=executable`, and public HTTP-only metadata.
- **Browser Session**: Run-scoped state containing safe session id, current URL, final URL, status, content type, title, text excerpt, bounded links, byte and truncation metadata, and update time.
- **Browser Link Summary**: Safe link index, text, href, host, and blocked flag derived from the current snapshot.
- **Browser Result Summary**: Safe output for open/snapshot/click including operation, session id, URL metadata, status, title, text excerpt, links, and truncation metadata.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A Work mode smoke can approve `browser.open` and observe exactly one browser session creation followed by provider continuation.
- **SC-002**: A Work mode smoke can approve `browser.open -> browser.click_link -> browser.snapshot` within the bounded tool loop and observe safe page state updates.
- **SC-003**: Safety tests prove Chat-mode, unsafe schemes, credentialed URLs, localhost/private/link-local hosts, blocked redirects, unknown sessions, denied/stopped/terminal states, and unsupported browser requests do not execute network reads.
- **SC-004**: Settings Tools and RunRail expose browser scope/risk/public HTTP-only lifecycle metadata without leaking raw HTML, cookies, credentials, or local paths.
- **SC-005**: The feature passes backend tests, web tests/build, docs build, diff check, and browser smoke for visible browser tool states.

## Assumptions

- The first browser foundation may use deterministic HTTP-backed page state and link navigation rather than a full browser engine.
- Production defaults reject private/local hosts; tests may inject a test-only executor configuration for local `httptest` servers.
- JavaScript rendering, user-profile browser automation, screenshots, cookies, authenticated browsing, and artifact capture are separate later features.
