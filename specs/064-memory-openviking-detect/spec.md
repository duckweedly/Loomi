# Feature Spec: M64 Memory OpenViking Detect

## Goal

Settings > Memory should offer the same explicit local-instance detection convenience for OpenViking that Loomi already offers for Nowledge, without enabling providers automatically or exposing secrets.

## User Story

As a user configuring memory, I can select OpenViking, open its configuration modal, click "Detect local instance", and get a safe detected/not-detected result. If detected, the default OpenViking base URL is filled into the draft config.

## Functional Requirements

- Add a safe `GET /v1/memory/provider/openviking/detect` endpoint.
- The endpoint must probe only the default localhost OpenViking API.
- The endpoint must use a short timeout.
- The endpoint response must include only `detected`, optional `base_url`, `message`, and `request_id`.
- Settings > Memory must expose the same detect action for OpenViking as Nowledge.
- A detected OpenViking result may fill `http://127.0.0.1:8282` into the OpenViking base URL draft.

## Non-Goals

- No OpenViking install/start/restart bridge.
- No key discovery, key validation, remote host scanning, or provider auto-enable.
- No change to external memory read/write adapters.
- No brand/copy/visual cloning beyond mechanism parity.

## Success Criteria

- Backend tests cover safe miss behavior and no key leakage.
- Frontend tests cover Settings wiring and real API endpoint mapping.
- Web build and docs build pass.
