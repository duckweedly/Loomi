# Contract: Backend Capability Status

## Purpose

Define the user-readable mode and capability statuses shown in Chat Canvas, Timeline, and debug surfaces.

## Status values

| Status | Meaning | User-facing implication |
|--------|---------|-------------------------|
| `mock` | Deterministic frontend mock execution | This is local/mock behavior, not real model output |
| `local-simulated` | Real API path with local simulated execution | Backend path is connected, but output is simulated |
| `real-model` | Real provider/model execution is available | Model output can be treated as real provider output |
| `backend-unavailable` | Backend cannot be reached or lacks runtime capability | Runtime action cannot proceed as real execution |
| `model-setup-missing` | Required model configuration is missing | User or developer setup is required before real generation |
| `provider-unavailable` | Provider is unavailable or rejects execution | The provider failed before or during generation |
| `stream-disconnected` | Event stream disconnected before terminal reconciliation | The run may still be active; UI needs reconciliation |
| `run-recovering` | UI is restoring latest known run/draft state | Visible state is temporary until reconciliation completes |

## Precedence rules

1. `run-recovering` while the selected run is actively reconciling.
2. `stream-disconnected` when an active run loses event stream before terminal state.
3. `provider-unavailable` for provider-originated execution failures.
4. `model-setup-missing` for missing model/key/configuration capability.
5. `backend-unavailable` when the configured backend path cannot provide runtime behavior.
6. `real-model` when real provider execution is available.
7. `local-simulated` when real API path is connected but output is simulated.
8. `mock` when using deterministic frontend mock mode.

## Copy rules

- Never describe mock or local simulated mode as real model execution.
- Distinguish backend unavailable from model failure.
- Distinguish stream disconnected from completed/failed run state.
- Keep status copy short enough for header chips while allowing longer detail in timeline/debug.

## Acceptance checks

- Mock mode is visibly labeled as mock.
- Local simulated mode is visibly distinct from real model mode.
- Backend unavailable does not show model-thinking copy.
- Stream disconnected status remains visible until the latest run state is reconciled.
- Provider unavailable appears as a provider/runtime problem, not a user input validation error.
