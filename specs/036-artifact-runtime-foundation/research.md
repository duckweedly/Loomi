# Research: M28 Artifact Runtime Foundation

## Decision: Start With Text-only Artifact Storage

Use `artifact.create_text`, `artifact.read`, and `artifact.list` as the first artifact runtime slice.

**Rationale**: Text artifacts prove the agent can create durable deliverables without introducing binary storage, rendering, downloads, or execution risk.

**Alternatives rejected**:

- Rendered HTML artifacts first: higher XSS/runtime surface.
- File export first: overlaps with workspace write tools and filesystem permissions.
- Binary artifacts first: needs storage limits, MIME validation, preview rules, and download UI.

## Decision: Reuse ToolBroker and Worker Approval

Artifact tools are normal builtin tools, not a separate agent output path.

**Rationale**: Keeps approval, run events, loop limits, provider continuation, and Settings/RunRail visibility consistent with workspace/sandbox/LSP/web/browser tools.

## Decision: Safe Summaries in Events

Run events store title/type/size/excerpt/truncation/source ids only.

**Rationale**: Keeps event replay useful without persisting raw unbounded content in the timeline.
