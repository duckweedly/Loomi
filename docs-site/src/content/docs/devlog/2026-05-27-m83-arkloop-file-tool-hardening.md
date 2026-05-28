---
title: M83 Arkloop File Tool Hardening
description: Closeout notes for provider-facing load_tools and bounded workspace file discovery hardening.
---

Compared the first-use failure path against the Arkloop reference under `tmp/arkloop-reference/Arkloop` and hardened Loomi without copying Arkloop product expression.

What changed:

- `tool.load_tools` remains backward-compatible with older `names` arguments, but the provider-facing schema is now query-only with required `queries`, matching the safer catalog-keyword flow.
- Catalog query matching now tokenizes short natural phrases, so queries like `workspace list files directory glob ls` can resolve the relevant enabled workspace tool descriptions instead of returning an empty catalog.
- `workspace.glob` now skips generated dependency/cache folders and returns `skipped_dir_count`.
- `workspace.grep` now skips the same generated folders, skips oversized files, stops no-match searches after a bounded scanned-file budget, and returns `scanned_file_count` / `skipped_file_count`.

Validation:

- `go test ./internal/runtime -run 'TestDiscoveryLoadToolsMatchesCatalogKeywordsInQueryPhrase|TestGatewayLoadToolsProviderSchemaIsQueriesOnly|TestWorkspaceGlobSkipsGeneratedDirectories|TestWorkspaceGrepNoMatchIsBoundedByScannedFiles' -count=1`
