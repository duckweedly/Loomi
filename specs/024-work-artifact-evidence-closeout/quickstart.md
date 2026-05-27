# Quickstart: M17 Work Artifact Evidence Closeout

## Automated Validation

```sh
go test ./...
bun test --cwd web
bun run --cwd web build
bun run --cwd docs-site build
git diff --check
```

## Local Evidence Seed

Start local API dependencies as usual, then run:

```sh
LOOMI_SEED_SCENARIO=m17-work-artifact go run ./cmd/loomi-seed
```

Record the emitted `thread_id`, `message_id`, `run_id`, and `event_id`.

## Browser Smoke

1. Start the local API.
2. Start web in real API mode.
3. Open the Work mode thread created by the seed.
4. Verify goal, steps, status, artifact references, and recent progress are visible.
5. Verify artifact cards have no executable controls and unsafe metadata is not rendered as actions.
6. Open a Chat mode thread and verify Work Plan View is absent.
7. Record ports, identifiers, screenshot path, and console error status.
