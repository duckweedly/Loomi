# Data Model: M5.5 Settings Placeholder

## Settings Category

Represents a navigation item in the settings sidebar.

| Field | Type | Rules |
| --- | --- | --- |
| `id` | string | Stable category id used for selection |
| `label` | string | Loomi-owned display label |
| `group` | enum | `primary`, `agent_core`, or `management` |
| `status` | enum | `working`, `mock`, `disabled`, `preview`, `read_only`, or `mixed` |
| `description` | string | Short explanation for placeholder/detail panel |

Rules:

- General is the default working category.
- Future categories must be marked as mock/preview/disabled.
- Category labels must use Loomi product language, not copied reference labels where they imply unsupported behavior.

## Setting Section

Represents a grouped card in the settings content area.

| Field | Type | Rules |
| --- | --- | --- |
| `id` | string | Stable section id |
| `title` | string | User-visible group heading |
| `description` | string | Optional helper text |
| `category_id` | string | Parent settings category |
| `rows` | list | Ordered setting rows |

Rules:

- Working and placeholder rows may appear in the same category only when clearly distinguished.
- Sections must support compact desktop-style cards.

## Setting Row

Represents one configurable or placeholder row.

| Field | Type | Rules |
| --- | --- | --- |
| `id` | string | Stable row id |
| `label` | string | User-visible label |
| `helper_text` | string | Explains behavior or placeholder status |
| `control_type` | enum | `toggle`, `select`, `button`, `status`, `placeholder`, `text`, `password`, or `segmented` |
| `status` | enum | `working`, `mock`, `disabled`, or `read_only` |
| `value` | string/boolean/null | Current visible value |

Rules:

- Working rows must have a visible current-session effect.
- Placeholder rows must not perform external actions or persist real settings.
- Provider draft rows may collect current-session Base URL, model ID, and masked key presence only; provider capability rows remain read-only.

## Local Settings State

Represents current-session settings that affect existing behavior.

| Field | Type | Rules |
| --- | --- | --- |
| `default_mode` | enum | `chat` or `work`; affects future workspace entry/new conversation preference |
| `locale` | enum | `zh` or `en`; defaults to `zh` and affects current-session interface copy |
| `mock_runtime_scenario` | enum | `success` or `failure`; affects future mock sends only |
| `settings_category_id` | string | Last selected settings category during current session |
| `provider_draft` | object | Current-session provider draft state only |

Rules:

- State is session-local for M5.5.
- Changing locale does not persist a user preference.
- Changing mock scenario does not mutate active runs.
- Returning from Settings preserves selected thread/workspace context.

## Provider Draft State

Represents current-session provider configuration notes without real provider management.

| Field | Type | Rules |
| --- | --- | --- |
| `base_url` | string | Current-session OpenAI-compatible gateway URL draft only |
| `model` | string | Current-session model ID draft only |
| `api_key_set` | boolean | Whether the user typed a key; the key string is not retained |

Rules:

- Draft values are not persisted, written to the backend, or used for provider calls in M5.5.
- API key entry must be masked and must not be echoed, documented, logged, or stored as a string.

## Runtime Capability Summary

Read-only summary of current runtime/provider capability.

| Field | Type | Rules |
| --- | --- | --- |
| `data_source_mode` | enum | `mock` or `real_api` |
| `backend_capability` | enum | `available`, `unavailable`, or `misconfigured` |
| `provider_capabilities` | list | Redacted provider status items when available |
| `last_status_message` | string/null | User-safe status copy only |

Rules:

- Must not include API keys, Authorization headers, raw provider payloads, or secret URL fragments.
- If provider capability cannot be loaded, show a user-safe unavailable or not-connected state.

## Placeholder Setting

Represents future product IA without functional behavior.

| Field | Type | Rules |
| --- | --- | --- |
| `id` | string | Stable placeholder id |
| `future_area` | string | Planned capability area |
| `status` | enum | `mock`, `preview`, or `disabled` |
| `safe_copy` | string | Explains that the control is not connected |

Rules:

- Placeholder controls never call providers, tools, file systems, external connectors, or backend write endpoints.
- Placeholder copy should help users understand roadmap intent without promising availability.
