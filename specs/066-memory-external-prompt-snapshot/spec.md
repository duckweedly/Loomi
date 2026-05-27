# Feature Spec: M66 Memory External Prompt Snapshot

## Goal

External memory providers should contribute safe memory context before the initial model request, matching the reference pipeline pattern where provider recall can be injected into the prompt without waiting for an explicit tool call.

## User Story

As a user with OpenViking or Nowledge configured, when a run starts, Loomi searches the selected external memory provider using the latest user message and injects safe summaries into the `<memory>` prompt block.

## Functional Requirements

- Before the initial model provider request, enrich the prepared run context with external memory hits when an external provider is active and configured.
- Use the latest user message as the bounded recall query.
- Preserve the existing `<memory>` prompt block format and safe summary projection.
- Do not mutate the original prepared run context in place.
- If the provider search fails or returns no hits, continue with the existing prepared context.

## Non-Goals

- No background snapshot cache.
- No recursive OpenViking tree snapshot.
- No LLM distillation or impression refresh.
- No raw provider payloads, credentials, or provider traces in prompts.

## Success Criteria

- Runtime test proves an OpenViking search hit appears in the initial system prompt memory block.
- Existing local memory and notebook prompt behavior remains unchanged.
