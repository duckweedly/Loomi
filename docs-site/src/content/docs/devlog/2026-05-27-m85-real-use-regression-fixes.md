---
title: M85 Real Use Regression Fixes
description: Closeout notes for assistant event redaction and directory-read run failure regressions.
---

Fixed two regressions found during real desktop testing:

- Assistant final event content now preserves normal product language such as "Design Tokens", "API tokens", and "key ideas". Sensitive error text and metadata are still redacted, but provider-visible assistant output is not collapsed to `[redacted]` just because it contains benign words.
- `workspace.read` pointed at a directory now returns a safe directory summary instead of raising `workspace path is a directory` and failing the whole run.

Validation:

- `go test ./internal/productdata -count=1`
- `go test ./internal/runtime -run TestWorkspaceReadDirectoryReturnsSafeSummary -count=1`
