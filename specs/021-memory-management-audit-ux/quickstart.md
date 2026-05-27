# Quickstart: M14 Memory Management Audit UX

1. Apply existing migrations through the M13 memory tables.
2. Start the API with the local Postgres database.
3. Start the web shell with `VITE_LOOMI_API_BASE_URL` pointing at the local API.
4. Seed or exercise memory write proposal, approve, deny, snapshot load, and delete flows.
5. Seed one approved entry for browser smoke.
6. Open Settings > Memory.
7. Verify approved memories list with safe summaries.
8. Search and apply grounded scope/source filters.
9. Open a memory detail drawer/modal and confirm only safe metadata appears.
10. Delete a memory only through the confirmation flow.
11. Open memory history and verify real safe audit events appear.

Prep/blocker checks:

- Thread-scoped memory detail and delete are authorized for the owner.
- Memory audit survives if the source run is already terminal.
- Redaction covers `/home`, Windows paths, stdout/stderr, provider trace, Authorization/env/token-like strings.
- List/search share the same filter shape.

Required validation:

```sh
go test ./...
bun test --cwd web
bun run --cwd web build
bun run --cwd docs-site build
git diff --check
```

Browser smoke must cover Settings > Memory list, search, detail, delete confirmation, and audit history.
