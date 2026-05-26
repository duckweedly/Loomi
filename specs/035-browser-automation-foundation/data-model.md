# Data Model: M27 Browser Automation Foundation

## BrowserToolCatalogEntry

Builtin browser automation tools.

Names:

- `browser.open`
- `browser.snapshot`
- `browser.click_link`

Shared fields:

- `source`: `builtin`
- `group`: `browser`
- `risk_level`: `medium`
- `approval_policy`: `always_required`
- `execution_state`: `executable`
- `safe_metadata.scope`: `browser`
- `safe_metadata.network_access`: `public_http_only`
- `safe_metadata.stateful`: `true`

## BrowserSession

Run-scoped browser state.

Fields:

- `session_id`: safe deterministic session id scoped to a run/tool execution
- `run_id`: owning run id
- `url`: requested safe URL
- `final_url`: current safe URL after redirects/navigation
- `status_code`: current HTTP status
- `content_type`: safe content type
- `title`: optional extracted title
- `text_excerpt`: bounded UTF-8 safe text excerpt
- `links`: bounded BrowserLinkSummary list
- `bytes_read`: body bytes read
- `byte_limit`: configured byte limit
- `truncated`: boolean
- `updated_at`: runtime timestamp

Validation:

- No cookies, raw headers, raw HTML, full body, credentials, local paths, or secret-looking values.
- Session is valid only for the owning run.

## BrowserLinkSummary

Safe link available from the current snapshot.

Fields:

- `index`: zero-based integer
- `text`: bounded safe link text
- `href`: resolved safe href without credentials/fragments
- `host`: safe host
- `blocked`: boolean

Validation:

- Link list is bounded.
- Blocked links are visible as non-clickable metadata but must not be navigated.

## BrowserResultSummary

Safe result envelope.

Fields:

- `tool`: browser tool name
- `scope`: `browser`
- `operation`: `open`, `snapshot`, or `click_link`
- `session_id`: safe session id
- `url`: requested/current safe URL
- `final_url`: final safe URL
- `previous_url`: optional for click
- `status_code`: HTTP status
- `content_type`: safe content type
- `title`: optional title
- `text_excerpt`: bounded page text
- `links`: bounded BrowserLinkSummary list
- `bytes_read`: integer
- `byte_limit`: integer
- `truncated`: boolean
- `redaction_applied`: boolean
