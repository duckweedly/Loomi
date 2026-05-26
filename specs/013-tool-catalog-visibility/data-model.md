# Data Model: M11 Tool Catalog Visibility

## Tool Catalog Entry

- `name`: canonical tool name
- `label`: short human label
- `group`: `runtime` or `workspace`
- `capability`: `time`, `read`, `write`, or `exec`
- `approval_policy`: `always_required`
- `safety_class`: current runtime safety class
- `risk_level`: `low`, `medium`, or `high`
- `side_effect`: `none`, `read`, `write`, or `process`
- `enabled`: boolean
- `description`: short safe description

## Tool Catalog

- `tools`: deterministic ordered list of entries
- `updated_at`: response timestamp
