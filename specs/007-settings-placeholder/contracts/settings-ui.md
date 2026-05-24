# Contract: M5.5 Settings UI

## Entry contract

- Settings opens from the existing sidebar Settings entry.
- Opening Settings must not change selected thread, active run, current messages, or right panel state unless the user explicitly changes a setting.
- Settings includes a back affordance that returns to the prior workspace view.

## Layout contract

Settings uses a desktop-style layout:

```text
settings-shell
├── settings-sidebar
│   ├── back/header
│   ├── primary categories
│   ├── Agent Core group
│   └── management group
└── settings-content
    ├── category title/status
    ├── grouped setting cards
    └── placeholder panels when selected category is mock
```

## Category contract

Required categories:

| Category | Status | Expected behavior |
| --- | --- | --- |
| General | working | Shows working local settings and read-only runtime state |
| Appearance | mock | Preview only |
| Providers | mixed | Shows redacted provider capability plus session-local draft fields for Base URL, model ID, and masked API key presence; no provider calls, backend writes, persistence, or secret echo |
| Connectors | mock | Preview only |
| Plugins | mock | Preview only |
| Skill | mock | Preview only |
| MCP | mock | Preview only |
| Notebook | mock | Preview only |
| Memory | mock | Preview only |
| Activity Recorder | mock | Preview only |
| Context | mock | Preview only |
| Safety | mock | Preview only |
| Tools | mock | Preview only |
| Routes | mock | Preview only |
| About | mixed | Shows known local app/version/status, other values as mock |
| Advanced | mock | Preview only |

## Working setting rows

| Setting | Control | Behavior |
| --- | --- | --- |
| Default workspace mode | segmented/select | Changes current-session default for future workspace/new conversation flow |
| Interface language | segmented/select | Defaults to Chinese and changes current-session shell/settings copy to English or Chinese |
| Mock runtime scenario | segmented/select | Changes future mock sends between success/failure scripts |
| Data source mode | status | Displays mock or real API mode |
| Backend capability | status | Displays available/unavailable/misconfigured |
| Provider capability | status list | Displays redacted provider id/family/model/status when available |
| Provider draft Base URL | text input | Stores a current-session OpenAI-compatible gateway URL draft only |
| Provider draft model ID | text input | Stores a current-session model ID draft only |
| Provider draft API key | password input | Masks entry and retains only whether a value was entered; the key string is not echoed, persisted, or sent to the backend |

## Placeholder safety contract

Placeholder controls must:

- show mock/preview/not connected copy
- avoid provider calls, tool execution, connector calls, filesystem writes, or backend write operations
- avoid asking for API keys or secrets outside the Providers draft panel
- not persist values as real settings

Provider draft controls must:

- keep Base URL and model ID in current-session UI state only
- use a password input for API key entry
- retain only whether an API key value was entered
- never echo, document, log, persist, or send the API key string

## Visual contract

- Use Loomi-owned copy and naming.
- The reference image is used for layout direction only.
- Cards must have clear group titles, row labels, helper text, and right-aligned controls/status.
- Working and placeholder rows must be distinguishable without reading documentation.
