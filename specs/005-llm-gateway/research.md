# Research: M5 LLM Gateway

## Decision: Use a server-side model gateway with provider-normalized runtime events

**Rationale**: M5 needs real model-backed responses while preserving M4's durable run/event/SSE model and M3.5's frontend runtime boundary. A backend gateway can convert Anthropic, OpenAI, Gemini, and OpenAI-compatible custom provider streams into the same user-visible run events before they reach the web shell.

**Alternatives considered**:

- Frontend calls providers directly: rejected because browser-visible credentials and provider-specific streaming logic would leak into product UI and weaken the existing backend boundary.
- Replace M4 run/event/SSE with provider streams: rejected because Loomi already has observable execution history and SSE recovery semantics that should remain the product contract.
- Add a worker queue immediately: rejected because M6 owns worker/job queue and recoverable background execution; M5 should remain a runnable request-scoped slice.

## Decision: Implement provider clients through Go standard HTTP primitives first

**Rationale**: The repository currently uses Go stdlib plus pgx and the user preference favors existing/native dependencies. Anthropic, OpenAI, Gemini, and OpenAI-compatible custom providers all expose HTTP streaming APIs that can be adapted with request builders, streaming readers, and JSON normalization without adding SDK dependencies in the first slice.

**Alternatives considered**:

- Add official provider SDKs: rejected for M5 planning because each SDK brings different abstractions and retry/error semantics before Loomi's own gateway contract is stable.
- Use a third-party unified LLM library: rejected because it would obscure provider event mapping, credentials, and failure categories that Loomi needs to make observable.

## Decision: Support Anthropic, OpenAI, Gemini, and OpenAI-compatible custom providers

**Rationale**: Clarification confirmed these are product requirements, with custom providers expected to become the user's long-term path. M5 should validate the same basic model-backed response flow against all four provider classes while keeping user-visible states provider-neutral.

**Alternatives considered**:

- One configured provider only: rejected by clarification.
- Full arbitrary provider plugins: rejected because local command/plugin execution belongs to later safety and platform work.

## Decision: Treat custom providers as OpenAI-compatible HTTP endpoints

**Rationale**: OpenAI-compatible chat streaming is the most common interface for custom gateways and hosted proxy services. It keeps custom provider support concrete: configurable base URL, API key, and model, with normalized text delta, completion, refusal, timeout, rate-limit, and failure outcomes.

**Alternatives considered**:

- Generic request/response field mapping: rejected because it introduces a configuration language before there is evidence it is needed.
- Local command adapters: rejected because command execution changes safety and sandbox boundaries and belongs outside M5.

## Decision: Manage provider configuration outside the product UI for M5

**Rationale**: Clarification selected local configuration only. This keeps M5 focused on gateway capability and runtime visibility rather than a settings product surface. The web UI can display capability status returned by the backend without allowing provider editing.

**Alternatives considered**:

- Minimal settings UI: rejected because it expands M5 into settings management and validation UX.
- Hybrid UI for custom providers only: rejected because it still creates product configuration behavior before the gateway contract is stable.

## Decision: Limit model request context to the current thread

**Rationale**: Clarification selected current thread only. The gateway should include the current user message plus necessary recent messages from the same conversation. Broader context, attachments, RAG, memory, and pipeline context are later roadmap capabilities.

**Alternatives considered**:

- Current message only: rejected because basic conversational continuity would be too weak for a chat runtime.
- All available Loomi context: rejected because it pulls Pipeline/Context concerns forward and increases data-boundary risk.

## Decision: Persist assistant messages and model-gateway runs through a schema migration

**Rationale**: M3 messages currently allow only `user` role, and M4 runs constrain `source` to `local_simulated`. M5 requires final assistant messages and real model-backed run history, so migration `000004` should extend the existing tables instead of inventing parallel storage.

**Alternatives considered**:

- Store assistant output only in run events: rejected because completed assistant responses should appear in durable thread message history.
- Add separate assistant_drafts table in M5: rejected because drafts can be represented as run events until the worker/recovery model requires stronger persistence.

## Decision: Map provider streams to existing event categories with new event types

**Rationale**: The M4 schema already supports `lifecycle`, `progress`, `message`, `error`, and `final` categories plus flexible `type`, `content`, and `metadata`. M5 can add event types such as `model_request_started`, `model_output_delta`, `model_output_completed`, `provider_rate_limited`, and `tool_call_blocked` without changing event categories.

**Alternatives considered**:

- Add new event categories for model/provider/tool: rejected for M5 because it requires schema and UI grouping changes beyond the minimum needed for gateway observability.
- Store provider-native event names as product event types: rejected because it would expose provider differences directly to Loomi's user-facing model.

## Decision: Tool-like provider output becomes a non-executed boundary event

**Rationale**: The spec requires tool-like model requests to be visible but not executed. M5 should detect provider-native tool call/function call content when available and record a `tool_call_blocked` style event without performing external actions.

**Alternatives considered**:

- Ignore tool-like output: rejected because users need to see why no tool action happened.
- Execute provider tool calls: rejected because tool permissions, audit, and execution belong to later milestones.

## Decision: Redact provider errors into visible failure states

**Rationale**: Clarification selected redacted visible states. Provider errors, timeouts, and rate limits should map to stable error codes and user-safe summaries in run events, while raw provider details stay out of user-visible history.

**Alternatives considered**:

- Generic failure only: rejected because it would make local validation and debugging too opaque.
- Developer-only details: rejected for M5 because the timeline/debug surface is part of the learning product surface.
