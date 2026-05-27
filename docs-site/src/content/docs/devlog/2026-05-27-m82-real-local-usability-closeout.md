---
title: 2026-05-27 M82 Real Local Usability Closeout
description: Local CLI doctor and real agent smoke usability closeout.
---

Implemented:

- Changed the CLI default API host to `http://127.0.0.1:18080`, matching the local API runbook and avoiding unrelated local services on 8080.
- Added CLI bearer-token support through `LOOMI_API_TOKEN` and `loomi config set api_token <token>`. Text and JSON config output expose only whether a token is set.
- Made `loomi doctor` map `401 missing bearer token` to explicit API-token setup guidance.
- Made `loomi doctor --provider local_codex` report a clear blocked reason when Local Codex is detected but not enabled, and a ready state after explicit enablement.
- Stabilized sidebar archive behavior so deleting another thread keeps the current selection, while deleting the selected thread moves to an adjacent thread instead of jumping to the top of the refreshed list.
- Flattened assistant markdown code blocks inside message bubbles and normalized escaped or indented heading markers so rendered headings do not show literal `#` prefixes.

Real local evidence:

```text
doctor ok
ok api http://127.0.0.1:18080
ok providers local_codex status=available execution=supported model=gpt-5.5
ok tools 50 tools, 50 enabled, 11 groups
```

```text
smoke ok
thread_id thr_1779861294575417000_71c96fe2b8eb
run_id run_1779861294596954000_78e89c7fc75e
final_stage run_completed
provider local_codex status=available execution=supported model=gpt-5.5
events 24 total, 2 tool, 0 approvals
```

Tool event evidence:

```text
0015 requested workspace.read ... args=path=AGENTS.md limit=1000000
0016 approved workspace.read ...
0018 executing workspace.read ...
0019 succeeded workspace.read ... result=path=AGENTS.md bytes=1327 truncated=false
```

Boundaries:

- No new tools, Docker services, Redis, Firecracker, or multi-agent architecture were added.
- No API tokens, Codex tokens, or Authorization values are printed in CLI output or docs.
- `tmp/` remains out of scope for this closeout.

Validation:

```bash
go test ./cmd/loomi ./internal/cli ./internal/runtime ./internal/httpapi -count=1
go test ./...
bun test --cwd web
bun run --cwd web build
bun run --cwd docs-site build
git diff --check
```
