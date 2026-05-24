# Contract: Thread and Message States

## Purpose

Define how thread selection, message history, loading, error, retry, and recovered run states appear across Chat Canvas and Timeline.

## State contract

| State | Trigger | Required UI behavior |
|-------|---------|----------------------|
| No thread | No selected thread | Prompt to select or start a thread; show no stale messages |
| Empty thread | Selected thread has no messages and no run | Invite the user to start a conversation |
| Loading | Thread or message history is loading | Show loading/skeleton state instead of blank history |
| History | Messages loaded and no active runtime state dominates | Show ordered messages for selected thread |
| Error | Thread or message load failed | Show clear error and retry path while preserving selection |
| Backend unavailable | Runtime action cannot be provided by backend | Show capability problem and preserve user-visible context |
| Active run | Selected thread has pending/running/recovering run | Chat Canvas, Composer, and Timeline reflect the same run |
| Terminal run | Latest run completed/failed/stopped/cancelled | Terminal state remains visible until next user action |

## Synchronization rules

- Thread selection controls the messages, latest run, timeline/debug surface, and composer guards.
- Switching threads clears selected-thread presentation state but must not mutate the old thread's run.
- Replayed or late events from an old thread must not update the new selected thread.
- Message load retry keeps the selected thread context.

## Acceptance checks

- No-thread state never shows stale messages.
- Empty-thread state is distinct from loading.
- Loading state is distinct from error.
- Error state offers retry and preserves selected thread.
- Persisted assistant message and related run events agree on the latest run outcome when a thread is selected.
