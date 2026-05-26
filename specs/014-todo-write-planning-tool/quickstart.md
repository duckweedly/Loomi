# Quickstart: M12 Todo Write Planning Tool

## Backend

Run focused tests:

```bash
go test ./internal/productdata ./internal/runtime
```

Expected:

- valid `runtime.todo_write` requests are accepted only when approval is required
- invalid item arrays are rejected
- approved worker execution stores todo counts in result summary

## Frontend

Run focused tests:

```bash
bun test --cwd web ./src/components/ToolCallCard.test.tsx ./src/components/settingsCatalog.test.ts
```

Expected:

- ToolCallCard renders todo_write item/count summaries
- Settings tool catalog can include the planning tool

## Full Validation

```bash
go test ./...
bun test --cwd web
bun run --cwd web build
cd docs-site && bun run build
git diff --check
```
