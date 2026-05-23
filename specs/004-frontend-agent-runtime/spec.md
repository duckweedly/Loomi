# Feature Specification: Frontend Agent Runtime Skeleton

**Feature Branch**: `004-frontend-agent-runtime`

**Created**: 2026-05-23

**Status**: Draft

**Input**: User description: "M3.5 Frontend Agent Runtime Skeleton：在后端 M4/M5 未完成前，为 Loomi 前端设计 Agent 交互骨架，包括 Chat Canvas 状态机、mock run/event 成功/失败剧本、Chat/Run Timeline/Agent 状态徽章联动，以及 future real adapter 接入点 sendMessage/createRun/subscribeRunEvents/appendAssistantDelta/completeRun/failRun/stopRun。要求 mock 和 real 共用同一 UI 状态机，中文学习文档，避免继续只做 UI 外观抛光。"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - 看到明确的 Chat 工作区状态 (Priority: P1)

作为 Loomi 的使用者，我需要 Chat 主区在没有真实后端执行能力时也明确展示当前工作状态，而不是出现无法解释的大空白。无论是没选线程、新线程还没有消息、正在加载、加载失败、后端能力未接入，还是已有历史消息，我都能判断下一步应该做什么。

**Why this priority**: 这是停止“只凭感觉抛光外观”的第一步。只要 Chat Canvas 有明确状态，后续 mock run、真实 run 和错误处理才能挂到同一套产品语义上。

**Independent Test**: 可以只实现状态骨架并用 mock 数据切换每个状态；用户不发送消息也能看到主区含义从“空白”变成明确的工作区状态。

**Acceptance Scenarios**:

1. **Given** 没有可选 thread，**When** 用户打开 Chat 模式，**Then** 主区显示“未选择会话”状态，并提供创建新对话的明确入口。
2. **Given** 用户选择了新建但还没有消息的 thread，**When** Chat Canvas 渲染，**Then** 主区显示新线程空状态，说明可以输入第一条消息。
3. **Given** thread 或消息正在加载，**When** 加载尚未完成，**Then** 主区显示加载状态，而不是空白历史消息区。
4. **Given** 加载失败，**When** 错误返回，**Then** 主区显示失败状态、错误摘要和重试入口。
5. **Given** 后端 run/event 能力未接入，**When** 用户处于真实 API 模式但 run 功能不可用，**Then** 主区显示“后端能力未接入”的清晰状态，不伪装成真实执行。

---

### User Story 2 - 用 mock 剧本体验一次完整 Agent 执行 (Priority: P1)

作为产品设计和开发者，我需要在后端 M4/M5 未完成时，通过 mock 剧本体验一次完整 Agent 执行：发送消息后立即看到用户消息、run 创建、上下文加载、思考、草拟、助手回复和完成结果。失败剧本也必须能展示失败路径。

**Why this priority**: 这让前端可以提前验证 Agent 产品体验，而不是等待后端完成后才发现 Chat、Timeline 和状态徽章无法联动。

**Independent Test**: 可以在 mock 模式下发送一条消息并选择成功或失败剧本，观察消息区和执行区是否按剧本推进。

**Acceptance Scenarios**:

1. **Given** 用户在 mock 模式下输入消息，**When** 提交消息，**Then** 用户消息立即追加到 Chat Canvas，并创建新的 run 记录。
2. **Given** 成功剧本正在执行，**When** 剧本推进，**Then** Timeline 依次显示 run created、context loading、assistant thinking、assistant drafting、assistant message、run completed。
3. **Given** 失败剧本正在执行，**When** 剧本进入失败步骤，**Then** Timeline 显示失败事件，Chat Canvas 显示失败状态，并且不会追加伪成功的 assistant 回复。
4. **Given** mock 剧本执行中，**When** 用户停止 run，**Then** Timeline 和 Chat Canvas 都进入停止或失败类状态，并停止后续 mock 事件。

---

