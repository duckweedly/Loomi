---
title: M96 Real Desktop Regression Smoke
description: P0-2 CLI and frontend render regression coverage for real API/provider/workspace desktop smoke.
---

## Completed

- Extended `loomi smoke agent` closeout checks so a terminal run with unresolved approvals blocks instead of passing.
- Kept the existing CLI smoke path as the single harness and added output fields for `thread_id`, `run_id`, `final_stage`, `tool_chain`, and final message excerpts.
- Blocked final assistant messages that are empty, `[redacted]`, or the generated failure placeholder `未生成成功回复`.
- Added frontend projection coverage for completed-without-final, failed run tool history, and Markdown table/code rendering.
- Documented a one-command real desktop acceptance path in the M79 harness runbook.

## Validation

- Red tests first:
  - `go test ./cmd/loomi -run 'TestSmokeAgentCommandBlocks(CompletedRunWithPendingApproval|GeneratedFailurePlaceholderFinalMessage)' -count=1` failed before the CLI checks were implemented.
  - `bun test --cwd web src/components/ChatCanvas.states.test.ts src/components/RunRail.runtime.test.ts --test-name-pattern 'flags a completed real API run that has no final assistant content|keeps failed run tool history visible|renders real smoke final markdown'` failed before the frontend missing-final and Markdown normalization fixes.
- Green targeted tests:
  - `go test ./cmd/loomi -run 'TestSmokeAgentCommandBlocks(CompletedRunWithPendingApproval|GeneratedFailurePlaceholderFinalMessage)' -count=1`
  - `bun test --cwd web src/components/ChatCanvas.states.test.ts src/components/RunRail.runtime.test.ts --test-name-pattern 'flags a completed real API run that has no final assistant content|keeps failed run tool history visible|renders real smoke final markdown'`
- Full validation:
  - `go test ./cmd/loomi ./internal/cli ./internal/httpapi ./internal/runtime -count=1`
  - `go test ./...`
  - `bun run --cwd web build`
  - `bun run --cwd docs-site build`
  - `git diff --check`

## Blockers

- Packaged desktop/browser automation is not yet wired. Current P0-2 coverage is the real CLI smoke plus frontend render regression.
- `bun test --cwd web` currently fails in existing `RunTimeline.runtime.test.ts` cases that render the right-panel menu instead of RunRail; the new P0-2 projection tests pass.
- The `loomi` binary is not installed on PATH in this shell, so live smoke used `go run ./cmd/loomi ...`.
- `go run ./cmd/loomi doctor --host http://127.0.0.1:18080 --provider local_codex` reached the API but reported `local_codex` detected and not enabled for this API session.
- `go run ./cmd/loomi smoke agent --host http://127.0.0.1:18080 --provider local_codex --workspace /Users/xuean/Repos/personal-projects/Loomi --auto-approve` selected workspace `Loomi` but blocked at `provider_check`.
