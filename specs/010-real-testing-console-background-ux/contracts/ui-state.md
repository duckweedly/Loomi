# UI State Contract: Real Testing Console & Background UX

## Provider Test Console

### Sections

1. Header
   - Title: Provider Test Console
   - Subtitle: Explains backend configured providers affect real model calls.

2. Configured Providers
   - Source: backend provider readiness/configuration.
   - Visible fields:
     - provider id
     - family
     - model
     - base URL
     - status
     - message
   - Actions:
     - Test connection

3. Local Draft
   - Source: browser session only.
   - Required copy:
     - Local draft
     - Stored only in this browser session
     - Does not change backend configured providers
     - Does not affect real model calls

4. Empty State
   - Trigger: no backend configured providers.
   - Required copy:
     - no providers configured
     - configure provider env vars and restart API
     - local draft does not enable real calls

### Provider Check State

| State | Label | Action Behavior |
|-------|-------|-----------------|
| `unknown` | Not checked | Test connection enabled |
| `checking` | Checking | Test connection disabled |
| `success` | Connected | Test connection enabled |
| `failed` | Failed | Test connection enabled |
| `unconfigured` | Not configured | Test connection hidden or disabled |

### Provider Test Action

Test connection is a per-provider UI action. If existing backend readiness is aggregate-only, the UI maps the aggregate result back to provider rows without adding provider secret storage APIs.

## Chat / Composer Provider Readiness

| Runtime Capability | Provider Available | Composer Behavior |
|--------------------|--------------------|-------------------|
| `mock` | no | Send allowed if mock mode is otherwise valid |
| `mock` | yes | Send allowed if mock mode is otherwise valid |
| `real_api` | no | Send disabled; provider unavailable CTA |
| `real_api` | yes | Normal queued/running flow |
| `model_gateway` | no | Send disabled; provider unavailable CTA |
| `model_gateway` | yes | Normal queued/running flow |

### Provider Unavailable Message

Required content:

- Provider unavailable
- Current backend capability requires configured provider
- Open Settings > Providers CTA
- No claim that generation is running

## Background Tasks Snapshot Selection

The panel selects one read-only snapshot using this priority:

| Priority | Source | Label |
|----------|--------|-------|
| 1 | Job for selected Chat run | Current run job |
| 2 | No selected run job | Empty state |

The panel MUST NOT mutate jobs. It does not expose retry, recover, cancel, claim, or lease controls.

## Background Tasks Panel

### Empty State

Trigger:

- no selected Chat run job
- no worker/job events available

Required content:

- No background task is running
- Run a real model message to observe queued jobs and worker events

### Current Task State

Required fields when available:

- run id or readable run label
- job id or readable job label
- job status
- worker diagnostics summary
- latest worker/job events

### Job Status Labels

| Raw State | English | Chinese | Source |
|-----------|---------|---------|--------|
| `queued` | Queued | 排队中 | job status |
| `leased` | Leased by worker | Worker 已领取 | job status |
| `retrying` | Retrying | 重试中 | job status |
| `recovering` | Recovering | 恢复中 | job status or derived from `job_recovering` |
| `completed` | Completed | 已完成 | job status |
| `failed` | Failed | 失败 | job status |
| `cancelled` | Cancelled | 已取消 | job status |
| `dead` | Dead | 已终止 | job status |

## RunRail / Timeline Event Groups

| Group | English | Chinese |
|-------|---------|---------|
| `runtime` | Runtime | 运行时 |
| `worker` | Worker | Worker |
| `job` | Job | 任务 |
| `diagnostics` | Diagnostics | 诊断 |

## M6 Event Labels

| Event | English | Chinese | Severity |
|-------|---------|---------|----------|
| `job_claimed` | Job claimed by worker | Worker 已领取任务 | info |
| `lease_renewed` | Lease renewed | Lease 已续期 | info |
| `job_recovering` | Job recovering | 任务恢复中 | warning |
| `job_retry_scheduled` | Retry scheduled | 已安排重试 | warning |
| `job_attempt_failed` | Job attempt failed | 本次尝试失败 | error |
| `job_retry_exhausted` | Retries exhausted | 重试已耗尽 | error |
| `cancellation` | Cancellation requested | 已请求取消 | warning |
| `worker_diagnostics` | Worker diagnostics | Worker 诊断 | info |
| unknown worker/job event | Unknown worker event | 未知 Worker 事件 | info |

## Timeline Ordering

RunRail/Timeline preserves incoming stream order. It does not reorder events by severity, group, or display category during M6.5 productization. Raw event type remains visible or accessible for debugging.

## Composer State Matrix

| State | Send | Stop/Cancel | Retry | Regenerate | Required text |
|-------|------|-------------|-------|------------|---------------|
| `idle` | enabled if input valid and capability available | hidden | disabled | conditional if previous valid turn exists | Ready |
| `provider_unavailable` | disabled | hidden | disabled | disabled | Provider unavailable |
| `queued` | disabled | enabled if app already supports cancelling queued work | disabled | disabled | Queued |
| `running` | disabled | enabled if app already supports stopping active generation | disabled | disabled | Generating |
| `retrying` | disabled | hidden or disabled | disabled | disabled | Retrying |
| `recovering` | disabled | hidden or disabled | disabled | disabled | Recovering |
| `stopped` | enabled if input valid and capability available | hidden | conditional if previous valid prompt exists | conditional if previous valid turn exists | Stopped |
| `failed` | enabled if input valid and capability available | hidden | enabled until backend retryability is exposed | conditional if previous valid turn exists | Failed |
| `closed` | disabled | hidden | disabled | conditional only when conversation policy allows regenerate | Closed |
| `cancelled` | enabled if input valid and capability available | hidden | conditional if previous valid prompt exists | conditional if previous valid turn exists | Cancelled |

## Disabled Reasons

Required disabled reasons include:

- provider unavailable
- generation in progress
- retry already scheduled
- recovery in progress
- no valid prompt

Non-retryable failure display is deferred until backend retryability is exposed.
