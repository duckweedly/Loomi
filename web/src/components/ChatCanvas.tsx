import type { BackendCapabilityState, ChatCanvasState, Message, Run, StreamState, Thread } from '../domain'
import { deriveBackendCapabilityStatus, getBackendCapabilityCopy } from '../runtime/backendCapabilityStatus'
import { deriveChatCanvasState } from '../runtime/chatCanvasState'
import { Composer } from './Composer'
import { ToolCallCard } from './ToolCallCard'

type Props = {
  sidebarCollapsed: boolean
  thread: Thread | null
  messages: Message[]
  run: Run | null
  loading: boolean
  error?: string | null
  dataSourceMode: 'mock' | 'real_api'
  streamState: StreamState
  backendCapability?: BackendCapabilityState
  backendUnavailableAttempted?: boolean
  capabilitySignals?: {
    backendUnavailable?: boolean
    modelSetupMissing?: boolean
    providerUnavailable?: boolean
    streamDisconnected?: boolean
  }
  onSendMessage: (content: string) => void
  onStopRun: () => void
  onRetryRun?: () => void
  onRegenerateRun?: () => void
}

const stateCopy: Record<Exclude<ChatCanvasState, 'history'>, { title: string; detail: string }> = {
  'no-thread': { title: '未选择会话', detail: '创建新对话' },
  'empty-thread': { title: '新对话', detail: '输入第一条消息' },
  loading: { title: '加载中', detail: '同步会话' },
  error: { title: '加载失败', detail: '重试' },
  'waiting-run': { title: '等待执行', detail: '消息已发送' },
  running: { title: '执行中', detail: '查看右侧时间线' },
  completed: { title: '已完成', detail: '回复已生成' },
  failed: { title: '执行失败', detail: '未生成成功回复' },
  stopped: { title: '已停止', detail: '保留已生成内容' },
  recovering: { title: '恢复中', detail: '正在恢复运行状态' },
  'backend-unavailable': { title: '后端能力未接入', detail: '等待 M4/M5 run/event' },
}

function MessageHistory({ messages }: { messages: Message[] }) {
  return messages.map((message) => (
    <article key={message.id} className={`message-row ${message.role}`}>
      <div className="message-avatar">{message.role === 'assistant' ? 'L' : 'U'}</div>
      <div className="message-bubble">
        <div className="message-meta">{message.role === 'assistant' ? 'Loomi' : 'You'} · {message.createdAt}</div>
        <p className="message-markdown">{message.content}</p>
        {message.toolCalls?.map((toolCall) => <ToolCallCard key={toolCall.id} toolCall={toolCall} />)}
      </div>
    </article>
  ))
}

function AssistantDraftBubble({ run }: { run: Run }) {
  const draft = run.assistantDraft
  if (!draft) return null

  const statusFallback = draft.status === 'pending'
    ? '等待回复…'
    : draft.status === 'failed'
      ? '未生成成功回复'
      : draft.status === 'stopped'
        ? '已停止生成，保留已生成内容'
        : draft.status === 'recovering'
          ? '恢复中…'
          : draft.status === 'empty'
            ? '模型正在生成回复'
            : ''
  const content = draft.content || statusFallback
  if (!content) return null

  const statusLabel = draft.status === 'streaming' || draft.status === 'drafting'
    ? '生成中'
    : draft.status === 'completed'
      ? '已完成'
      : draft.status === 'failed'
        ? '执行失败'
        : draft.status === 'stopped'
          ? '已停止'
          : draft.status === 'recovering'
            ? '恢复中'
            : '等待执行'

  return (
    <article className={`message-row assistant draft ${draft.status}`}>
      <div className="message-avatar">L</div>
      <div className="message-bubble">
        <div className="message-meta">Loomi · {statusLabel}</div>
        <p className="message-markdown">{content}</p>
      </div>
    </article>
  )
}

function ToolBoundaryNotice({ run }: { run: Run | null }) {
  if (!run?.events.some((event) => event.type === 'progress.tool_call_blocked')) return null
  return <div className="api-error">工具调用未执行：M5 只记录边界事件，不执行外部动作。</div>
}

function StatePanel({ state, error }: { state: Exclude<ChatCanvasState, 'history'>; error?: string | null }) {
  const copy = stateCopy[state]
  return (
    <div className={`empty-state chat-state ${state}`}>
      <strong>{copy.title}</strong>
      <span>{error ?? copy.detail}</span>
    </div>
  )
}

