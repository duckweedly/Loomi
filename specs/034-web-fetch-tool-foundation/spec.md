---
description: "M26 Web Fetch Tool Foundation feature specification"
---

# Feature Specification: M26 Web Fetch Tool Foundation

**Feature Branch**: `[034-web-fetch-tool-foundation]`

**Created**: 2026-05-26

**Status**: Draft

**Input**: User description: "Continue Arkloop-level code-agent coverage after M25 MCP/LSP. Add the first web runtime slice through the existing catalog/broker boundary. Keep it read-only, bounded, available to Chat and Work through persona allowlists, observable, and do not add browser automation, JavaScript rendering, crawler behavior, authenticated fetch, artifact runtime, channels, desktop activity recording, or multi-agent orchestration yet."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Fetch a Public Web Page Safely (Priority: P1)

As a Chat or Work user, I want the agent to request a bounded `web.fetch` tool for an explicit public HTTP(S) URL, so Loomi can inspect external reference material without shelling out or opening browser automation.

**Why this priority**: M21-M25 covered workspace read/write, sandbox exec, MCP, and LSP. The next platform surface should prove bounded public network read access before richer browser or artifact runtimes.

**Independent Test**: Start a Chat or Work run whose provider requests `web.fetch` for a public URL and verify Loomi auto-approves exactly one bounded HTTP(S) fetch, records safe request/result events, and continues the provider with a summarized result.

**Acceptance Scenarios**:

1. **Given** a run and a provider `web.fetch` request for an allowed `https://` URL, **When** Loomi records the request, **Then** it auto-approves the bounded read and returns status, final URL, content type, title, text excerpt, bytes read, and truncation metadata.
2. **Given** the response exceeds the byte limit, **When** `web.fetch` completes, **Then** Loomi truncates the stored result and marks `truncated=true`.
3. **Given** the response is not text-like, **When** `web.fetch` completes, **Then** Loomi returns a safe unsupported-content result without storing binary body bytes.

---

### User Story 2 - Keep Web Fetch Out of Private Networks (Priority: P2)

As an operator, I want web fetch to reject private, local, credentialed, or non-HTTP URLs, so a model cannot use Loomi as an SSRF or local-network probing tool.

**Why this priority**: Network reads have broader risk than local read-only tools. URL validation must be proven before any browser/crawler expansion.

**Independent Test**: Submit unsafe `web.fetch` requests through validation and worker paths and verify they fail before network execution.

**Acceptance Scenarios**:

1. **Given** a Chat mode run asks for `web.fetch` for a public HTTP(S) URL, **When** the gateway validates tools, **Then** the request can enter the same read-only auto-approved path as Work mode.
2. **Given** a URL with `file:`, `data:`, `ftp:`, credentials, localhost, loopback, link-local, or private IP host, **When** `web.fetch` validates arguments, **Then** Loomi rejects it before issuing an HTTP request.
3. **Given** a redirect targets a blocked host or scheme, **When** `web.fetch` follows redirects, **Then** Loomi stops before reading the redirected response body.

---

### User Story 3 - Make Web Fetch Visible in Settings and RunRail (Priority: P3)

As a user reviewing a run, I want Settings and RunRail to show that `web.fetch` is read-only, network-scoped, auto-approved, and bounded, so I can audit what network access occurred.

**Why this priority**: New tool surfaces must be visible and distinguishable from workspace, sandbox, MCP, and LSP tools.

**Independent Test**: Render Settings > Tools and a RunRail tool lifecycle containing `web.fetch`, and verify safe labels and metadata appear without raw response body, secrets, or local paths.

**Acceptance Scenarios**:

1. **Given** the tool catalog is shown, **When** `web.fetch` is present, **Then** it is marked builtin, web-scoped, read-only, auto-approved, executable, and medium risk.
2. **Given** a `web.fetch` lifecycle is replayed, **When** RunRail renders it, **Then** Loomi shows URL host/status/truncation metadata without raw body content.
3. **Given** web fetch fails validation or network execution, **When** the run is replayed, **Then** the failure is visible as a tool error with redacted metadata.

### Edge Cases

- Missing URL, invalid URL, relative URL, unsupported scheme, URL credentials, private hosts, local hosts, blocked redirects, unsupported content types, timeout, too-large body, and failed DNS/HTTP requests.
- The first slice does not crawl links, render JavaScript, preserve cookies, use browser profiles, submit forms, download binary files, authenticate to websites, or perform search-provider queries.
- Response excerpts must be UTF-8 safe and bounded.
- Tool request/result events must not persist full response bodies, cookies, Authorization headers, Set-Cookie values, raw provider payloads, or secret-looking content.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Loomi MUST add a builtin `web.fetch` tool catalog entry.
- **FR-002**: `web.fetch` MUST be available to Chat and Work runs through the existing RunContext enabled-tool snapshot when the persona allowlist permits it.
- **FR-003**: `web.fetch` MUST be auto-approved as a bounded public read and execute only through ToolBroker/worker resume.
- **FR-004**: `web.fetch` MUST accept only absolute `http://` or `https://` URLs without credentials.
- **FR-005**: `web.fetch` MUST reject localhost, loopback, link-local, private, multicast, unspecified, and otherwise blocked private-network hosts before network execution.
- **FR-006**: `web.fetch` MUST validate redirects before reading redirected response bodies.
- **FR-007**: `web.fetch` MUST enforce bounded timeout, response byte limit, redirect limit, and text-like content handling.
- **FR-008**: `web.fetch` result metadata MUST include safe URL, final URL, status, content type, title, excerpt, bytes read, byte limit, truncation flag, and redaction flag.
- **FR-009**: Settings Tools and RunRail MUST distinguish web/network read tools from workspace, sandbox, MCP, LSP, and runtime tools.
- **FR-010**: Normal events and UI MUST NOT persist or display full response bodies, cookies, credentials, local absolute paths, Authorization values, Set-Cookie values, raw provider payloads, or secret-looking content.
- **FR-011**: This feature MUST NOT add browser automation, JavaScript rendering, crawler behavior, web search provider integration, authenticated fetch, artifact runtime, channels, heartbeat, desktop activity recording, plugin marketplace, or multi-agent orchestration.

### Key Entities *(include if feature involves data)*

- **Web Fetch Tool**: Builtin network read tool with `source=builtin`, `group=web`, `risk_level=medium`, read-only approval policy, `execution_state=executable`, and read-only metadata.
- **Web Fetch Request Summary**: Safe request preview with normalized URL, host, max bytes, timeout, and redaction status.
- **Web Fetch Result Summary**: Safe bounded result with status code, final URL, content type, title, text excerpt, bytes read, limit, truncation flag, and redaction status.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A Chat or Work smoke can auto-execute `web.fetch` and observe exactly one bounded network read followed by provider continuation.
- **SC-002**: Safety tests prove unsafe schemes, credentialed URLs, localhost/private/link-local hosts, blocked redirects, stopped/terminal states, and unsupported tool names do not execute network reads.
- **SC-003**: Settings Tools and RunRail expose `web.fetch` scope/risk/read-only lifecycle metadata without leaking full bodies, cookies, credentials, or local paths.
- **SC-004**: The feature passes backend tests, web tests/build, docs build, diff check, and browser smoke for visible web tool states.

## Assumptions

- The first web runtime slice may use Go stdlib HTTP execution with deterministic test servers; production defaults reject private/local hosts.
- Tests may inject a test-only executor configuration to allow local `httptest` servers while production runtime keeps private-network denial enabled.
- Web search, browser automation, cookies, authenticated sessions, and artifact capture are separate later features.
