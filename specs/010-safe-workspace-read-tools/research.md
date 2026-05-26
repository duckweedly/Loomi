# Research: M8 Safe Workspace Read Tools

## Decision: Reuse M7 tool-call lifecycle

**Rationale**: M7 already provides request, approval, decision, execution, terminal events, worker resume, and UI replay. A separate read API would bypass the audit path.

**Alternatives rejected**: Direct HTTP file browser; hidden agent reads without approval; frontend-only mock reads.

## Decision: Require approval for every M8 read

**Rationale**: Local filesystem reads can expose sensitive user data. Approval keeps the first slice explicit and reviewable.

**Alternatives rejected**: Auto-approve read-only tools; per-session broad permission grant. These can be designed later after the narrow flow is proven.

## Decision: Root-contained relative paths only

**Rationale**: Returning relative paths and resolving all inputs against the workspace root makes leakage easier to test and reason about.

**Alternatives rejected**: Absolute path inputs; home-directory tools; OS-wide search.

## Decision: Deny sensitive paths before read

**Rationale**: Redaction after reading is not enough for credentials. Known secret patterns must be blocked before file content is opened.

**Alternatives rejected**: Best-effort content redaction only; user-visible warnings without enforcement.

## Decision: Bounded summaries, not full dumps

**Rationale**: Run events and UI are audit surfaces, not a bulk file transport. Bounded glob, grep, and read previews keep replay fast and safer.

**Alternatives rejected**: Full file result payloads; unbounded recursive glob or grep.
