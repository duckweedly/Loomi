---
title: Local M12 Todo Write Planning Tool
description: Validate runtime.todo_write locally.
---

## Focused Validation

```bash
go test ./internal/productdata ./internal/runtime
bun test --cwd web ./src/components/ToolCallCard.test.tsx
```

Expected:

- `runtime.todo_write` is allowlisted and approval-required.
- invalid todo item arrays are rejected before approval.
- approved worker execution stores total/pending/in-progress/completed counts.
- ToolCallCard renders item summaries without raw object output.

## Full Validation

```bash
go test ./...
bun test --cwd web
bun run --cwd web build
cd docs-site && bun run build
git diff --check
```

## Browser Smoke

Start the web shell:

```bash
bun run --cwd web dev -- --host 127.0.0.1 --port 5173
```

Open Settings > Tools and confirm `runtime.todo_write` appears with capability `plan`, risk `low`, side effect `none`, and safety class `no_side_effect_internal`.