export function ChatCanvas({ sidebarCollapsed, thread, messages, run, loading, error, dataSourceMode, streamState, backendCapability = 'available', backendUnavailableAttempted = false, capabilitySignals, onSendMessage, onStopRun, onRetryRun, onRegenerateRun }: Props) {
  const state = deriveChatCanvasState({
    loading,
    error,
    backendCapability,
    backendUnavailableAttempted,
    selectedThreadId: thread?.id ?? null,
    messageCount: messages.length,
    run,
  })
  const composerDisabled = state === 'loading' || state === 'error' || state === 'no-thread' || state === 'backend-unavailable' || state === 'waiting-run' || state === 'running' || state === 'recovering'
  const composerPlaceholder = state === 'history' ? 'Message Loomi' : stateCopy[state].title
  const capabilityStatus = deriveBackendCapabilityStatus({
    dataSourceMode,
    runtimeSource: run?.context === 'model_gateway' ? 'model_gateway' : 'local_simulated',
    backendUnavailable: backendCapability === 'unavailable' || backendUnavailableAttempted || capabilitySignals?.backendUnavailable,
    modelSetupMissing: capabilitySignals?.modelSetupMissing,
    providerUnavailable: capabilitySignals?.providerUnavailable,
    activeRun: Boolean(run && (run.status === 'pending' || run.status === 'running' || run.status === 'retrying' || run.status === 'recovering')),
    streamDisconnected: Boolean(run && (run.status === 'pending' || run.status === 'running' || run.status === 'retrying' || run.status === 'recovering') && (capabilitySignals?.streamDisconnected || streamState === 'recoverable_error')),
    runRecovering: run?.status === 'recovering' || run?.assistantDraft?.status === 'recovering',
  })
  const capabilityCopy = getBackendCapabilityCopy(capabilityStatus)
  const visibleMessages = messages
  const hasPersistedCompletedDraftMessage = run?.assistantDraft?.status === 'completed' && messages.some((message) => (
    message.role === 'assistant' && (
      message.id === run.assistantDraft?.messageId ||
      message.runId === run.id ||
      (message.threadId === run.threadId && message.content === run.assistantDraft?.content)
    )
  ))
  const shouldShowAssistantDraft = Boolean(run && !hasPersistedCompletedDraftMessage)
  const shouldShowHistory = state === 'history' || state === 'waiting-run' || state === 'running' || state === 'completed' || state === 'failed' || state === 'stopped' || state === 'recovering'

  return (
    <section className="chat-shell glass-panel" data-chat-state={state}>
      <div className="context-bar">
        <span>Context</span>
        <strong>{run?.context === 'local_simulated' ? 'Local simulated' : run?.context === 'model_gateway' ? '模型网关' : run?.context ?? '-'}</strong>
        <span className="context-line" />
        {sidebarCollapsed && <strong>{thread?.title ?? 'Untitled'}</strong>}
        <span>{thread?.mode ?? 'work'}</span>
        <span className={`capability-chip ${capabilityStatus}`}>{capabilityCopy.title}</span>
        <span className="capability-detail">{capabilityCopy.detail}</span>
        <span>{streamState}</span>
        {run?.status === 'running' && <button className="titlebar-button" onClick={onStopRun}>Stop</button>}
      </div>

      {error && <div className="api-error">{error}</div>}
      <ToolBoundaryNotice run={run} />

      <div className="message-list">
        {state === 'history' ? (
          <MessageHistory messages={visibleMessages} />
        ) : (
          <>
            {shouldShowHistory && <MessageHistory messages={visibleMessages} />}
            {shouldShowAssistantDraft && run && <AssistantDraftBubble run={run} />}
            {(state === 'no-thread' || state === 'empty-thread' || state === 'loading' || state === 'error' || state === 'backend-unavailable') && <StatePanel state={state} error={state === 'error' ? error : null} />}
          </>
        )}
      </div>

      <Composer disabled={composerDisabled} placeholder={composerPlaceholder} threadSelected={Boolean(thread)} run={run} messages={messages} onSubmit={onSendMessage} onStop={onStopRun} onRetry={onRetryRun} onRegenerate={onRegenerateRun} />
    </section>
  )
}
