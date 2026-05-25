# Research: Tool Result Model Continuation

## Decision: Build continuation context from run-event projection, not persisted tool messages

**Rationale**: Tool lifecycle is already modeled as persisted run events and a tool-call projection. Reusing that audit stream keeps Timeline, SSE replay, and provider continuation aligned. A persisted `messages.role = tool` row would add another source of truth and force chat history to contain implementation-only tool traffic.

**Alternatives considered**:

- Persist `messages.role = tool`: rejected for MVP because chat messages are user/assistant-facing, while tool result is execution context. It may be revisited only if replay or provider compatibility cannot be satisfied from events.
- Store raw provider transcript: rejected because provider payloads are not the Loomi contract and can leak unsafe data.

## Decision: Use an in-memory synthetic tool message for provider adapters

**Rationale**: Most tool-capable providers need a provider-specific item that links a tool result to the original tool call id. Loomi should express that in gateway-neutral terms, then let each adapter serialize it as OpenAI-compatible `tool` content or another provider's equivalent.

**Alternatives considered**:

- Add `role = tool` to public message schema: rejected for MVP because it makes implementation context visible as durable chat data.
- Inline tool result into a system/developer instruction: rejected because it weakens the explicit tool-call/result link and makes prompt injection boundaries blurrier.

## Decision: Keep persistent message role schema unchanged for MVP

**Rationale**: The final user-visible artifact is still an assistant message. Tool context can be reconstructed from `tool_call_requested` and `tool_call_succeeded` events. Keeping message roles unchanged avoids migration churn while Window A is still landing execution APIs.

**Alternatives considered**:

- Extend roles with `tool`: deferred. It is only justified if later multi-turn persistence requires cross-run tool context as first-class chat history.

## Decision: Gateway exposes a continuation request, not a provider-specific transcript

**Rationale**: Runtime code should ask the gateway to continue a run with one redacted tool result. The gateway can then format OpenAI-compatible or future provider calls internally. This keeps provider details out of worker/product data layers.

**Alternatives considered**:

- Worker constructs provider-native message arrays: rejected because it couples runtime execution to one provider family.
- Provider guesses continuation from raw events: rejected because the worker must control loop limits, redaction, and terminal-state decisions.

## Decision: MVP allows one tool call per run

**Rationale**: One approved tool and one continuation call prove the core flow. Multiple tool calls require loop limits, repeated approval UX, context budgeting, and concurrency rules.

**Alternatives considered**:

- Support repeated tool calls immediately: rejected as multi-step agent loop scope.
- Block second tool request without failing the run: rejected because the user would receive an incomplete or misleading answer.

## Decision: If continuation asks for another tool, fail safely

**Rationale**: The model is attempting behavior outside the MVP loop limit. A redacted unsupported-loop failure is clearer and safer than blocking for a second approval path the UI and worker do not yet support.

**Alternatives considered**:

- Enter approval-required again: rejected because it silently adds multi-round tool execution.
- Ignore the tool request and ask the model to answer without it: rejected because it hides provider behavior from Timeline.

## Decision: Denied and tool-failed paths do not call the model again

**Rationale**: Denial is a user safety decision, and tool failure is an execution failure. Calling the model after either path can produce explanations that sound authoritative without tool data and can encourage re-requesting the tool.

**Alternatives considered**:

- Ask the model to explain denial/failure: deferred until Loomi has explicit "model may summarize blocked tool state" UX and policy.
- Persist a templated assistant message: rejected for MVP; the ToolCallCard and terminal run event are enough.

## Decision: Second model phase uses ordinary model events with phase metadata

**Rationale**: SSE already carries ordered persisted run events. The second stream can reuse `model_delta`/`model_final` ordering while metadata such as `model_phase = continuation` lets Timeline and assistantDraft separate phases.

**Alternatives considered**:

- Add a separate continuation SSE endpoint: rejected because it breaks history-first replay.
- Invent new delta event types only for continuation: rejected unless the existing frontend cannot distinguish phases from metadata.

## Decision: Redact result before persistence and before provider continuation

**Rationale**: Once a result is persisted or sent to a provider it is outside the executor boundary. The only safe contract is a small `result_for_model_redacted` shape.

**Alternatives considered**:

- Persist raw and redact on display: rejected because raw secrets would live in storage and logs.
- Send raw to model but display redacted: rejected because provider calls are an external boundary.
