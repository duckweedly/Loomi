# Research: M15 Chat Real Integrated Smoke Closeout

## Decision: Use a gated Go integration smoke as the authoritative M15 evidence

**Rationale**: The required path spans HTTP handlers, productdata service/repository, runtime provider, worker, MCP execution, memory snapshot, and event replay. A Go smoke can exercise these boundaries without relying on UI-only mocks or external model calls.

**Alternatives considered**: Browser-only smoke was rejected because it can pass while backend execution is mocked or unavailable. External provider smoke was rejected because it is non-deterministic and can spend money.

## Decision: Reuse deterministic provider fixtures

**Rationale**: M15 is a closeout/evidence slice. A provider fixture can deterministically request one MCP tool, then emit one final assistant message after redacted tool result continuation.

**Alternatives considered**: Live model calls were rejected for repeatability. Adding a new provider abstraction was rejected because existing provider test seams are sufficient.

## Decision: Reuse existing M7/M9/M11/M12/M13 boundaries

**Rationale**: The goal is integrated evidence, not new platform behavior. The smoke should compose existing approval, RunContext, MCP discovery, worker execution, continuation, memory, and replay paths.

**Alternatives considered**: A new queue/smoke harness or sandbox was rejected as out of scope and contrary to the M15 non-goals.

## Decision: Use explicit sensitive canaries for redaction assertions

**Rationale**: Redaction is only convincing if the fixture includes known secret-looking values and checks all shareable evidence surfaces for absence.

**Alternatives considered**: Spot-checking a single event was rejected because leaks can occur in API response, RunContext safe summary, tool result summary, replay, or docs independently.
