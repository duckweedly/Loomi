# Quickstart: Local Real Provider Testing

This runbook verifies the M6.5 real testing path: provider readiness, real Chat generation, worker job state, timeline events, cancellation/recovery/failure visibility.

## 1. Configure provider environment

Start from the project’s supported provider environment variables.

Required values depend on the configured provider family. At minimum, local testing needs:

- provider family / provider id
- model name
- API key supplied through environment variable
- base URL if the provider does not use the default endpoint

Rules:

- Do not paste API keys into Settings local draft for persistence.
- Do not commit `.env` files containing secrets.
- Do not expect local draft provider fields to affect real model calls.

## 2. Start API with provider env

Start the API process with provider env loaded.

Expected result:

- API boots without exposing secret values.
- Provider readiness endpoint/source can discover configured providers.
- Missing or invalid env results in user-readable provider unavailable state.

## 3. Start web app

Start the web app against the local API.

Expected result:

- Chat and Settings can reach backend readiness state.
- Mock mode remains available even if provider env is missing.

## 4. Open Settings > Providers

Navigate to Settings > Providers.

Expected result:

- Provider Test Console is visible.
- Configured providers section shows backend providers.
- Local draft section is clearly labeled as browser-session-only and unsaved.
- If no providers are configured, an env setup empty state is shown.

## 5. Test provider connection

Click Test connection for a configured provider.

Expected result:

- Provider status changes to checking.
- Success path shows success message.
- Failure path shows readable sanitized error.
- API key is never displayed.

## 6. Send a real model message

Switch runtime/backend capability to a provider-dependent real mode such as `real_api` or `model_gateway`.

Send a simple message.

Expected result:

- If provider is available, message enters queued/running state and eventually completes or fails clearly.
- If provider is unavailable, Composer shows provider unavailable and Open Settings > Providers CTA.
- UI does not show misleading generation while provider is unavailable.

## 7. Observe Background tasks

Open the right-side Background tasks panel.

Expected result:

- Selected Chat run job appears when job evidence is available.
- Empty state appears when no selected Chat run job evidence exists.
- queued/leased/retrying/recovering/completed/failed/cancelled/dead states are readable.
- Worker diagnostics summary appears when available.
- Latest worker/job events are listed.

## 8. Observe RunRail / Timeline

Open the run timeline.

Expected result:

Timeline displays readable labels for:

- job_claimed
- lease_renewed
- job_recovering
- job_retry_scheduled
- job_attempt_failed
- job_retry_exhausted
- cancellation
- worker diagnostics
- unknown worker/job event fallback

Raw event type remains visible or accessible for debugging. Event display preserves incoming stream order.

## 9. Validate cancel/recovery/failure states

Use local test conditions, deterministic fixtures, or mocked event streams to trigger:

- queued cancellation
- running cancellation
- provider failure
- retry scheduling
- recovery after lease interruption
- retry exhausted / dead state

Expected result:

- Composer status and buttons match each state.
- Background tasks panel shows the same state consistently.
- Timeline shows the corresponding worker/job event.
