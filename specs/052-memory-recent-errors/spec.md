# Feature Spec: M52 Memory Recent Errors

## Goal

Expose recent memory provider diagnostics in Settings > Memory so users can see why a configured memory service is unhealthy or incomplete.

## Requirements

- Add a safe `/v1/memory/errors` endpoint.
- Derive current errors from backend memory provider diagnostic state.
- Redact provider secrets and key material.
- Show recent errors in Settings > Memory provider panel.

## Non-goals

- No external provider log ingestion.
- No desktop activity recorder errors.
- No write-failure analytics beyond the current provider diagnostic.
