# Plan: M58 Memory Prompt Notebook Snapshot

1. Add a notebook snapshot field to `RunContext` and safe run summaries.
2. Build the notebook snapshot during in-memory and Postgres run-context preparation.
3. Inject safe `<memory>` and `<notebook>` blocks into the gateway system prompt.
4. Add tests for RunContext notebook loading and prompt block formatting.
5. Update memory architecture/API/runbook/devlog documentation.
