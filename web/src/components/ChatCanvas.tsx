import type { Message, Run, Thread } from '../domain'
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
  onSendMessage: (content: string) => void
}

export function ChatCanvas({ sidebarCollapsed, thread, messages, run, loading, error, dataSourceMode, onSendMessage }: Props) {
  return (
    <section className="chat-shell glass-panel">
      <div className="context-bar">
        <span>Context</span>
        <strong>{run?.context ?? '-'}</strong>
        <span className="context-line" />
        {sidebarCollapsed && <strong>{thread?.title ?? 'Untitled'}</strong>}
        <span>{thread?.mode ?? 'work'}</span>
        <span>{dataSourceMode === 'real_api' ? 'Real API' : 'Mock'}</span>
      </div>

      {error && <div className="api-error">{error}</div>}

      <div className="message-list">
        {loading ? (
          <div className="empty-state">Loading</div>
        ) : messages.map((message) => (
          <article key={message.id} className={`message-row ${message.role}`}>
            <div className="message-avatar">{message.role === 'assistant' ? 'L' : 'U'}</div>
            <div className="message-bubble">
              <div className="message-meta">{message.role === 'assistant' ? 'Loomi' : 'You'} · {message.createdAt}</div>
              <p className="message-markdown">{message.content}</p>
              {message.toolCalls?.map((toolCall) => <ToolCallCard key={toolCall.id} toolCall={toolCall} />)}
            </div>
          </article>
        ))}
      </div>

      <Composer disabled={loading} onSubmit={onSendMessage} />
    </section>
  )
}