### User Story 3 - Chat、Timeline 和 Agent 状态徽章联动 (Priority: P2)

作为 Loomi 的使用者，我需要消息区、执行时间线和 Agent 状态徽章看起来像同一次执行，而不是三个互不相关的装饰区域。发送消息后，Chat 区、Run Timeline 和刺猬状态动效要一起响应同一个 run 状态。

**Why this priority**: 这是前端从静态壳变成 Agent 产品交互骨架的关键闭环，也为 M4 的真实事件流打基础。

**Independent Test**: 可以只用 mock run/event 数据驱动三块 UI，验证同一条用户消息导致三块区域同步变化。

**Acceptance Scenarios**:

1. **Given** 用户发送消息，**When** run 刚创建，**Then** Chat Canvas 显示等待执行状态，Timeline 显示第一个执行事件，Agent 徽章进入非 idle 状态。
2. **Given** run 正在执行上下文加载或思考事件，**When** Timeline 更新，**Then** Agent 徽章显示对应运动状态，Chat Canvas 显示等待或执行中状态。
3. **Given** assistant 回复完成，**When** run completed，**Then** Chat Canvas 追加最终 assistant 回复，Timeline 标记完成，Agent 徽章进入完成状态。
4. **Given** run 失败，**When** 失败事件出现，**Then** Chat Canvas、Timeline 和 Agent 徽章都显示失败语义。

---

### User Story 4 - 为真实后端接入预留同一套状态机 (Priority: P2)

作为开发者，我需要 mock 和未来真实后端共用同一套 UI 状态机。后端 M4/M5 完成后，替换数据来源即可接入真实 run/event/SSE，不需要重写 Chat Canvas、Timeline 或 Agent 状态徽章。

**Why this priority**: 如果 mock 和 real 各自走一套逻辑，前端现在做的体验验证会在后端完成后被推翻。统一状态机能保护当前投入。

**Independent Test**: 可以通过 mock adapter 和 real adapter 的契约测试验证二者输出相同的前端运行状态和事件语义。

**Acceptance Scenarios**:

1. **Given** mock adapter 提供成功 run 剧本，**When** UI 消费其状态，**Then** UI 使用同一套状态机渲染 Chat、Timeline 和 Agent 徽章。
2. **Given** real adapter 暂时不支持 run/event 能力，**When** 用户尝试触发执行，**Then** UI 进入“后端能力未接入”状态，而不是走独立 mock-only 逻辑。
3. **Given** 未来 real adapter 提供 run/event 流，**When** 它产生与 mock adapter 同语义的事件，**Then** UI 不需要改变用户可见行为。

---

### Edge Cases

