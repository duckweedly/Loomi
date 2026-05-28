---
title: M86 Persisted Message Finalization
description: Closeout notes for Arkloop-style terminal message reconciliation.
---

Real desktop testing showed that a completed run can still display a bad stream draft if the UI treats run events as the final chat source. M86 moves the finalization boundary closer to Arkloop's pattern:

- `GET /v1/threads/{thread_id}/messages` now projects assistant `run_id` as a first-class response field when it is present in safe metadata.
- The real API client maps `run_id` from either the top-level field or safe metadata for compatibility.
- `useWorkspaceState` reconciles terminal runs with the persisted assistant message for the same `run_id` during refresh and SSE closeout. If a stream draft differs from the stored assistant message, the stored message wins.
- Live `model_output_completed` still promotes the draft immediately for responsive UI, but it is no longer the only terminal source of truth.

Validation:

- `go test ./internal/httpapi -run 'TestMessageHandlers|TestMessageListProjectsAssistantRunID' -count=1`
- `bun test --cwd web ./src/realApiClient.test.ts ./src/state.runtime.test.ts -t 'terminal reconciliation|source of truth|run id from safe metadata'`
