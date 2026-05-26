---
title: Local M11 Tool Catalog
description: Validate the read-only tool catalog API and Settings panel.
---

## API Smoke

Start the local API, then request:

```bash
curl -s http://127.0.0.1:8080/v1/tools/catalog
```

Expected:

- seven tools in deterministic order
- all tools show `approval_policy: "required"`
- workspace write and exec tools show `risk_level: "high"`
- no credentials, file contents, provider payloads, command output, or executable examples

## Frontend Smoke

Start the web shell:

```bash
bun run --cwd web dev -- --host 127.0.0.1 --port 5173
```

Open Settings > Tools. Expected:

- Tools category is read-only, not placeholder
- all catalog rows render name, description, approval policy, risk, side effect, and safety class
- no controls execute or mutate tools
- browser console has no errors

## Validation

```bash
go test ./...
bun test --cwd web
bun run --cwd web build
cd docs-site && bun run build
git diff --check
```
