# Plan: M70 Memory Provider Error UI

## Scope

Use the existing Settings > Memory recent errors panel and enrich each item with optional runtime fields.

## Frontend

- Add a small formatter for memory provider error summaries.
- Include `eventType` and `runId` when present.
- Add wrapping style for long run ids.

## Documentation

- Update memory docs and current Spec Kit status.
- Add devlog validation evidence.

## Validation

```bash
bun test --cwd web src/components/SettingsView.runtime.test.tsx src/memory.test.ts
bun run --cwd web build
cd docs-site && bun run build
git diff --check
```
