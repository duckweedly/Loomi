# Feature Specification: M5 LLM Gateway

**Feature Branch**: `[005-llm-gateway]`

**Created**: 2026-05-23

**Status**: Draft

**Input**: User description: "M5"

## Clarifications

### Session 2026-05-23

- Q: What context may be included in model requests for M5? → A: Current thread only.
- Q: How should provider errors, timeouts, and rate limits appear in M5 execution history? → A: Redacted visible states.
- Q: What provider scope must M5 support? → A: Anthropic, OpenAI, Gemini, and custom providers.
- Q: What custom provider interface must M5 support? → A: OpenAI-compatible HTTP.
- Q: How should provider configuration be managed in M5? → A: Local configuration only.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Receive a model-backed assistant response (Priority: P1)

As a Loomi user working in an existing conversation, I want my submitted message to produce an assistant response from a real model-capable execution path so the product moves beyond deterministic mock output while preserving the same conversational flow.

**Why this priority**: This is the core M5 value: turning the existing thread, message, run, and timeline foundation into a real model-backed slice without pulling in later platform complexity.

**Independent Test**: Can be fully tested by submitting a message in a model-capable environment and verifying that the conversation shows incremental assistant output, a final assistant response, and a completed execution history for that request.

**Acceptance Scenarios**:

1. **Given** an existing conversation and a model-capable environment, **When** the user submits a message, **Then** the user sees the message accepted, the assistant response begin, and the final assistant answer appear in the same conversation.
2. **Given** a model-backed response is active, **When** output is produced incrementally, **Then** the user can see the assistant response grow without losing the ability to inspect execution progress.
3. **Given** a model-backed response completes, **When** the user reviews the conversation and execution history, **Then** the final assistant response is clearly linked to the original user message.

---

### User Story 2 - Understand model execution progress and failures (Priority: P2)

As a user or developer validating Loomi, I want model execution progress, cancellation, and failure states to be visible and understandable so I can trust what happened during a run and recover when something fails.

**Why this priority**: Loomi's constitution requires observable agent execution. M5 must preserve explainability while replacing mock output with real model behavior.

**Independent Test**: Can be fully tested by running successful, stopped, unavailable, and failed model-response attempts and verifying that each outcome has a distinct visible state and leaves the conversation usable.

**Acceptance Scenarios**:

1. **Given** a model-backed response is active, **When** the user stops it, **Then** the response reaches a canceled terminal state and the conversation remains available for another message.
2. **Given** the model-capable environment is unavailable or misconfigured, **When** the user submits a message that requires model execution, **Then** the user sees an explicit unavailable state instead of a simulated successful response.
3. **Given** model execution starts but fails before completion, **When** the failure is shown, **Then** the user sees a clear failure state and can continue using the conversation.

---

### User Story 3 - Preserve safety boundaries for future tool use (Priority: P3)

As a user, I want Loomi to distinguish model text generation from tool execution so that model output cannot silently perform external actions before tool permissions and audit rules are ready.

**Why this priority**: The M5 roadmap mentions tool-calling protocol work, but current project status keeps actual tool execution deferred. The slice must create a safe boundary rather than silently introducing side effects.

**Independent Test**: Can be fully tested by using a prompt that causes the model to request or describe a tool-like action and verifying that no external action is executed and the user can see that tool execution is unavailable in this slice.

**Acceptance Scenarios**:

1. **Given** a model output requests a tool-like action, **When** the request appears during a response, **Then** Loomi does not execute the action and shows that tool execution is outside the current capability boundary.
2. **Given** a model-backed response includes sensitive operational details such as provider errors, **When** those details are displayed, **Then** secrets and credentials are not exposed to the user-facing history.

---

### Edge Cases