- 用户在没有选中 thread 时直接提交消息：系统必须阻止提交并显示“未选择会话”状态。
- 用户在新线程空状态下提交第一条消息：系统必须先显示用户消息，再进入等待 run 或执行中状态。
- 用户快速切换 thread 时，旧 thread 的 mock run 后续事件不得污染当前 thread。
- 用户在 run 执行中再次发送消息：系统必须给出明确状态，避免两个 run 的事件混在同一个 timeline 中。
- mock 成功剧本和失败剧本都必须可重复触发，且每次执行都能产生独立 run 记录。
- 后端真实 API 已配置但 run/event 能力不可用时，系统必须显示能力未接入，而不是静默 fallback 到 mock。
- assistant delta 为空或中断时，系统必须保留用户消息和失败/停止状态。
- 停止 run 后，后续事件不得继续推进到 completed。

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST define a visible Chat Canvas state for no selected thread, empty new thread, loading, load failure, message history, waiting for run, run running, run completed, run failed, and backend capability unavailable.
- **FR-002**: System MUST render each Chat Canvas state with concise Chinese user-facing text and an obvious next action when an action is available.
- **FR-003**: System MUST preserve existing thread/message history rendering when a selected thread already has messages.
- **FR-004**: System MUST provide a mock successful run script that starts from a submitted user message and ends with a final assistant message and completed run state.
- **FR-005**: System MUST provide a mock failed run script that starts from a submitted user message and ends with a failed run state without adding a fake successful assistant reply.
- **FR-006**: System MUST update Chat Canvas, Run Timeline, and Agent state motion from the same run state and event sequence.
- **FR-007**: System MUST show user messages immediately after submit before the mock or real run finishes.
- **FR-008**: System MUST show run lifecycle events in order: run creation, context loading, assistant thinking, assistant drafting, assistant message completion, and run completion for the successful script.
- **FR-009**: System MUST expose a front-end execution boundary with these capabilities: send message, create run, subscribe to run events, append assistant delta, complete run, fail run, and stop run.
- **FR-010**: System MUST make mock and future real execution adapters produce the same UI-visible state transitions.
- **FR-011**: System MUST represent backend run/event capability absence as a first-class state instead of silently falling back to mock behavior when real API mode is configured.
- **FR-012**: System MUST prevent stale events from a previously selected thread from changing the visible state of the currently selected thread.
- **FR-013**: System MUST support stopping an in-progress mock run and reflect the stopped state consistently in Chat Canvas, Timeline, and Agent state motion.
- **FR-014**: System MUST keep product UI microcopy sparse while keeping learning documentation in Chinese with enough detail to explain the state model and adapter boundary.
- **FR-015**: System MUST keep Chat and Work recent thread lists mode-specific while applying the same runtime-state rules to the selected Chat thread.

### Key Entities *(include if feature involves data)*

- **Chat Canvas State**: The user-visible state of the main chat work area. It captures whether the UI has a selected thread, messages, loading/error conditions, run progress, completed output, failure, or missing backend capability.
- **Runtime Script**: A deterministic mock execution scenario that describes how a submitted user message becomes ordered run events and, for success, a final assistant response.
- **Run Event**: A user-visible execution milestone such as run created, context loading, thinking, drafting, assistant message completed, completed, failed, or stopped.
- **Assistant Draft**: The in-progress assistant response content shown while a run is drafting. It may be empty, partial, complete, stopped, or failed.
- **Execution Adapter**: A source of runtime actions and events. The mock adapter and future real adapter must expose the same user-visible semantics.
- **Backend Capability State**: Whether the currently selected data source can provide run/event behavior. Unsupported capability is a visible state, not an invisible fallback.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A user can identify the Chat Canvas state in 100% of the listed states without interpreting an empty area as missing UI.
- **SC-002**: In mock mode, a submitted message produces the first visible user message update in under 300 milliseconds on a local development machine.
- **SC-003**: In mock success mode, one submitted message visibly progresses through at least 6 ordered execution milestones and ends with one assistant reply.
- **SC-004**: In mock failure mode, one submitted message visibly progresses to a failed state and does not append a successful assistant reply.
- **SC-005**: Chat Canvas, Timeline, and Agent state motion reflect the same selected run state for 100% of mock success, failure, and stopped script checks.
- **SC-006**: Switching threads during a mock run results in 0 stale event updates being applied to the newly selected thread.
- **SC-007**: When real API mode lacks run/event support, users see a backend capability unavailable state within 1 second of attempting an execution.
- **SC-008**: Future real adapter planning can point to the same execution boundary used by mock without defining a second UI state model.

## Assumptions

- This is an M3.5 frontend milestone that sits between M3 thread/message persistence and M4 run/event/SSE.
- The feature does not require real backend run/event/SSE, LLM gateway, worker queue, tool execution, or desktop runtime changes.
- Mock scripts are deterministic by default so frontend behavior is reproducible in tests and screenshots.
- User-facing UI state labels and learning documentation should be Chinese; code identifiers and public technical contracts remain English.
- The first implementation focuses on Chat mode. Work mode keeps separate recent threads and can adopt the same runtime model later when work-specific execution flows are designed.
- Real API mode must stay honest: if backend capability is missing, it shows a capability state instead of pretending to execute with mock data.
