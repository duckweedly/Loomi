---
title: Provider Testing Runbook
description: Local M6.5 path for provider readiness, real model messages, worker jobs, and timeline observation.
---

Use this runbook to validate the real local testing path. Settings > Providers can persist one local OpenAI-compatible provider in the product data store; API responses never echo the key.

## Configure provider env

Start the API with the provider environment variables required by the selected provider family:

- provider id or family
- model
- API key through environment variable
- base URL when the provider does not use a default endpoint

Do not commit `.env` files with secrets. Settings > Providers can save one local OpenAI-compatible provider into the product data store and reload it after API restart. If the provider was saved before the persistence migration existed, re-enter the key once.

## Start the API

Start the local API with provider env loaded and the database migrations applied. `/readyz` should report ready.

Expected provider behavior:

- configured providers appear in Settings > Providers
- missing provider env produces a clear empty/unavailable state
- provider errors are sanitized before display

## Start the web app

```bash
VITE_LOOMI_API_BASE_URL=http://127.0.0.1:8080 bun run --cwd web dev
```

## Test provider readiness

Open Settings > Providers.

Expected:

- the provider management toolbar shows search plus All/Enabled/Local/Cloud filters
- configured provider cards show safe display name, route or base URL, Local/Read-only badges when relevant, and no secrets
- Add provider opens the modal, the provider type menu opens, and Save uses the local provider save path
- Test connection enters checking and then connected or failed for configured providers
- failed messages do not display API keys or bearer tokens

## Send a real message

Switch to a real provider-dependent capability such as `real_api` or model gateway.

Expected:

- if provider readiness is available, Composer can send the message
- if provider readiness is unavailable, Composer shows provider unavailable and Open Settings > Providers
- provider unavailable does not show a generating state
- mock mode remains usable without real provider readiness

## Observe worker jobs

Open Background tasks and RunRail/Timeline.

Expected:

- Background tasks shows an empty state when no selected Chat run job evidence exists
- selected Chat run job appears when job evidence is available
- Timeline shows queued, claim, lease, retry, recovery, cancellation, failure, and diagnostics events
- unknown worker events keep raw event type visible

## Failure and recovery validation

Use deterministic fixtures or mocked event streams for automated coverage. Manual real-provider failure testing is useful for smoke validation but should not be the only coverage for retry, recovery, cancellation, failed, or dead states.
