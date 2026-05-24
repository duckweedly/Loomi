import type { AssistantDraft as AssistantDraftState, BackendCapabilityState, ChatCanvasState, Message, Run, StreamState, Thread } from '../domain'
import type { Locale } from '../i18n'
import { getDictionary } from '../i18n'
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
  locale: Locale
}

function createStateCopy(locale: Locale): Record<Exclude<ChatCanvasState, 'history'>, { title: string; detail: string }> {
  const copy = getDictionary(locale).chatCanvas
  return {
    'no-thread': { title: copy.noThreadTitle, detail: copy.noThreadDetail },
    'empty-thread': { title: copy.emptyThreadTitle, detail: copy.emptyThreadDetail },
    loading: { title: copy.loadingTitle, detail: copy.loadingDetail },
    error: { title: copy.errorTitle, detail: copy.errorDetail },
    'waiting-run': { title: copy.waitingRunTitle, detail: copy.waitingRunDetail },
    running: { title: copy.runningTitle, detail: copy.runningDetail },
    completed: { title: copy.completedTitle, detail: copy.completedDetail },
    failed: { title: copy.failedTitle, detail: copy.failedDetail },
    'backend-unavailable': { title: copy.backendUnavailableTitle, detail: copy.backendUnavailableDetail },
  }
}

function MessageHistory({ messages, locale }: { messages: Message[]; locale: Locale }) {
  const copy = getDictionary(locale).chatCanvas
  return messages.map((message) => (
    <article key={message.id} className={`message-row ${message.role}`}>
      <div className="message-avatar">{message.role === 'assistant' ? 'L' : 'U'}</div>
      <div className="message-bubble">
        <div className="message-meta">{message.role === 'assistant' ? copy.assistant : copy.user} · {message.createdAt}</div>
        <p className="message-markdown">{message.content}</p>
        {message.toolCalls?.map((toolCall) => <ToolCallCard key={toolCall.id} toolCall={toolCall} />)}
      </div>
    </article>
  ))
}

function draftFallback(status: AssistantDraftState['status'], locale: Locale) {
  const copy = getDictionary(locale).chatCanvas
  if (status === 'failed') return copy.failedDetail
  if (status === 'stopped') return copy.stoppedDraft
  return copy.modelDrafting
}

function AssistantDraft({ run, locale }: { run: Run | null; locale: Locale }) {
  const copy = getDictionary(locale).chatCanvas
  if (!run?.assistantDraft || run.assistantDraft.status === 'empty') return null
  return (
    <article className={`message-row assistant ${run.assistantDraft.status}`}>
      <div className="message-avatar">L</div>
      <div className="message-bubble">
        <div className="message-meta">{copy.assistant} · {copy.modelGateway}</div>
        <p className="message-markdown">{run.assistantDraft.content || draftFallback(run.assistantDraft.status, locale)}</p>
      </div>
    </article>
  )
}

function ToolBoundaryNotice({ run, locale }: { run: Run | null; locale: Locale }) {
  if (!run?.events.some((event) => event.type === 'progress.tool_call_blocked')) return null
  return <div className="api-error">{getDictionary(locale).chatCanvas.toolBoundaryNotice}</div>
}

function StatePanel({ state, error, locale }: { state: Exclude<ChatCanvasState, 'history'>; error?: string | null; locale: Locale }) {
  const copy = createStateCopy(locale)[state]
  return (
    <div className={`empty-state chat-state ${state}`}>
      <strong>{copy.title}</strong>
      <span>{error ?? copy.detail}</span>
    </div>
  )
}

export function ChatCanvas({ sidebarCollapsed, thread, messages, run, loading, error, dataSourceMode, streamState, backendCapability = 'available', backendUnavailableAttempted = false, onSendMessage, onStopRun, locale }: Props) {
  const state = deriveChatCanvasState({
    loading,
    error,
    backendCapability,
    backendUnavailableAttempted,
    selectedThreadId: thread?.id ?? null,
    messageCount: messages.length,
    run,
  })
  const copy = getDictionary(locale).chatCanvas
  const stateCopy = createStateCopy(locale)
  const composerDisabled = state === 'loading' || state === 'error' || state === 'no-thread' || state === 'backend-unavailable' || state === 'waiting-run' || state === 'running'
  const composerPlaceholder = state === 'history' ? copy.messageLoomi : stateCopy[state].title

  return (
    <section className="chat-shell glass-panel" data-chat-state={state}>
      <div className="context-bar">
        <span>{copy.context}</span>
        <strong>{run?.context === 'local_simulated' ? copy.localSimulated : run?.context === 'model_gateway' ? copy.modelGateway : run?.context ?? '-'}</strong>
        <span className="context-line" />
        {sidebarCollapsed && <strong>{thread?.title ?? 'Untitled'}</strong>}
        <span>{thread?.mode ?? 'work'}</span>
        <span>{dataSourceMode === 'real_api' ? 'Real API' : 'Mock'}</span>
        <span>{streamState}</span>
        {run?.status === 'running' && <button className="titlebar-button" onClick={onStopRun}>{copy.stop}</button>}
      </div>

      {error && <div className="api-error">{error}</div>}
      <ToolBoundaryNotice run={run} locale={locale} />

      <div className="message-list">
        {state === 'history' ? (
          <>
            <MessageHistory messages={messages} locale={locale} />
            <AssistantDraft run={run} locale={locale} />
          </>
        ) : (
          <>
            {(state === 'waiting-run' || state === 'running' || state === 'completed' || state === 'failed') && <MessageHistory messages={messages} locale={locale} />}
            <AssistantDraft run={run} locale={locale} />
            <StatePanel state={state} error={state === 'error' ? error : null} locale={locale} />
          </>
        )}
      </div>

      <Composer disabled={composerDisabled} placeholder={composerPlaceholder} attachLabel={copy.attach} onSubmit={onSendMessage} />
    </section>
  )
}
