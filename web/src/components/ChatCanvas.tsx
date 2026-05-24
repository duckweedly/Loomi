import type { AssistantDraft as AssistantDraftState, BackendCapabilityState, ChatCanvasState, Message, ProviderCapability, Run, StreamState, Thread } from '../domain'
import type { Locale } from '../i18n'
import { getDictionary } from '../i18n'
import { deriveBackendCapabilityStatus, getBackendCapabilityCopy, shouldShowProviderUnavailableWarning } from '../runtime/backendCapabilityStatus'
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
  providerCapabilities?: ProviderCapability[]
  onOpenProviderSettings?: () => void
  onSendMessage: (content: string) => void
  onStopRun: () => void
  onRetryRun?: () => void
  onRegenerateRun?: () => void
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
    stopped: { title: copy.stoppedTitle, detail: copy.stoppedDetail },
    recovering: { title: copy.recoveringTitle, detail: copy.recoveringDetail },
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
  if (status === 'recovering') return copy.recoveringDraft
  return copy.modelDrafting
}

function draftStatusLabel(status: AssistantDraftState['status'], locale: Locale) {
  const copy = getDictionary(locale).chatCanvas
  if (status === 'streaming' || status === 'drafting') return copy.generating
  if (status === 'completed') return copy.completedTitle
  if (status === 'failed') return copy.failedTitle
  if (status === 'stopped') return copy.stoppedTitle
  if (status === 'recovering') return copy.recoveringTitle
  return copy.waitingRunTitle
}

function AssistantDraft({ run, locale }: { run: Run | null; locale: Locale }) {
  const copy = getDictionary(locale).chatCanvas
  const draft = run?.assistantDraft
  if (!run || !draft || draft.status === 'empty') return null

  return (
    <article className={`message-row assistant draft ${draft.status}`}>
      <div className="message-avatar">L</div>
      <div className="message-bubble">
        <div className="message-meta">{copy.assistant} · {draftStatusLabel(draft.status, locale)}</div>
        <p className="message-markdown">{draft.content || draftFallback(draft.status, locale)}</p>
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

export function ChatCanvas({ sidebarCollapsed, thread, messages, run, loading, error, dataSourceMode, streamState, backendCapability = 'available', backendUnavailableAttempted = false, capabilitySignals, providerCapabilities = [], onOpenProviderSettings, onSendMessage, onStopRun, onRetryRun, onRegenerateRun, locale }: Props) {
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
  const composerDisabled = state === 'loading' || state === 'error' || state === 'no-thread' || state === 'backend-unavailable' || state === 'waiting-run' || state === 'running' || state === 'recovering'
  const composerPlaceholder = state === 'history' ? copy.messageLoomi : stateCopy[state].title
  const providerUnavailableBeforeSend = shouldShowProviderUnavailableWarning(dataSourceMode, providerCapabilities)
  const capabilityStatus = deriveBackendCapabilityStatus({
    dataSourceMode,
    runtimeSource: run?.context === 'model_gateway' ? 'model_gateway' : 'local_simulated',
    backendUnavailable: backendCapability === 'unavailable' || backendUnavailableAttempted || capabilitySignals?.backendUnavailable,
    modelSetupMissing: capabilitySignals?.modelSetupMissing,
    providerUnavailable: capabilitySignals?.providerUnavailable || providerUnavailableBeforeSend,
    activeRun: Boolean(run && (run.status === 'pending' || run.status === 'running' || run.status === 'retrying' || run.status === 'recovering')),
    streamDisconnected: Boolean(run && (run.status === 'pending' || run.status === 'running' || run.status === 'retrying' || run.status === 'recovering') && (capabilitySignals?.streamDisconnected || streamState === 'recoverable_error')),
    runRecovering: run?.status === 'recovering' || run?.assistantDraft?.status === 'recovering',
  })
  const capabilityCopy = getBackendCapabilityCopy(capabilityStatus, locale)
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
        <span>{copy.context}</span>
        <strong>{run?.context === 'local_simulated' ? copy.localSimulated : run?.context === 'model_gateway' ? copy.modelGateway : run?.context ?? '-'}</strong>
        <span className="context-line" />
        {sidebarCollapsed && <strong>{thread?.title ?? 'Untitled'}</strong>}
        <span>{thread?.mode ?? 'work'}</span>
        <span className={`capability-chip ${capabilityStatus}`}>{capabilityCopy.title}</span>
        <span className="capability-detail">{capabilityCopy.detail}</span>
        <span>{streamState}</span>
        {run?.status === 'running' && <button className="titlebar-button" onClick={onStopRun}>{copy.stop}</button>}
      </div>

      {error && <div className="api-error">{error}</div>}
      <ToolBoundaryNotice run={run} locale={locale} />
      {providerUnavailableBeforeSend && (
        <div className="provider-warning" role="status">
          <span>{copy.providerUnavailableWarning}</span>
          <button type="button" onClick={onOpenProviderSettings}>{copy.openProviderSettings}</button>
        </div>
      )}

      <div className="message-list">
        {state === 'history' ? (
          <>
            <MessageHistory messages={messages} locale={locale} />
            {shouldShowAssistantDraft && <AssistantDraft run={run} locale={locale} />}
          </>
        ) : (
          <>
            {shouldShowHistory && <MessageHistory messages={messages} locale={locale} />}
            {shouldShowAssistantDraft && <AssistantDraft run={run} locale={locale} />}
            {(state === 'no-thread' || state === 'empty-thread' || state === 'loading' || state === 'error' || state === 'backend-unavailable') && <StatePanel state={state} error={state === 'error' ? error : null} locale={locale} />}
          </>
        )}
      </div>

      <Composer
        disabled={composerDisabled}
        placeholder={composerPlaceholder}
        threadSelected={Boolean(thread)}
        run={run}
        messages={messages}
        attachLabel={copy.attach}
        stopLabel={copy.stop}
        retryLabel={copy.retry}
        regenerateLabel={copy.regenerate}
        onSubmit={onSendMessage}
        onStop={onStopRun}
        onRetry={onRetryRun}
        onRegenerate={onRegenerateRun}
      />
    </section>
  )
}
