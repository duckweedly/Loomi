# Implementation Plan: Memory Agent Tools

## Scope

Implement the second memory parity slice by adding memory tools to the existing Loomi tool runtime. The tools reuse current productdata memory APIs and provider status; no external semantic memory service or automatic distillation is introduced.

## Architecture

- `internal/productdata`: add memory tool constants, catalog entries, request validation, and default persona allowlist.
- `internal/runtime`: add `MemoryToolExecutor` and route memory tools through `DefaultToolExecutor`.
- `internal/runtime` provider serialization: expose function schemas for memory tools through the existing provider tool schema path.
- `cmd/loomi-api`: pass the product memory executor into the worker router.
- `web`: rely on existing tool catalog/RunRail surfaces; add focused tests for memory tool labels if needed.
- `docs-site`: document tool semantics, approval behavior, safety, and deferred distillation.

## Safety

All memory tools are approval-gated in this slice. Search/read/status are read-only but still require approval for observability. Write creates proposals only. Forget tombstones through existing scoped delete logic. Outputs are passed through existing redaction and safe summary filters.

## Validation

Run focused productdata/runtime tests for catalog, validation, executor, and worker continuation; then focused web tests, docs build, and browser smoke for Settings > Tools/RunRail visibility.
