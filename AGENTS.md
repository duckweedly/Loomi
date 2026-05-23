<!-- SPECKIT START -->
For Spec Kit workflows, use the project skills in `.agents/skills/`.

Always read `.specify/memory/constitution.md` before non-trivial Loomi work and follow it over ad-hoc preferences. Documentation is part of done: when changing code, architecture, workflow, APIs, data models, runtime behavior, tools, workers, UI flows, or safety boundaries, update the Starlight documentation site under `docs-site/src/content/docs/` in the same work session. Do not wait for a separate user reminder.

Use these documentation targets:
- `architecture/` for module boundaries, state flows, event models, permissions, and observability.
- `api/` for endpoints, event payloads, schemas, and compatibility notes.
- `runbooks/` for commands, environment variables, local setup, validation, and troubleshooting.
- `adr/` for durable technical decisions and trade-offs.
- `devlog/` for completed work, validation results, known limitations, and next steps.
- `spec-kit/` for links or summaries of feature specs, plans, and tasks.

Use Bun for the docs site. Before claiming completion, run the relevant code validation and, when docs changed, run `bun run build` from `docs-site/`. If validation cannot run, report the exact reason.
<!-- SPECKIT END -->
