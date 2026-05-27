# Research: Memory Agent Tools

## Decision: Use existing ToolBroker path

Memory tools should not bypass approval, run events, or worker continuation. The existing ToolBroker already validates catalog membership, allowlist resolution, approval state, execution state, and safe result summaries.

## Decision: Write creates proposals only

Direct approved memory creation remains service-internal. Tool writes create `memory_write_proposals`, keeping the user's approval boundary intact.

## Decision: Keep all memory tools approval-gated

Even read-only recall can reveal user memory context to a model continuation. Requiring approval makes recall visible in the run timeline while later slices can decide whether some reads become pre-approved.
