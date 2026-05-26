# Research: M24 Sandbox Exec Command Tools

## Decision: Use argv-form execution only

**Rationale**: `sandbox.exec_command` should not interpret shell strings in its first slice. An argv array avoids implicit shell expansion, command chaining, redirects, and quoting ambiguity.

**Alternatives considered**:
- Shell string execution: rejected because it expands scope and risk.
- Persistent terminal session: rejected because it needs streaming, cancellation, session state, and stronger sandbox controls.

## Decision: Workspace root is the cwd boundary

**Rationale**: M21/M23 already established workspace root resolution, relative path validation, and sensitive path denial. Reusing that boundary keeps the command slice testable and avoids introducing container orchestration prematurely.

**Alternatives considered**:
- Host current directory: rejected because it is implicit and harder to audit.
- Docker/container sandbox now: deferred to a later slice because it needs setup, image policy, volume policy, and cleanup semantics.

## Decision: Deny destructive commands in the first slice

**Rationale**: The user’s global safety rules require explicit care around destructive operations. The product tool should start by rejecting obvious destructive command patterns before spawn, even if a model requests them.

**Alternatives considered**:
- Rely only on approval: rejected because accidental model-generated destructive commands are too high risk.
- Full policy engine: deferred until more command categories exist.

## Decision: Capture bounded output after completion

**Rationale**: The existing tool result path records a single result summary. Bounded post-completion stdout/stderr previews are enough for the first vertical slice.

**Alternatives considered**:
- Streaming stdout/stderr events: useful but requires additional event contracts and UI states; scheduled for a later sandbox expansion.

## Decision: Event metadata uses previews

**Rationale**: Tool-call request events need to show what would run, but should not expose env values, absolute roots, provider payloads, or unbounded command output.

**Alternatives considered**:
- Persist raw args/output everywhere: rejected by Loomi safety constitution.
