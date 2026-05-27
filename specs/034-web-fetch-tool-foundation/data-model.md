# Data Model: M26 Web Fetch Tool Foundation

## WebFetchToolCatalogEntry

Builtin read-only network tool.

Fields:

- `name`: `web.fetch`
- `source`: `builtin`
- `group`: `web`
- `risk_level`: `medium`
- `approval_policy`: `read_only`
- `execution_state`: `executable`
- `safe_metadata.scope`: `web`
- `safe_metadata.read_only`: `true`
- `safe_metadata.network_access`: `public_http_only`

## WebFetchRequestSummary

Safe argument preview.

Fields:

- `url`: normalized absolute HTTP(S) URL without credentials
- `host`: normalized host
- `max_bytes`: bounded integer
- `timeout_ms`: bounded integer
- `redaction_applied`: boolean

Validation:

- URL must be absolute `http://` or `https://`.
- URL must not contain username/password credentials.
- Host must not be localhost, loopback, link-local, private, multicast, unspecified, or otherwise blocked private-network address.
- Redirect targets must pass the same validation before body read.

## WebFetchResultSummary

Safe bounded result.

Fields:

- `tool`: `web.fetch`
- `scope`: `web`
- `operation`: `fetch`
- `url`: requested safe URL
- `final_url`: final safe URL after redirects
- `status_code`: HTTP status code
- `content_type`: response content type without parameters that could contain secrets
- `title`: optional extracted HTML title
- `text_excerpt`: bounded UTF-8 safe text excerpt
- `bytes_read`: bytes read from response body
- `byte_limit`: configured byte limit
- `truncated`: boolean
- `redaction_applied`: boolean

Validation:

- No full response body, cookies, Authorization values, Set-Cookie values, raw headers, local paths, or secret-looking values.
- Unsupported non-text content returns metadata without body excerpt.
