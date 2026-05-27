# Feature Spec: M47 Memory Provider Config Details

## Goal

Align Loomi's memory provider configuration shape with the target mechanism: local memory remains the safe default, while Settings and `/v1/memory/provider` can configure Nowledge and OpenViking without leaking secrets.

## Requirements

- Support provider ids `local`, `nowledge`, `openviking`, and the existing legacy `semantic` placeholder.
- Persist OpenViking base URL, root key, embedding model settings, VLM model settings, and optional rerank settings.
- Persist Nowledge base URL, optional API key, and request timeout.
- Return only safe provider status: key presence booleans are allowed; raw keys, provider traces, Authorization headers, and credential-bearing endpoints are not.
- Settings > Memory must expose Local / Nowledge / OpenViking provider selection and the same model configuration groups.
- `commit_after_run` remains an approval-gated proposal toggle and must not auto-approve memories.

## Non-goals

- No external OpenViking or Nowledge read/write adapter execution in this slice.
- No snapshot/impression rebuild endpoints.
- No notebook dual-provider system.
- No activity recorder ingestion.
