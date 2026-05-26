# Feature Specification: M28 Artifact Runtime Foundation

**Feature Branch**: `[036-artifact-runtime-foundation]`  
**Created**: 2026-05-26  
**Status**: Draft  
**Input**: Continue Arkloop-level coverage after M27 by adding a safe artifact runtime foundation.

## User Scenarios & Testing

### User Story 1 - Create a Safe Text Artifact (Priority: P1)

As a Work mode user, I want an approved tool call to create a bounded text artifact so the agent can preserve generated deliverables as first-class workspace evidence.

**Independent Test**: A Work mode run requests `artifact.create_text`, requires approval, creates one safe artifact record, records safe metadata, and continues the provider with the artifact summary.

### User Story 2 - Read and List Artifacts (Priority: P2)

As a Work mode user, I want the agent to read or list previously created artifacts without exposing raw unsafe payloads in run events.

**Independent Test**: A run creates an artifact, then approved `artifact.list` and `artifact.read` return bounded summaries and content excerpts for artifacts in the same thread/workspace scope.

### User Story 3 - Keep Artifact Runtime Non-executable (Priority: P3)

As a user, I want artifact tools to be visibly non-executable beyond storage so artifact creation cannot become a hidden code/browser/shell runtime.

**Independent Test**: Chat mode omits artifact tools; oversized content, unsupported artifact types, credential-like metadata, denied/stopped/terminal/out-of-scope calls, and missing artifacts fail before creating or reading records.

## Requirements

### Functional Requirements

- **FR-001**: Tool catalog MUST include builtin `artifact.create_text`, `artifact.read`, and `artifact.list`.
- **FR-002**: Artifact tools MUST be Work-mode only and always approval required.
- **FR-003**: `artifact.create_text` MUST create exactly one bounded UTF-8 text artifact from approved arguments.
- **FR-004**: `artifact.read` MUST return a bounded safe excerpt for one existing artifact in scope.
- **FR-005**: `artifact.list` MUST return bounded artifact summaries for the current thread/workspace scope.
- **FR-006**: Artifact tools MUST route through ToolBroker, worker approved-tool resume, provider continuation, and run events.
- **FR-007**: Run events and continuation context MUST include only safe summaries, not raw unbounded content.
- **FR-008**: Settings > Tools and RunRail MUST label artifact tools separately from workspace, sandbox, LSP, web, and browser tools.

### Safety Requirements

- **SR-001**: Artifact tools MUST NOT execute, render, download, compile, open a browser, call shell, call network, or write files.
- **SR-002**: Artifact tools MUST reject unsupported artifact types and oversized content.
- **SR-003**: Artifact metadata MUST redact secret-looking values and local host paths.
- **SR-004**: Chat mode MUST NOT enable artifact tools.

## Non-goals

- Binary artifacts, uploads, downloads, screenshots, rendered previews, live artifact execution, iframe execution, filesystem export, browser integration, shell integration, artifact version graph, marketplace/plugin packaging, and multi-agent orchestration.
