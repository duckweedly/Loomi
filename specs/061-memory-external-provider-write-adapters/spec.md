# Feature Spec: M61 Memory External Provider Write Adapters

## Goal

Route approved external-provider memory mutations through the configured OpenViking or Nowledge provider.

## Requirements

- Route `memory.write` to OpenViking session/message/commit when OpenViking is selected and configured.
- Route `memory.edit` for `viking://...` URIs to OpenViking content replacement.
- Route `memory.forget` for `viking://...` URIs to OpenViking delete.
- Route `memory.write` to Nowledge `/memories` when Nowledge is selected and configured.
- Route `memory.forget` for `nowledge://memory/...` URIs to Nowledge delete.
- Keep mutation tool results safe-summary-only and omit raw content, provider traces, credentials, local paths, and content hashes.

## Non-Goals

- Do not add Nowledge edit semantics; provider-aware tool availability keeps `memory.edit` disabled for Nowledge.
- Do not execute real external provider writes in tests.
- Do not bypass Loomi tool approval.
