# Data Model: Persona Skill Foundation

## Persona

Durable identity for a behavior configuration.

Fields:

- `id`
- `slug`
- `name`
- `description`
- `source`: built-in for this slice
- `isDefault`
- `isActive`
- `activeVersion`
- `createdAt`, `updatedAt`

Validation rules:

- `slug` is stable and unique per source.
- Exactly one active default built-in persona should exist for fallback.
- Inactive personas cannot be newly resolved for runs unless a run already holds a historical snapshot.

## Persona Version

Versioned body of a persona.

Fields:

- `personaID`
- `version`
- `systemPrompt`
- `modelRoute`: provider/model route label and optional model override
- `allowedToolNames`
- `reasoningMode`
- `budgetSummary`
- `safeSummary`: name, description, model route label, allowed tool names/count, reasoning mode, budget summary
- `createdAt`

Validation rules:

- `systemPrompt` is required but excluded from normal Timeline/debug summaries.
- `allowedToolNames` must be known to Loomi's existing runtime tool registry or rejected by sync/resolution validation.
- `modelRoute` must point to an existing provider/model route label or fail safely before runtime invocation.
- `(personaID, version)` is unique.

## Built-in Persona Config

Repository-local source data synced into durable Persona and Persona Version records.

Fields:

- `slug`
- `name`
- `description`
- `systemPrompt`
- `modelRoute`
- `allowedToolNames`
- `reasoningMode`
- `budgetSummary`
- `version`
- `isDefault`

Lifecycle:

```text
config loaded
-> validate required fields and allowlisted tools
-> upsert persona identity
-> create or update active persona version
-> mark default if configured
```

## Persona Selection

Durable thread/run reference used to resolve a Persona Version.

Fields:

- `threadPersonaID` or `runPersonaID`
- optional `requestedVersion`
- `resolvedPersonaID`
- `resolvedVersion`
- resolution source: run override, thread selection, default built-in

Resolution order:

```text
run override
-> thread selection
-> default built-in persona
-> safe prepare_context failure
```

## Persona Snapshot

Run-scoped record of the exact persona version used.

Fields:

- `runID`
- `personaID`
- `personaSlug`
- `version`
- `name`
- `description`
- `systemPrompt`
- `modelRoute`
- `allowedToolNames`
- `reasoningMode`
- `budgetSummary`
- `resolvedFrom`
- `createdAt`

Rules:

- Snapshot is prepared before provider/runtime invocation.
- Snapshot remains attributable even if the built-in config changes later.
- Prompt text may be used in runtime memory and persisted in a protected snapshot field, but must not be copied into normal Timeline/debug events.

## RunContext Persona Summary

Safe metadata attached to RunContext and pipeline/debug summaries.

Fields:

- `personaID`
- `personaSlug`
- `version`
- `name`
- `description`
- `modelRouteLabel`
- `allowedToolNames`
- `allowedToolCount`
- `reasoningMode`
- `budgetSummary`
- `resolvedFrom`

Forbidden fields:

- raw system prompt
- provider credentials
- raw provider request/response bodies
- raw tool result payloads
- file contents, shell output, browser/desktop captured state
- hidden local state
