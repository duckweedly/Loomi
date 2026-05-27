---
title: 2026-05-27 M79 Agent Harness Smoke
description: Real CLI/API worker/provider/tool approval smoke closeout path.
---

Implemented:

- Added `loomi smoke agent` as a real harness command that checks API readiness, checks provider completion, starts a real run, streams events, optionally auto-approves tool calls, and prints `thread_id`, `run_id`, final stage, provider check stage, event counts, and last event summaries.
- Added actionable provider-boundary output for blocked smokes. HTTP 401/403 maps to token refresh, 429 to quota/provider switch, and 503 to retry or provider switch.
- Added `loomi doctor --provider <id>` plus provider `check_stage` in doctor output, so local readiness can distinguish config presence from upstream completion availability.
- Updated provider completion diagnostics to emit semantic check codes for auth, rate limit, and provider-unavailable failures without leaking response bodies or tokens.

Arkloop comparison:

- Arkloop treats smoke as a real stack check with explicit API URL, durable run creation, provider boundary verification, and resume/replay evidence.
- Loomi keeps its own CLI/API/event names and does not adopt Arkloop service names, install flow, Docker stack, or copy. M79 borrows only the acceptance shape: a reproducible live run path with clear blocked reasons.

Focused validation:

```bash
go test ./cmd/loomi -run 'TestDoctorCommandExplainsProviderAuthFailure|TestSmokeAgentCommand' -count=1
go test ./internal/runtime -run 'TestCheckProviderCompletionReportsHTTP503WithoutLeakingBody|TestCheckProviderCompletionReportsHTTP401AsAuthFailure' -count=1
```

Full validation:

```bash
go test ./cmd/loomi ./internal/cli ./internal/runtime ./internal/httpapi -count=1
go test ./...
bun run --cwd docs-site build
git diff --check
```

Real smoke evidence:

```text
smoke ok
thread_id	thr_1779856040659901000_8802413111eb
run_id	run_1779856040675017000_bdd99acb9f26
final_stage	run_completed
provider	local_codex status=available execution=supported model=gpt-5.5
events	35 total, 3 tool, 1 approvals
```
