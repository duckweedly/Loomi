---
title: P1/P2 Readiness And Final Guards
description: Tightened desktop workspace readiness and structured provider final extraction.
---

## Completed

- Desktop readiness now treats a missing workspace root config as `workspace_unselected` once API, DB, provider, and tool catalog checks are otherwise ready.
- Structured provider finals now preserve common nested answer shapes, including `result.summary`, `output_text`, and arrays of output text segments.
- The structured final guard still avoids persisting raw tool protocol JSON when no natural-language answer can be extracted.

## Validation

- `bun test --cwd web src/runtime/desktopReadiness.test.ts`
- `go test ./internal/runtime -count=1`
