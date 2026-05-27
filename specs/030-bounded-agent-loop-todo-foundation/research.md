# Research: M22 Bounded Agent Loop + Todo Foundation

## Decision: Reuse the existing Gateway continuation path

**Rationale**: M21 already proves provider tool-call request, approval, worker execution, tool result persistence, and provider continuation. M22 should remove the hard stop after one continuation while preserving the same event and approval model.

**Alternatives considered**:

- New agent runner: rejected because it would bypass the already-tested Gateway/worker/broker path.
- Full workflow DAG: rejected as platform complexity before the minimal loop works.

## Decision: Sequential one-tool-at-a-time loop

**Rationale**: The current approval UI and tool-call projection are scoped to one pending tool at a time. Sequential execution is enough for the next code-agent slice: read/search one file, continue, read another file, then answer.

**Alternatives considered**:

- Parallel tool calls: rejected because it complicates approval state, ordering, replay, and cancellation.
- Batch approval: rejected because each local action needs explicit approval.

## Decision: Small loop limit as a hard safety boundary

**Rationale**: A bounded loop prevents runaway provider behavior and makes tests deterministic. The exact default belongs in implementation, but the contract requires a small configured maximum and a visible loop-limit failure.

**Alternatives considered**:

- Unlimited loop until provider final: rejected as unsafe and hard to reason about.
- One global process timeout only: rejected because it does not explain why a run stopped.

## Decision: Todo state as safe run metadata, not a new task system

**Rationale**: Work Plan View already projects safe plan/progress metadata from run events. M22 can extend that projection with todo snapshots without adding a durable project/task subsystem.

**Alternatives considered**:

- New `todos` table: deferred until todo state needs cross-run persistence or collaboration.
- Client-only todo state: rejected because refresh/replay would lose the agent plan.

## Decision: No write/shell/browser/web tools in M22

**Rationale**: The user wants Arkloop coverage, but Loomi must move in safe vertical slices. Bounded loop and todo visibility are prerequisites for mutation and execution tools.

**Alternatives considered**:

- Add write/edit immediately: rejected because multi-step approval/replay is still missing.
- Add sandbox shell immediately: rejected because command execution needs a separate isolation plan.
