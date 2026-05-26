# Research: M10 Safe Workspace Exec Command

## Decision: argv-only execution

**Rationale**: Running through a shell makes validation unreliable and expands injection risk. `exec.CommandContext` with argv preserves clear boundaries.

**Alternatives rejected**: shell strings, PTY sessions, persistent terminal processes.

## Decision: Reject shell wrappers and destructive first tokens

**Rationale**: Approval is necessary but not sufficient for the first slice. The tool must refuse obvious high-risk operations before execution.

**Alternatives rejected**: pure approval-only policy; broad shell access.

## Decision: Bounded output and timeout

**Rationale**: Run events and UI cannot safely carry unbounded command output or hanging processes.

**Alternatives rejected**: streaming raw output; unlimited background process execution.