- What happens when the model-capable environment is not configured, unavailable, rejects a request, times out, or reaches a rate limit?
- What happens when output begins but the model response fails before a final answer is available?
- What happens when the user stops a response while output is still arriving?
- What happens when the user switches conversations while a model-backed response is active?
- What happens when the model returns an empty, refused, or policy-limited response?
- What happens when model output requests a tool-like action before tool execution is in scope?
- What happens when repeated output fragments would otherwise create duplicate or confusing assistant text?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Users MUST be able to submit a message in an existing conversation and start a model-backed assistant response when the environment is model-capable.
- **FR-002**: The system MUST show assistant output incrementally while a model-backed response is active and show a final assistant message when it completes.
- **FR-003**: The system MUST provide a visible execution history for each model-backed response, including request accepted, execution started, output in progress, completed, failed, and canceled states.
- **FR-004**: The system MUST clearly link each model-backed assistant response and execution history to the user message that started it.
- **FR-005**: Users MUST be able to stop an active model-backed response and see a canceled terminal state.
- **FR-006**: The system MUST show an explicit unavailable or misconfigured state when model capability is not available, and MUST NOT present simulated output as a real model response.
- **FR-007**: The system MUST show understandable user-facing states for model failures, rejected requests, empty responses, interrupted responses, provider errors, timeouts, and rate limits while keeping the conversation usable.
- **FR-008**: The system MUST avoid exposing secrets, credentials, or raw sensitive provider details in user-facing conversation history or execution history; provider failures MUST be represented with user-safe messages.
- **FR-009**: The system MUST treat tool-like model requests as non-executed capability-boundary events in this feature and MUST NOT perform external actions through tools.
- **FR-010**: The system MUST allow a developer or operator to verify whether a local environment is ready for model-backed execution before relying on the feature in a demo or validation session.
- **FR-011**: Model-backed responses MUST limit request context to the current user message plus necessary recent messages from the same conversation.
- **FR-012**: The system MUST support Anthropic, OpenAI, Gemini, and custom provider configurations for model-backed responses while keeping user-visible behavior provider-neutral.
- **FR-013**: Custom provider configurations MUST use an OpenAI-compatible HTTP chat and streaming interface with configurable endpoint, credential, and model values.
- **FR-014**: Provider configuration for M5 MUST be managed outside the product UI through local development configuration.

### Key Entities

- **Model-Backed Response**: A single assistant answer produced for a user message through a real model-capable path; includes active, completed, failed, and canceled outcomes.
- **Execution History**: The chronological user-visible record of what happened during a model-backed response, including progress and terminal states.
- **Output Segment**: A visible portion of assistant text that arrives before the final response is complete.
- **Request Context**: The current user message plus necessary recent messages from the same conversation used to produce a model-backed response.
- **Capability State**: Whether the current environment can produce model-backed responses, is unavailable, or is misconfigured.
- **Provider Configuration**: A locally managed Anthropic, OpenAI, Gemini, or OpenAI-compatible custom provider setup used to produce model-backed responses.
- **Tool Boundary Event**: A visible, non-executed indication that model output requested or implied a tool action that is outside the current feature scope.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: In a model-capable local validation environment, at least 95% of standard test prompts show first visible assistant output within 5 seconds.
- **SC-002**: 100% of completed model-backed responses can be traced by a tester from the submitted user message to visible progress states and the final assistant response.
- **SC-003**: In at least 95% of stop attempts during local validation, the response reaches a canceled terminal state within 2 seconds and no further assistant text is appended afterward.
- **SC-004**: 100% of attempts made without available model capability show an explicit unavailable or misconfigured state and do not display simulated output as if it were real.
- **SC-005**: 100% of tested model failure, rejection, timeout, rate-limit, empty-response, and interruption cases leave the conversation usable for a follow-up message and show a user-safe execution state.
- **SC-006**: 100% of tested tool-like model requests result in no external action and a visible indication that tool execution is outside the current capability boundary.
- **SC-007**: A tester can complete the same basic model-backed response flow with Anthropic, OpenAI, Gemini, and one locally configured OpenAI-compatible custom provider, with consistent user-visible states.

## Assumptions

- M5 follows the current docs-site roadmap, where the next milestone is LLM Gateway on top of completed thread/message and run/event foundations; older roadmap references that label M5 as Web Chat Timeline are treated as superseded by the current status page.
- Existing conversation, message, execution history, and frontend runtime surfaces remain available and are reused as user-facing product surfaces.
- Actual tool execution, worker/job queue, desktop runtime, multi-agent behavior, and broad platform automation remain out of scope for this feature.
- Anthropic, OpenAI, Gemini, and OpenAI-compatible custom provider support are product requirements for M5; provider setup is local configuration rather than product UI in this slice.
- Development validation may use a locally configured OpenAI-compatible custom provider at `https://apikey.tgjqr.com/v1` with model `gpt-5.5`.
- This feature prioritizes a runnable vertical slice for learning and validation over production-grade provider coverage or enterprise administration.
