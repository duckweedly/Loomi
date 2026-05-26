# Research: M26 Web Fetch Tool Foundation

## Decision: Start with `web.fetch` Only

**Rationale**: A single explicit URL fetch is the smallest runnable network-read slice after workspace, sandbox, MCP, and LSP tools. It proves catalog, approval, execution, event, and UI semantics before search, browser automation, crawler, or artifact complexity.

**Alternatives considered**:

- Add `web.search`: deferred because it needs provider choice, quotas, result trust policy, and network credentials.
- Add browser automation first: rejected because it adds cookies, profile state, JavaScript, screenshots, and DOM/action permissions.
- Add crawler behavior: rejected because multi-URL traversal needs robots/rate/scope policy and much larger safety review.

## Decision: Reject Private/Local Network Targets by Default

**Rationale**: A model-requested network tool can otherwise probe localhost, link-local, VPN, Docker, router, cloud metadata, or private intranet services. Production runtime must reject those targets before dialing and must validate redirects before reading bodies.

**Alternatives considered**:

- Allow localhost for local development: rejected for production behavior. Tests can inject an executor that allows private hosts for deterministic `httptest` coverage.
- Rely only on URL string checks: rejected because DNS names can resolve to private IPs.

## Decision: Persist Summaries, Not Full Bodies

**Rationale**: Run events and UI need auditability, but full response bodies can contain credentials, account data, or large copyrighted content. The tool returns bounded title/excerpt/status metadata to the provider continuation while storing only safe summaries.

**Alternatives considered**:

- Store complete text body: rejected because it increases leakage and storage risk.
- Store no excerpt: rejected because the provider continuation needs enough content to be useful in the first slice.
