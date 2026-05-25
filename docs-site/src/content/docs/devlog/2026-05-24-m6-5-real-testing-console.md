---
title: M6.5 Real Testing Console and Background UX
description: Provider readiness, Background tasks, Timeline, Composer polish, docs, and validation for local real testing.
---

## Summary

M6.5 productizes the real local testing path on top of the M5.5/M6 provider readiness and worker job pipeline foundations.

Completed scope:

- Provider Test Console wording and provider base URL display in Settings
- provider unavailable gating for real provider-dependent modes including model gateway
- mock mode remains unaffected by provider readiness
- read-only Background tasks observer with empty state, current run job status, worker diagnostics, and latest worker/job events
- productized M6 worker/job event labels in RunRail using existing i18n copy
- unknown worker event fallback with raw event type
- incoming Timeline event order preserved inside groups
- Composer action coverage for provider unavailable, queued/running/retrying/recovering, and terminal retry/regenerate cases
- Spec Kit 010 artifacts under `specs/010-real-testing-console-background-ux/`

## Boundaries

M6.5 does not add M7 tool calls, approval, tool execution protocol, provider secret storage, persistent API key management, or a worker job data-model rewrite.

## Validation

Focused validation run during implementation:

```bash
bun test web/src/components/SettingsView.runtime.test.tsx
bun test web/src/components/ChatCanvas.states.test.ts web/src/runtime/backendCapabilityStatus.test.ts
bun test web/src/runtime/runtimeEventGroups.test.ts web/src/components/RunRail.polish.test.ts
bun test web/src/components/RightToolDrawer.backgroundTasks.test.tsx
bun test web/src/runtime/composerActions.test.ts web/src/components/Composer.test.ts
```

Final validation completed:

```bash
zsh -o null_glob -c 'bun test ./web/src/*.test.ts ./web/src/*.test.tsx ./web/src/components/*.test.ts ./web/src/components/*.test.tsx ./web/src/runtime/*.test.ts'
# 202 pass, 0 fail

bun run --cwd web build
# pass

bun run --cwd docs-site build
# pass, 45 pages built
```

## Known limitations

The Background tasks panel derives its observer snapshot from the selected run and its persisted events; cross-run active-job discovery is deferred until a richer right-panel data contract exists. Provider Test Console still uses existing provider readiness/check surfaces and does not add secret storage. Non-retryable failure display is deferred until backend retryability is exposed.
