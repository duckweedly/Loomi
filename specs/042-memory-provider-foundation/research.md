# Research: Memory Provider Foundation

## Decision: Keep local approved memory as the default provider

**Rationale**: Loomi already has approved memory entries, search, detail, proposal approval, deletion, and audit behavior. The provider foundation should make that current capability explicit instead of replacing it with a semantic store before tools and distillation exist.

**Rejected alternative**: Add an external semantic memory service first. This would pull in embedding keys, health probes, migrations, and failure paths before the user can benefit from agent tools.

## Decision: Model semantic providers as status-capable records, not full adapters

**Rationale**: Settings needs grounded readiness and later slices need provider identity. A status-capable record can represent configured/unconfigured/healthy/unhealthy/degraded safely without pretending read/write integration is complete.

**Rejected alternative**: Implement a full provider interface now. That adds unused abstraction because this slice does not expose memory tools or automatic distillation.

## Decision: Redact diagnostics at the service boundary

**Rationale**: API, UI, run metadata, and tests should all share the same safe projection. Redacting only in the frontend would still leave unsafe backend surfaces.

**Rejected alternative**: Return raw provider errors and hide them in UI. That violates the safety boundary and makes tests less trustworthy.

## Decision: Add readiness metadata without making runs depend on provider health

**Rationale**: Run preparation must continue when memory is disabled, unconfigured, or unhealthy. Later memory tools can inspect readiness and decide whether to operate.

**Rejected alternative**: Fail run preparation when semantic memory is unhealthy. This would make an optional future provider block the core chat/run path.
