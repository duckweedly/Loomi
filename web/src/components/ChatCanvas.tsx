import type { AssistantDraft as AssistantDraftState, BackendCapabilityState, ChatCanvasState, Message, Run, StreamState, Thread } from '../domain'
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
  onSendMessage: (content: string) => void
  onStopRun: () => void
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

function draftFallback(status: AssistantDraftState['status']) {
  if (status === 'failed') return '未生成成功回复'
  if (status === 'stopped') return '已停止生成'
  return '模型正在生成回复'
}

function AssistantDraft({ run }: { run: Run | null }) {
  if (!run?.assistantDraft || run.assistantDraft.status === 'empty') return null
  return (
    <article className={`message-row assistant ${run.assistantDraft.status}`}>
      <div className="message-avatar">L</div>
      <div className="message-bubble">
        <div className="message-meta">Loomi · 模型网关</div>
        <p className="message-markdown">{run.assistantDraft.content || draftFallback(run.assistantDraft.status)}</p>
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

export function ChatCanvas({ sidebarCollapsed, thread, messages, run, loading, error, dataSourceMode, streamState, backendCapability = 'available', backendUnavailableAttempted = false, onSendMessage, onStopRun }: Props) {
  const state = deriveChatCanvasState({
    loading,
    error,
    backendCapability,
    backendUnavailableAttempted,
    selectedThreadId: thread?.id ?? null,
    messageCount: messages.length,
    run,
  })
  const composerDisabled = state === 'loading' || state === 'error' || state === 'no-thread' || state === 'backend-unavailable' || state === 'waiting-run' || state === 'running'
  const composerPlaceholder = state === 'history' ? 'Message Loomi' : stateCopy[state].title

  return (
    <section className="chat-shell glass-panel" data-chat-state={state}>
      <div className="context-bar">
        <span>Context</span>
        <strong>{run?.context === 'local_simulated' ? 'Local simulated' : run?.context === 'model_gateway' ? '模型网关' : run?.context ?? '-'}</strong>
        <span className="context-line" />
        {sidebarCollapsed && <strong>{thread?.title ?? 'Untitled'}</strong>}
        <span>{thread?.mode ?? 'work'}</span>
        <span>{dataSourceMode === 'real_api' ? 'Real API' : 'Mock'}</span>
        <span>{streamState}</span>
        {run?.status === 'running' && <button className="titlebar-button" onClick={onStopRun}>Stop</button>}
      </div>

      {error && <div className="api-error">{error}</div>}
      <ToolBoundaryNotice run={run} />

      <div className="message-list">
        {state === 'history' ? (
          <>
            <MessageHistory messages={messages} />
            <AssistantDraft run={run} />
          </>
        ) : (
          <>
            {(state === 'waiting-run' || state === 'running' || state === 'completed' || state === 'failed') && <MessageHistory messages={messages} />}
            <AssistantDraft run={run} />
            <StatePanel state={state} error={state === 'error' ? error : null} />
          </>
        )}
      </div>

      <Composer disabled={composerDisabled} placeholder={composerPlaceholder} onSubmit={onSendMessage} />
    </section>
  )
}
