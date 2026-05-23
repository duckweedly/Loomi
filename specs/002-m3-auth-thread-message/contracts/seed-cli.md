# Seed CLI Contract: M3 Local Demo Data

M3 provides an explicit seed command for local demo data. Migrations must not insert demo threads or messages.

## Command

From the repository root, after applying M3 migrations:

```bash
go run ./cmd/loomi-seed
```

Required environment:

```bash
DATABASE_URL=postgres://loomi:loomi@127.0.0.1:55433/loomi_m2?sslmode=disable
LOG_LEVEL=info
```

## Behavior

The command must:

1. Load and validate local configuration using the same redaction rules as the API.
2. Ensure the fixed local development user exists.
3. Upsert one deterministic demo thread owned by the local user.
4. Upsert one deterministic user-authored message in that thread using a deterministic `client_message_id`.
5. Emit structured diagnostics with an operation id.

The command must not:

- Create assistant messages.
- Create run events, streaming deltas, tool calls, model outputs, worker jobs, or LLM requests.
- Delete existing local developer data.
- Print secrets or the full `DATABASE_URL`.

## Idempotency

Running the command multiple times must leave one copy of the seeded thread and one copy of the seeded message.

Expected repeated-run behavior:

```text
first run  -> creates or updates local user, creates demo thread, creates demo user message
second run -> local user exists, demo thread exists, demo user message exists, no duplicate message
```

## Success Output

The command should emit a final structured diagnostic equivalent to:

```json
{
  "level": "info",
  "component": "seed",
  "operation_id": "seed_000000000000000000000000",
  "message": "m3 seed complete",
  "thread_id": "thr_local_demo",
  "message_id": "msg_local_demo_001"
}
```

The exact operation id is generated at runtime.

## Failure Output

Failures should emit a structured diagnostic with a redacted error and non-zero exit code:

```json
{
  "level": "error",
  "component": "seed",
  "operation_id": "seed_000000000000000000000000",
  "message": "m3 seed failed",
  "error": "[redacted]"
}
```
