# Feature Spec: M63 Memory External Post-Run Commit

## Goal

Align `commit_after_run` with ArkLoop's external provider commit behavior when OpenViking or Nowledge is selected.

## Requirements

- Keep local memory `commit_after_run` behavior as pending proposal creation.
- When OpenViking or Nowledge is selected and configured, commit the post-run assistant outcome through the external provider write adapter.
- Record a safe terminal-run event for external commit success or failure.
- Make external post-run commit idempotent per run.
- Do not expose raw assistant output, provider payloads, credentials, local paths, or content hashes in events.

## Non-Goals

- Do not call real external services in tests.
- Do not add provider-specific background workers in this slice.
- Do not bypass the user-controlled `commit_after_run` toggle.
