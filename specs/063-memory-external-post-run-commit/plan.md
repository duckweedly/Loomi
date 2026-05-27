# Plan: M63 Memory External Post-Run Commit

1. Detect configured external provider state inside post-run memory closeout.
2. Route external post-run commits through the existing external memory write adapter.
3. Add terminal-run-safe commit events for success/failure and idempotency.
4. Preserve local proposal behavior.
5. Add local `httptest` coverage.
6. Update docs.
