---
title: M63 Memory External Post-Run Commit
description: External provider commit-after-run behavior.
---

## Completed

- Routed external-provider `commit_after_run` through the provider write adapter instead of local pending proposals.
- Preserved local memory's pending proposal behavior.
- Added terminal-run-safe `memory_provider_commit_completed` and `memory_provider_commit_failed` events.
- Made external post-run commit idempotent per run using the safe commit event.

## Validation

- `go test ./internal/productdata ./internal/runtime -run 'TestPostRunMemory|TestWorkerProposesPostRunMemoryWhenCommitAfterRunEnabled|TestMemoryToolExecutorSearchesNowledgeProvider' -count=1`
