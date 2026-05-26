# Research: M27 Browser Automation Foundation

## Decision: Start with HTTP-backed Stateful Browser Sessions

**Rationale**: M26 already proved one-off public fetch. Browser automation needs stateful page/session semantics and approved navigation before full browser engines. HTTP-backed sessions let Loomi prove open, snapshot, and click-link flows through the current tool runtime without cookies, profiles, JS, screenshots, or downloads.

**Alternatives considered**:

- Use a real browser engine immediately: deferred because it introduces process lifecycle, cookies/profile access, JavaScript execution, screenshots, and much larger permission boundaries.
- Treat `browser.open` as an alias for `web.fetch`: rejected because browser automation must create navigable state.

## Decision: Public HTTP-only and No Cookies/Profile State

**Rationale**: Browser automation can otherwise access authenticated pages, localhost apps, routers, metadata services, or VPN/private resources. The first slice must stay public-network-only and independent of user browser state.

**Alternatives considered**:

- Reuse the user's logged-in browser session: rejected for this slice because it needs explicit cookie/profile permissions and sensitive-data controls.
- Allow localhost for local apps: rejected for production defaults. Tests can inject a local allow-private configuration.

## Decision: Snapshot Summaries, Not Raw HTML

**Rationale**: Run events need auditability and provider continuation needs useful context, but raw HTML/body data can include secrets or large copyrighted content. Browser results persist bounded title, text excerpt, link summaries, status, and URL metadata.

**Alternatives considered**:

- Store full DOM/HTML: rejected due to leakage and storage risk.
- Store only URL/status: rejected because the agent cannot meaningfully inspect page state.
