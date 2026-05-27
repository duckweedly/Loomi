---
title: M44 Memory Post-run Proposals
description: Commit-after-run now creates approval-gated memory write proposals from completed runs.
---

M44 connects the Settings > Memory `commit_after_run` preference to runtime closeout.

Implemented:

- Added runtime post-run proposal helper that checks completed run status, memory provider readiness, and the commit-after-run toggle.
- Runtime closeout now attempts one idempotent pending `memory_write_proposed` record per completed run.
- Post-run proposals are thread-scoped, source-run-linked, and use `post_run_memory:{run_id}` idempotency.
- Pending proposals remain invisible to memory search until approved through the existing write-proposal flow.
- Settings copy now states that each-run organization creates approval-gated proposals.

Validation:

```bash
go test ./internal/runtime -run 'TestWorkerProposesPostRunMemory|TestPostRunMemory'
```

Deferred: LLM distillation, embeddings/vector retrieval, external semantic provider execution, automatic approval, background memory workers, and multi-agent long-term memory automation.
