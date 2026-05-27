# Plan: M60 Memory External Provider Read Adapters

1. Add `GetMemoryProviderConfig` to the internal productdata service/repository boundary.
2. Implement read-only OpenViking and Nowledge HTTP clients in runtime with bounded JSON parsing and redacted errors.
3. Route `memory.search`, `memory.read`, and derived `memory.context` through the selected external provider when configured.
4. Add local `httptest` coverage for OpenViking and Nowledge search/read without hitting real services.
5. Update architecture/API/runbook/devlog documentation.
