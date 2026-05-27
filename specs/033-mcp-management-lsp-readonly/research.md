# Research: M25 MCP Management + LSP Read-only Foundation

## Decision: MCP Management Is Read-only Over Existing Local Config and Discovery Events

**Rationale**: Loomi already accepts explicit local stdio MCP config through the runtime config path and records safe discovery metadata in run events. A read-only management surface gives users visibility without introducing config mutation, secret handling, remote OAuth, marketplace install, or restart semantics.

**Alternatives considered**:

- Add CRUD endpoints for MCP servers: rejected for M25 because it requires secret storage, validation UX, lifecycle restart semantics, and broader safety review.
- Productize remote MCP/OAuth: rejected because local stdio MCP is the existing implemented boundary.

## Decision: LSP Foundation Uses Bounded Read-only Workspace Analysis

**Rationale**: The next Arkloop parity layer is code intelligence after glob/grep/read/write/edit/exec. A bounded read-only implementation can prove tool catalog, approval, worker, event, UI, and safety semantics without spawning long-lived language servers or package-manager diagnostics.

**Alternatives considered**:

- Spawn real language servers: deferred because it introduces process lifecycle, project-specific setup, performance, and trust boundaries.
- Use sandbox exec to run language commands: rejected because LSP tools must stay read-only and must not depend on shell execution in this slice.

## Decision: LSP Tools Are Builtin, Work-mode-only, Approval-required, Low-risk/read-only

**Rationale**: LSP tools inspect local source files and should remain explicitly approved while the tool surface is new. They are lower risk than write/edit/exec but still access workspace content, so Chat mode must not enable them.

**Alternatives considered**:

- Make LSP tools approval-free: rejected because workspace access should remain explicit until tool policies are configurable.
- Make them MCP tools: rejected because LSP is a builtin code-intelligence surface, while MCP remains user-configured external tools.
