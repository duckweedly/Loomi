# Data Model: Real Testing Console & Background UX

## ConfiguredProviderView

Represents backend-discovered provider configuration that affects real model calls.

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| `id` | string | yes | Stable provider identifier shown to user |
| `family` | string | yes | Provider family |
| `model` | string | yes | Model configured for backend calls |
| `baseUrl` | string | no | Displayable base URL; must not include secrets |
| `status` | ProviderStatus | yes | `unknown`, `checking`, `success`, `failed`, `unconfigured` |
| `message` | string | no | Sanitized human-readable status/error |
| `lastCheckedAt` | string | no | Optional display timestamp |

## ProviderStatus

| Value | Meaning | UI Behavior |
|-------|---------|-------------|
| `unknown` | Provider exists but has not been checked in this session | Neutral status |
| `checking` | Connection/readiness check in progress | Disable duplicate test action |
| `success` | Provider check succeeded | Show success state |
| `failed` | Provider check failed | Show sanitized error |
| `unconfigured` | No backend provider exists | Show env guidance |

## LocalDraftProvider

Browser-session-only draft provider fields.

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| `family` | string | no | Draft only |
| `model` | string | no | Draft only |
| `baseUrl` | string | no | Draft only |
| `apiKeyPresent` | boolean | no | Avoid storing/showing actual key |
| `dirty` | boolean | no | Indicates local draft differs from initial draft |

Rules:

- Must not be sent as backend provider configuration.
- Must not affect real model calls.
- Must be labeled as browser-session-only and unsaved.

## BackendCapability

| Value | Provider Readiness Required | Behavior |
|-------|-----------------------------|----------|
| `mock` | no | Chat may proceed without real provider |
| `real_api` | yes | Block real generation if provider unavailable |
| `model_gateway` | yes | Block model gateway generation if provider unavailable |

## ProviderUnavailableState

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| `capability` | BackendCapability | yes | `real_api` or `model_gateway` |
| `reason` | string | yes | Sanitized user-readable reason |
| `cta` | string | yes | Open Settings > Providers |
| `canGenerate` | boolean | yes | Always false for unavailable real modes |
| `canUseMock` | boolean | yes | True if mock mode remains selectable |

## BackgroundTaskSnapshot

Read-only snapshot shown in Background tasks panel.

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| `source` | SnapshotSource | yes | selected run job or empty |
| `runId` | string | no | Selected Chat run |
| `jobId` | string | no | Worker job associated with the selected run |
| `jobStatus` | JobStatus | no | Known worker job state |
| `observationState` | ObservationState | no | Derived UI state such as recovering |
| `diagnostics` | WorkerDiagnosticsSummary | no | Optional summary |
| `events` | WorkerJobEventView[] | yes | Latest events, empty if none |
| `isEmpty` | boolean | yes | Drives empty state |

## SnapshotSource

| Value | Meaning |
|-------|---------|
| `selected_run_job` | Job associated with selected Chat run |
| `empty` | No selected run job |

## JobStatus

| Value | Meaning |
|-------|---------|
| `queued` | Job is waiting to be claimed |
| `leased` | Worker has claimed job lease |
| `retrying` | Job is scheduled for another attempt |
| `completed` | Job completed successfully |
| `failed` | Job failed without being dead |
| `cancelled` | Job was cancelled |
| `dead` | Retries exhausted or job unrecoverable |
| `recovering` | Job is being recovered, if backend exposes this as a job status |

## ObservationState

| Value | Meaning |
|-------|---------|
| `recovering` | Derived from `job_recovering` when backend does not expose recovering as job status |

## WorkerDiagnosticsSummary

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| `workerId` | string | no | Display if available |
| `leaseState` | string | no | Human-readable lease summary |
| `attempt` | number | no | Current attempt |
| `maxAttempts` | number | no | Retry ceiling |
| `lastHeartbeatAt` | string | no | Optional timestamp |
| `message` | string | no | Sanitized diagnostic message |

## WorkerJobEventView

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| `type` | string | yes | Raw event type retained for debugging |
| `group` | EventGroup | yes | `runtime`, `worker`, `job`, `diagnostics` |
| `label` | string | yes | Localized readable label |
| `message` | string | no | Sanitized detail |
| `timestamp` | string | no | Event time |
| `severity` | EventSeverity | no | `info`, `warning`, `error`, `success` |

## ComposerStateView

| State | Primary Text | Controls |
|-------|--------------|----------|
| `idle` | Ready | Send enabled if input valid and capability available |
| `provider_unavailable` | Provider unavailable | Send disabled; Settings CTA visible |
| `queued` | Queued | Stop/cancel if supported; send disabled |
| `running` | Generating | Stop enabled if supported |
| `retrying` | Retrying | Send disabled; status explains retry |
| `recovering` | Recovering | Send disabled; status explains recovery |
| `stopped` | Stopped | Retry/regenerate if valid previous input exists |
| `failed` | Failed | Retry remains available until backend retryability is exposed |
| `closed` | Closed | Regenerate only if conversation/run allows |
| `cancelled` | Cancelled | Retry/regenerate if valid previous input exists |
