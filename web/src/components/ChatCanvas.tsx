import type { ChatCanvasState, Message, Run, Thread } from '../domain'
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
  backendCapability?: 'available' | 'unavailable'
  backendUnavailableAttempted?: boolean
  onSendMessage: (content: string) => void
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

function StatePanel({ state, error }: { state: Exclude<ChatCanvasState, 'history'>; error?: string | null }) {
  const copy = stateCopy[state]
  return (
    <div className={`empty-state chat-state ${state}`}>
      <strong>{copy.title}</strong>
      <span>{error ?? copy.detail}</span>
    </div>
  )
}

export function ChatCanvas({ sidebarCollapsed, thread, messages, run, loading, error, dataSourceMode, backendCapability = 'available', backendUnavailableAttempted = false, onSendMessage }: Props) {
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
        <strong>{run?.context ?? '-'}</strong>
        <span className="context-line" />
        {sidebarCollapsed && <strong>{thread?.title ?? 'Untitled'}</strong>}
        <span>{thread?.mode ?? 'work'}</span>
        <span>{dataSourceMode === 'real_api' ? 'Real API' : 'Mock'}</span>
      </div>

      <div className="message-list">
        {state === 'history' ? (
          <MessageHistory messages={messages} />
        ) : (
          <>
            {(state === 'waiting-run' || state === 'running' || state === 'completed' || state === 'failed') && <MessageHistory messages={messages} />}
            <StatePanel state={state} error={state === 'error' ? error : null} />
          </>
        )}
      </div>

      <Composer disabled={composerDisabled} placeholder={composerPlaceholder} onSubmit={onSendMessage} />
    </section>
  )
}
