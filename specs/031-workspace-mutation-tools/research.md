# Research: M23 Workspace Mutation Tools

## Decision: Reuse M21 Workspace Root and Path Policy

**Decision**: `workspace.write_file` and `workspace.edit` reuse the same workspace root resolution, relative path normalization, symlink escape rejection, sensitive path denylist, and UTF-8 text boundary used by M21 read tools.

**Rationale**: Mutation tools should not widen the file access policy beyond the already validated read scope. Reusing the root and denylist keeps behavior explainable and reduces the chance of read/write policy divergence.

**Alternatives considered**:

- Separate mutation root: rejected because it adds configuration complexity before the existing single-root policy is exhausted.
- Allow absolute host paths with approval: rejected because approval does not make path scope safe or auditable.

## Decision: New File Creation Before Overwrite

**Decision**: `workspace.write_file` creates new text files only. Existing targets fail unless a later feature introduces explicit overwrite semantics.

**Rationale**: New file creation is the narrowest useful write primitive and avoids accidental loss of existing user work.

**Alternatives considered**:

- Overwrite by default: rejected because it is too destructive for the first mutation slice.
- Append mode: rejected because it is less generally useful than create/edit and creates ambiguous duplicate behavior.

## Decision: Exact Replacement for First Edit Tool

**Decision**: `workspace.edit` accepts exact `old_text` and `new_text` replacement and applies it only when `old_text` occurs exactly once.

**Rationale**: Exact replacement is easy to audit, deterministic, and testable without adding patch parsers or AST tooling. Requiring a single match prevents ambiguous edits.

**Alternatives considered**:

- Unified diff patch: rejected for M23 because hunk parsing/conflict behavior expands the safety surface.
- Regex replacement: rejected because it increases accidental broad-change risk.
- AST edits: rejected because it is language-specific and too broad for the first mutation slice.

## Decision: Safe Metadata Instead of Full Content Persistence

**Decision**: Run events and tool projections persist relative path, operation, byte/line counts, changed flag, and preview/redaction flags, not raw file contents.

**Rationale**: Tool arguments and results may contain source or secrets. The audit trail should prove what class of mutation happened without making event history a file-content store.

**Alternatives considered**:

- Persist full diff text in events: rejected because source and secrets can leak through event replay.
- Persist only success/failure: rejected because operators need enough metadata to understand the mutation.

## Decision: Keep Shell/Exec Out of M23

**Decision**: M23 stops at workspace file mutation and does not include `exec_command`.

**Rationale**: Command execution requires sandbox, streaming output, timeout, cancellation, and resource limits. It should follow mutation tools rather than share this first write slice.

**Alternatives considered**:

- Add write and exec together: rejected because it combines two different risk classes and weakens validation.
