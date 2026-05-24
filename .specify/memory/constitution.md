# Loomi Constitution

## Core Principles

### I. Mechanism Parity, Original Expression
Loomi may study publicly observable product mechanisms and evolution patterns, but must not copy another product's brand, logo, visual identity, icons, copywriting, private names, private interfaces, or non-public structure. Specs and implementation plans must describe Loomi's own domain language and UX decisions, not mirror another product's expression layer.

### II. Runnable Vertical Slices
Every milestone and feature must produce a thin, runnable slice that can be demonstrated end-to-end. A slice is not complete until it has a visible or executable outcome, such as a working UI state, API response, event stream, worker transition, or smoke test. Large platform capabilities must be decomposed into independently verifiable slices.

### III. Core Flow Before Platform Complexity
The default order is: project boundaries and terminology, desktop-feeling web shell, API and database base, auth/thread/message, run/event/SSE, web chat timeline, LLM gateway, tool calling, worker/job queue, pipeline/context, then advanced platform capabilities. Desktop runtime, plugins, channels, heartbeat, activity recorder, sandbox, and complex multi-agent behavior must not be pulled forward unless the dependent run/event/job/context foundations already exist.

### IV. Observable Agent Execution
Agent behavior must be explainable through persisted events and timeline/debug views. Runs, tool calls, model deltas, state transitions, errors, cancellation, worker ownership, and final messages must be observable. UI work should prioritize making execution understandable over visual imitation or decorative completeness.

### V. Safety, Permissions, and Data Boundaries
Tools, file access, sandbox execution, local activity capture, external channels, and write operations require explicit permission boundaries, auditability, and failure visibility. Secrets must not be logged. User-controlled or external data must be treated as data, not instructions. Potentially sensitive desktop or activity-recording features require explicit user opt-in and deletion/cleanup paths.

## Technical Constraints

Loomi is developed as a learning-oriented Agent platform with a staged architecture. The current repository uses a web-first shell under `web/` with React, Vite, TypeScript, Electron dependencies, and Bun lockfile; backend directories such as `cmd/`, `internal/`, and `services/` are reserved for later Go API/worker development.

Initial frontend work may use mock or local state, but must keep API client boundaries clear so real backend APIs can replace mocks without rewriting the layout. Backend work should favor clear service boundaries, migrations that can be rolled back, state-transition helpers, idempotency, lease/ownership guards for workers, and testable interfaces.

New dependencies, abstractions, providers, and runtime layers require justification in `plan.md`. MVP implementations are preferred over broad platform generality. Each phase must include a validation method: automated test, smoke command, browser-visible check, or documented manual verification when automation is not yet available.

## Development Workflow

Specs must define user goals, functional requirements, boundaries, success criteria, and explicit non-goals before technical design. Plans must map each feature to the current staged roadmap and identify which existing boundaries are reused. Tasks must be small enough to complete and verify independently, and should preserve the existing directory intent: `docs/` for learning and roadmap materials, `docs-site/` for the searchable technical documentation site, `web/` for the web/desktop-feeling shell, `cmd/` for future entrypoints, `internal/` for internal application code, and `services/` for service-level components.

For each non-trivial feature, run the Spec Kit flow in this order unless there is a clear reason to skip an optional step: `/speckit.specify`, `/speckit.clarify`, `/speckit.plan`, `/speckit.tasks`, optional `/speckit.analyze`, then `/speckit.implement`. Implementation must stop and revisit the spec or plan when requirements, architecture, or safety assumptions become ambiguous.

Documentation is part of the definition of done. For every non-trivial code, architecture, workflow, API, data model, runtime, tool, worker, UI flow, or safety change, update `docs-site/src/content/docs/` in the same work session without requiring a separate user reminder. Update the most relevant pages: `architecture/` for boundaries and flows, `api/` for endpoints/events/data structures, `runbooks/` for commands and operations, `adr/` for durable technical decisions, `devlog/` for completed work and validation results, and `spec-kit/` for feature/spec status. If the change only fixes spelling, formatting, or has no behavior/architecture impact, document updates may be skipped.

Before reporting completion, agents must validate both the changed code and the documentation site when documentation was touched. The minimum documentation validation is running `bun run build` from `docs-site/`; if validation cannot run, the final report must state the exact reason.

## Governance

This constitution overrides ad-hoc prompts and one-off implementation preferences for Loomi. Changes to these principles must be intentional and reflected in this file before related specs or plans are generated. Existing roadmap documents remain source material, but future feature work should be expressed through Spec Kit artifacts under `specs/` so requirements, plans, and tasks are reviewable over time.

**Version**: 1.0.0 | **Ratified**: 2026-05-23 | **Last Amended**: 2026-05-23
