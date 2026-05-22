# goLoomi

goLoomi is an agent platform learning project. It follows the core capability progression of Arkloop at the mechanism level, while keeping its own brand, UI, naming, copy, and implementation structure.

## Goals

- Build a runnable agent platform in small, verifiable milestones.
- Start with clear project boundaries before adding runtime complexity.
- Make each milestone visible and testable, from a desktop-style web shell to API, runs, events, workers, tools, memory, and desktop runtime.
- Learn from public behavior and commit themes without copying private structure, branding, assets, prompts, or non-public interfaces.

## Current Milestone: M0

M0 establishes the project boundary and repository shape.

Planned boundaries:

- `cmd/` — executable entry points.
- `internal/` — private application packages.
- `services/` — independently runnable service boundaries.
- `web/` — desktop-style web UI shell and frontend code.

M0 intentionally does not implement the API, worker, database, desktop runtime, plugin system, or external channels yet.

## Roadmap

High-level milestone sequence:

1. M0 Project boundaries and conventions.
2. M1 Desktop-style web UI shell with mock data.
3. M2 API and database foundation.
4. M3 Authentication, threads, and messages.
5. M4 Runs, events, and SSE.
6. M5 Real web chat timeline.
7. M6+ Model gateway, tools, workers, pipeline, persona, MCP, memory, artifacts, sandbox, admin console, desktop runtime, work mode, channels, plugins, and release tooling.

## Development Principles

- Mechanism parity, original expression.
- Core path first, experience second, complex runtime later.
- Every milestone should run, be testable, and be explainable.
- Prefer Go services and a web UI shell before introducing desktop runtime complexity.
- Do not copy Arkloop branding, logo, visual details, copy, private naming, private prompts, or non-public APIs.

## Repository Status

This repository is currently initialized for M0 only. Source code and runnable services will be added in later milestones.
