# Implementation Plan: M11 Tool Catalog Visibility

**Branch**: `013-tool-catalog-visibility` | **Date**: 2026-05-26 | **Spec**: [spec.md](./spec.md)

## Summary

M11 exposes the current allowlisted tools through a read-only backend catalog and replaces the Settings > Tools placeholder with a real read-only panel. It does not add permission editing or execution controls.

## Technical Context

**Language/Version**: Go 1.23 backend; TypeScript/React/Vite frontend; Bun docs/frontend.

**Primary Dependencies**: Existing runtime tool definitions, HTTP API, real/mock API clients, SettingsView. No new dependency required.

**Storage**: Static in-process catalog. No migration.

**Testing**: TDD required. Backend HTTP tests for deterministic catalog and redaction. Frontend realApiClient tests for mapping. SettingsView tests for catalog rendering and no placeholder copy.

## Constitution Check

- **Runnable Vertical Slice**: PASS. API and Settings UI are directly visible.
- **Core Flow Before Platform Complexity**: PASS. Visibility only; no new execution layer.
- **Observable Agent Execution**: PASS. The catalog clarifies available execution surfaces.
- **Safety/Data Boundaries**: PASS. Read-only catalog; no secrets or raw payloads.
- **Documentation Definition of Done**: PASS.
