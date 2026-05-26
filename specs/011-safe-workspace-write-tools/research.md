# Research: M9 Safe Workspace Write Tools

## Decision: Add write_file before shell execution

**Rationale**: Code-agent usefulness improves once the agent can persist a safe text file change. Shell execution is higher risk and should follow after write boundaries are proven.

**Alternatives rejected**: Jumping directly to `exec_command`; broad patch application; OS-level file access.

## Decision: Exact replacement edit only

**Rationale**: `old_text` plus `new_text` with exactly one match is deterministic, easy to inspect, and easy to fail without mutation.

**Alternatives rejected**: Unified diff parser, fuzzy matching, regex replacement, multi-hunk edits.

## Decision: Existing parent directory required

**Rationale**: Creating directories expands write scope and can hide surprising filesystem changes. Directory creation can be a later explicit tool.

**Alternatives rejected**: Auto-create parent directories; broad project scaffolding from one tool call.

## Decision: Text-only bounded content

**Rationale**: Runtime events and user review are built around bounded text summaries. Binary writes and huge payloads are not needed for the first mutation slice.

**Alternatives rejected**: Binary file writes; base64 payloads; unbounded content persistence.
