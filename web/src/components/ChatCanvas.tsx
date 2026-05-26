import { useState } from 'react'
import { Copy, ChevronDown, ChevronRight, RefreshCcw, RotateCcw } from 'lucide-react'
import type { AssistantDraft as AssistantDraftState, BackendCapabilityState, ChatCanvasState, Message, Persona, ProviderCapability, Run, StreamState, Thread, WorkspaceRootConfig } from '../domain'
import type { Locale } from '../i18n'
import { getDictionary } from '../i18n'
import { getProviderUnavailableWarning, shouldShowProviderUnavailableWarning } from '../runtime/backendCapabilityStatus'
import { deriveChatCanvasState } from '../runtime/chatCanvasState'
import { humanToolName } from '../runtime/toolPreview'
import type { ProviderSaveResult } from '../state'
import { deriveWorkPlanProjection } from '../workModeProjection'
import { Composer } from './Composer'
import type { ComposerAttachment, ComposerModelOption } from './Composer'
import { ToolCallCard } from './ToolCallCard'
import { WorkPlanView } from './WorkPlanView'

const activeRunStatuses = new Set<Run['status']>(['pending', 'queued', 'running', 'recovering', 'blocked_on_tool_approval', 'stopping', 'retrying'])

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
  workspaceRootConfig?: WorkspaceRootConfig | null
  workspaceRootSaveResult?: ProviderSaveResult
  personas?: Persona[]
  selectedPersonaId?: string
  onSelectPersona?: (personaId: string) => void
  onOpenProviderSettings?: () => void
  onChooseWorkspaceFolder?: () => void
  onSendMessage: (content: string, options?: { providerId?: string; model?: string; attachments?: ComposerAttachment[] }) => void
  onStopRun: () => void
  onRetryRun?: () => void
  onRegenerateRun?: () => void
  onApproveToolCall?: (toolCall: NonNullable<Run['toolCalls']>[number]) => Promise<void> | void
  onDenyToolCall?: (toolCall: NonNullable<Run['toolCalls']>[number]) => Promise<void> | void
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
    stopping: { title: copy.stoppingTitle, detail: copy.stoppingDetail },
    'backend-unavailable': { title: copy.backendUnavailableTitle, detail: copy.backendUnavailableDetail },
  }
}

function safeHref(href: string) {
  return /^(https?:|mailto:)/i.test(href) ? href : '#'
}

function renderInlineMarkdown(text: string, blockKey: string) {
  const parts = text.split(/(`[^`]+`|\*\*[^*]+\*\*|\[[^\]]+\]\([^)]+\))/g)
  return parts.filter(Boolean).map((part, index) => {
    const key = `${blockKey}-${index}`
    if (part.startsWith('`') && part.endsWith('`')) return <code key={key}>{part.slice(1, -1)}</code>
    if (part.startsWith('**') && part.endsWith('**')) return <strong key={key}>{part.slice(2, -2)}</strong>
    const link = part.match(/^\[([^\]]+)\]\(([^)]+)\)$/)
    if (link) return <a key={key} href={safeHref(link[2])} rel="noreferrer" target="_blank">{link[1]}</a>
    return part
  })
}

function isMarkdownTableSeparator(line: string) {
  const cells = line.trim().split('|').map((cell) => cell.trim()).filter(Boolean)
  return cells.length > 0 && cells.every((cell) => /^:?-{3,}:?$/.test(cell))
}

function parseMarkdownTableRow(line: string) {
  const trimmed = line.trim().replace(/^\|/, '').replace(/\|$/, '')
  return trimmed.split('|').map((cell) => cell.trim())
}

function isMarkdownTableStart(lines: string[], index: number) {
  return lines[index]?.includes('|') && isMarkdownTableSeparator(lines[index + 1] ?? '')
}

function MarkdownMessage({ content }: { content: string }) {
  const lines = content.replace(/\r\n/g, '\n').split('\n')
  const blocks = []
  for (let index = 0; index < lines.length;) {
    const line = lines[index]
    if (line.trim() === '') {
      index += 1
      continue
    }
    if (line.trim().startsWith('```')) {
      const codeLines = []
      index += 1
      while (index < lines.length && !lines[index].trim().startsWith('```')) {
        codeLines.push(lines[index])
        index += 1
      }
      index += 1
      blocks.push(<pre key={`code-${index}`}><code>{codeLines.join('\n')}</code></pre>)
      continue
    }
    if (isMarkdownTableStart(lines, index)) {
      const headers = parseMarkdownTableRow(lines[index])
      index += 2
      const rows = []
      while (index < lines.length && lines[index].trim() !== '' && lines[index].includes('|') && !lines[index].trim().startsWith('```')) {
        rows.push(parseMarkdownTableRow(lines[index]))
        index += 1
      }
      blocks.push(
        <div className="message-table-wrap" key={`table-${index}`}>
          <table>
            <thead>
              <tr>{headers.map((header, cellIndex) => <th key={`th-${cellIndex}`}>{renderInlineMarkdown(header, `th-${index}-${cellIndex}`)}</th>)}</tr>
            </thead>
            <tbody>
              {rows.map((row, rowIndex) => (
                <tr key={`tr-${index}-${rowIndex}`}>
                  {headers.map((_, cellIndex) => <td key={`td-${index}-${rowIndex}-${cellIndex}`}>{renderInlineMarkdown(row[cellIndex] ?? '', `td-${index}-${rowIndex}-${cellIndex}`)}</td>)}
                </tr>
              ))}
            </tbody>
          </table>
        </div>,
      )
      continue
    }
    const heading = line.match(/^(#{1,3})\s+(.+)$/)
    if (heading) {
      const children = renderInlineMarkdown(heading[2], `heading-${index}`)
      blocks.push(heading[1].length === 1 ? <h1 key={`heading-${index}`}>{children}</h1> : heading[1].length === 2 ? <h2 key={`heading-${index}`}>{children}</h2> : <h3 key={`heading-${index}`}>{children}</h3>)
      index += 1
      continue
    }
    if (/^\s*[-*]\s+/.test(line)) {
      const items = []
      while (index < lines.length && /^\s*[-*]\s+/.test(lines[index])) {
        items.push(<li key={`li-${index}`}>{renderInlineMarkdown(lines[index].replace(/^\s*[-*]\s+/, ''), `li-${index}`)}</li>)
        index += 1
      }
      blocks.push(<ul key={`ul-${index}`}>{items}</ul>)
      continue
    }
    const paragraph = [line.trim()]
    index += 1
    while (index < lines.length && lines[index].trim() !== '' && !lines[index].trim().startsWith('```') && !/^(#{1,3})\s+/.test(lines[index]) && !/^\s*[-*]\s+/.test(lines[index]) && !isMarkdownTableStart(lines, index)) {
      paragraph.push(lines[index].trim())
      index += 1
    }
    blocks.push(<p key={`p-${index}`}>{renderInlineMarkdown(paragraph.join(' '), `p-${index}`)}</p>)
  }
  return <div className="message-markdown">{blocks}</div>
}

function displayMessageTime(value: string, locale: Locale) {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return date.toLocaleTimeString(locale === 'zh' ? 'zh-CN' : 'en-US', { hour: '2-digit', minute: '2-digit' })
}

function runMessageId(run: Run | null) {
  if (!run) return ''
  for (const event of run.events) {
    const value = event.metadata?.message_id
    if (typeof value === 'string' && value.trim()) return value.trim()
  }
  return ''
}

function visibleRunForTranscript(run: Run | null, messages: Message[]) {
  if (!run) return null
  let latestUserIndex = -1
  for (let index = messages.length - 1; index >= 0; index -= 1) {
    if (messages[index].role === 'user') {
      latestUserIndex = index
      break
    }
  }
  if (latestUserIndex < 0) return run
  const latestUser = messages[latestUserIndex]
  const sourceMessageId = runMessageId(run)
  if (sourceMessageId && latestUser.id && sourceMessageId !== latestUser.id) return null
  const messagesAfterLatestUser = messages.slice(latestUserIndex + 1)
  const persistedAssistantAfterLatestUser = messagesAfterLatestUser.some((message) => message.role === 'assistant')
  const persistedAssistantFromAnotherRun = messagesAfterLatestUser.some((message) => message.role === 'assistant' && message.runId && message.runId !== run.id)
  const hasPendingApproval = run.toolCalls?.some((toolCall) => toolCall.status === 'approval_required') ?? false
  if (hasPendingApproval && run.assistantDraft?.status === 'completed') return null
  if ((activeRunStatuses.has(run.status) || hasPendingApproval) && (persistedAssistantFromAnotherRun || persistedAssistantAfterLatestUser)) return null
  return run
}

function MessageActions({ message, locale, canRegenerate, onRetry, onRegenerate }: { message: Message; locale: Locale; canRegenerate?: boolean; onRetry?: () => void; onRegenerate?: () => void }) {
  const copy = getDictionary(locale).chatCanvas
  if (message.role !== 'assistant') return null
  return (
    <div className="message-actions" aria-label={locale === 'zh' ? '消息操作' : 'Message actions'}>
      <button type="button" onClick={() => void navigator.clipboard?.writeText(message.content)} aria-label={copy.copy} title={copy.copy}>
        <Copy size={14} />
        <span>{copy.copy}</span>
      </button>
      {onRetry && (
        <button type="button" onClick={onRetry} aria-label={copy.retry} title={copy.retry}>
          <RotateCcw size={14} />
          <span>{copy.retry}</span>
        </button>
      )}
      {canRegenerate && onRegenerate && (
        <button type="button" onClick={onRegenerate} aria-label={copy.regenerate} title={copy.regenerate}>
          <RefreshCcw size={14} />
          <span>{copy.regenerate}</span>
        </button>
      )}
    </div>
  )
}

function MessageHistory({ messages, locale, canRegenerate, onRegenerate }: { messages: Message[]; locale: Locale; canRegenerate?: boolean; onRegenerate?: () => void }) {
  const copy = getDictionary(locale).chatCanvas
  const lastAssistant = [...messages].reverse().find((message) => message.role === 'assistant')
  return messages.map((message, index) => (
    <article key={`${message.id}-${index}`} className={`message-row ${message.role}`}>
      <div className="message-avatar">{message.role === 'assistant' ? 'L' : 'U'}</div>
      <div className="message-bubble">
        <div className="message-meta">{message.role === 'assistant' ? copy.assistant : copy.user} · {displayMessageTime(message.createdAt, locale)}</div>
        <MarkdownMessage content={message.content} />
        {message.toolCalls?.length ? <ToolCallList toolCalls={message.toolCalls} locale={locale} /> : null}
        <MessageActions message={message} locale={locale} canRegenerate={canRegenerate && message.id === lastAssistant?.id} onRegenerate={onRegenerate} />
      </div>
    </article>
  ))
}

function draftFallback(status: AssistantDraftState['status'], locale: Locale) {
  const copy = getDictionary(locale).chatCanvas
  if (status === 'failed') return copy.failedDetail
  if (status === 'stopped') return copy.stoppedDraft
  if (status === 'recovering') return copy.recoveringDraft
  if (status === 'stopping') return copy.stoppingDetail
  return copy.modelDrafting
}

function draftStatusLabel(status: AssistantDraftState['status'], locale: Locale) {
  const copy = getDictionary(locale).chatCanvas
  if (status === 'streaming' || status === 'drafting') return copy.generating
  if (status === 'completed') return copy.completedTitle
  if (status === 'failed') return copy.failedTitle
  if (status === 'stopped') return copy.stoppedTitle
  if (status === 'recovering') return copy.recoveringTitle
  if (status === 'stopping') return copy.stoppingTitle
  return copy.waitingRunTitle
}

function AssistantDraft({ run, locale, onRetry }: { run: Run | null; locale: Locale; onRetry?: () => void }) {
  const copy = getDictionary(locale).chatCanvas
  const draft = run?.assistantDraft
  if (!run || !draft || draft.status === 'empty') return null

  const failedMessage: Message = { id: draft.messageId ?? run.id, threadId: run.threadId, role: 'assistant', content: draft.content || draftFallback(draft.status, locale), createdAt: run.completedAt ?? run.createdAt ?? new Date().toISOString(), runId: run.id }

  return (
    <article className={`message-row assistant draft ${draft.status}`}>
      <div className="message-avatar">L</div>
      <div className="message-bubble">
        <div className="message-meta">{copy.assistant} · {draftStatusLabel(draft.status, locale)}</div>
        <MarkdownMessage content={failedMessage.content} />
        {draft.status === 'failed' && <MessageActions message={failedMessage} locale={locale} onRetry={onRetry} />}
      </div>
    </article>
  )
}

function ToolCallGroup({ toolCalls, locale }: { toolCalls: NonNullable<Run['toolCalls']>; locale: Locale }) {
  const [expanded, setExpanded] = useState(false)
  const copy = locale === 'zh'
    ? { title: `完成 ${toolCalls.length} 个工具`, details: '查看工具详情' }
    : { title: `${toolCalls.length} tools completed`, details: 'View tool details' }
  const names = [...new Set(toolCalls.map((toolCall) => humanToolName(toolCall.name, locale)))].slice(0, 3).join(' · ')

  return (
    <div className="tool-stack">
      <button className="tool-stack-toggle" type="button" aria-expanded={expanded} onClick={() => setExpanded((value) => !value)}>
        <span>{copy.title}</span>
        {names && <em>{names}</em>}
        <small>{copy.details}</small>
        {expanded ? <ChevronDown size={13} /> : <ChevronRight size={13} />}
      </button>
      {expanded && (
        <div className="tool-stack-list">
          {toolCalls.map((toolCall, index) => <ToolCallCard key={`${toolCall.toolCallId ?? toolCall.id}-${index}`} toolCall={toolCall} locale={locale} />)}
        </div>
      )}
    </div>
  )
}

function ToolCallList({ toolCalls, locale, onApproveToolCall, onDenyToolCall }: { toolCalls: NonNullable<Run['toolCalls']>; locale: Locale; onApproveToolCall?: (toolCall: NonNullable<Run['toolCalls']>[number]) => Promise<void> | void; onDenyToolCall?: (toolCall: NonNullable<Run['toolCalls']>[number]) => Promise<void> | void }) {
  const approvalCalls = toolCalls.filter((toolCall) => toolCall.status === 'approval_required')
  const completedCalls = toolCalls.filter((toolCall) => toolCall.status !== 'approval_required')
  return (
    <>
      {completedCalls.length > 1 ? (
        <ToolCallGroup toolCalls={completedCalls} locale={locale} />
      ) : completedCalls.map((toolCall, index) => <ToolCallCard key={`${toolCall.toolCallId ?? toolCall.id}-${index}`} toolCall={toolCall} locale={locale} />)}
      {approvalCalls.map((toolCall, index) => <ToolCallCard key={`${toolCall.toolCallId ?? toolCall.id}-approval-${index}`} toolCall={toolCall} locale={locale} onApprove={onApproveToolCall} onDeny={onDenyToolCall} />)}
    </>
  )
}

function ActiveToolCalls({ run, locale, onApproveToolCall, onDenyToolCall }: { run: Run | null; locale: Locale; onApproveToolCall?: (toolCall: NonNullable<Run['toolCalls']>[number]) => Promise<void> | void; onDenyToolCall?: (toolCall: NonNullable<Run['toolCalls']>[number]) => Promise<void> | void }) {
  if (!run?.toolCalls?.length) return null
  return (
    <article className="message-row assistant draft tool-call-draft">
      <div className="message-avatar">L</div>
      <div className="message-bubble">
        <ToolCallList toolCalls={run.toolCalls} locale={locale} onApproveToolCall={onApproveToolCall} onDenyToolCall={onDenyToolCall} />
      </div>
    </article>
  )
}

function ToolBoundaryNotice({ run, locale }: { run: Run | null; locale: Locale }) {
  if (!run?.events.some((event) => event.type === 'progress.tool_call_blocked')) return null
  return <div className="api-error">{getDictionary(locale).chatCanvas.toolBoundaryNotice}</div>
}

function ApprovalNotice({ run, locale, onStopRun }: { run: Run | null; locale: Locale; onStopRun: () => void }) {
  if (run?.status !== 'blocked_on_tool_approval' && !run?.toolCalls?.some((toolCall) => toolCall.status === 'approval_required')) return null
  const copy = getDictionary(locale).chatCanvas
  return (
    <div className="approval-notice" role="status">
      <div>
        <strong>{locale === 'zh' ? '等待你确认' : 'Waiting for your confirmation'}</strong>
        <span>{locale === 'zh' ? '确认下面的工具请求后，Loomi 才会继续。' : 'Review the tool request below before Loomi continues.'}</span>
      </div>
      <button type="button" onClick={onStopRun}>{copy.stop}</button>
    </div>
  )
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

export function ChatCanvas({ thread, messages, run, loading, error, dataSourceMode, backendCapability = 'available', backendUnavailableAttempted = false, providerCapabilities = [], workspaceRootConfig, workspaceRootSaveResult, onOpenProviderSettings, onChooseWorkspaceFolder, onSendMessage, onStopRun, onRetryRun, onRegenerateRun, onApproveToolCall, onDenyToolCall, locale }: Props) {
  const visibleRun = visibleRunForTranscript(run, messages)
  const state = deriveChatCanvasState({
    loading,
    error,
    backendCapability,
    backendUnavailableAttempted,
    selectedThreadId: thread?.id ?? null,
    messageCount: messages.length,
    run: visibleRun,
  })
  const copy = getDictionary(locale).chatCanvas
  const stateCopy = createStateCopy(locale)
  const composerDisabled = state === 'loading' || state === 'error' || state === 'no-thread' || state === 'backend-unavailable' || state === 'waiting-run' || state === 'running' || state === 'recovering' || state === 'stopping'
  const mode = thread?.mode ?? 'chat'
  const composerPlaceholder = state === 'history' || !composerDisabled ? (mode === 'work' ? copy.describeTask : copy.messageLoomi) : stateCopy[state].title
  const providerUnavailableBeforeSend = shouldShowProviderUnavailableWarning(dataSourceMode, providerCapabilities)
  const providerUnavailableWarning = getProviderUnavailableWarning(providerCapabilities, locale)
  const workspaceFolderStatus = mode === 'work'
    ? workspaceRootSaveResult?.message || (workspaceRootConfig?.configured ? copy.workspaceRootSelected(workspaceRootConfig.displayName) : copy.workspaceRootHome)
    : undefined
  const hasPersistedCompletedDraftMessage = visibleRun?.assistantDraft?.status === 'completed' && messages.some((message) => (
    message.role === 'assistant' && (
      message.id === visibleRun.assistantDraft?.messageId ||
      message.runId === visibleRun.id ||
      (message.threadId === visibleRun.threadId && message.content === visibleRun.assistantDraft?.content)
    )
  ))
  const shouldShowAssistantDraft = Boolean(visibleRun && !hasPersistedCompletedDraftMessage)
  const shouldShowHistory = state === 'history' || state === 'waiting-run' || state === 'running' || state === 'completed' || state === 'failed' || state === 'stopped' || state === 'recovering' || state === 'stopping'
  const workPlanProjection = deriveWorkPlanProjection(thread, messages, visibleRun)
  const composerModelOptions: ComposerModelOption[] = providerCapabilities
    .filter((provider) => provider.status === 'available' && provider.executionState !== 'unsupported')
    .map((provider) => ({
      key: `${provider.id}:${provider.model}`,
      providerId: provider.id,
      model: provider.model,
      label: `${provider.model} · ${provider.localProvider ? 'Local' : provider.family}`,
    }))
  const canRegenerateAnswer = Boolean(thread && visibleRun && !activeRunStatuses.has(visibleRun.status) && messages.some((message) => message.role === 'assistant'))

  return (
    <section className="chat-shell glass-panel" data-chat-state={state}>
      {error && <div className="api-error">{error}</div>}
      <ToolBoundaryNotice run={visibleRun} locale={locale} />
      <ApprovalNotice run={visibleRun} locale={locale} onStopRun={onStopRun} />
      {providerUnavailableBeforeSend && (
        <div className="provider-warning" role="status">
          <span>{providerUnavailableWarning}</span>
          <button type="button" onClick={onOpenProviderSettings}>{copy.openProviderSettings}</button>
        </div>
      )}

      <div className="message-list">
        {workPlanProjection && <WorkPlanView projection={workPlanProjection} loading={loading} error={error} locale={locale} />}
        {state === 'history' ? (
          <>
            <MessageHistory messages={messages} locale={locale} canRegenerate={canRegenerateAnswer} onRegenerate={onRegenerateRun} />
            {shouldShowAssistantDraft && <AssistantDraft run={visibleRun} locale={locale} onRetry={onRetryRun} />}
            <ActiveToolCalls run={visibleRun} locale={locale} onApproveToolCall={onApproveToolCall} onDenyToolCall={onDenyToolCall} />
          </>
        ) : (
          <>
            {shouldShowHistory && <MessageHistory messages={messages} locale={locale} canRegenerate={canRegenerateAnswer} onRegenerate={onRegenerateRun} />}
            {shouldShowAssistantDraft && <AssistantDraft run={visibleRun} locale={locale} onRetry={onRetryRun} />}
            <ActiveToolCalls run={visibleRun} locale={locale} onApproveToolCall={onApproveToolCall} onDenyToolCall={onDenyToolCall} />
            {(state === 'no-thread' || state === 'empty-thread' || state === 'loading' || state === 'error' || state === 'backend-unavailable') && <StatePanel state={state} error={state === 'error' ? error : null} locale={locale} />}
          </>
        )}
      </div>

      <Composer
        disabled={composerDisabled}
        providerUnavailable={providerUnavailableBeforeSend}
        placeholder={composerPlaceholder}
        mode={mode}
        threadSelected={Boolean(thread)}
        run={visibleRun}
        messages={messages}
        modelOptions={composerModelOptions}
        stopLabel={copy.stop}
        modelLabel={copy.model}
        modelUnavailableLabel={copy.modelUnavailable}
        attachLabel={copy.attach}
        pasteImageLabel={copy.pasteImage}
        attachmentPendingLabel={copy.attachmentPending}
        workspaceFolderLabel={copy.chooseWorkspaceFolder}
        workspaceFolderStatus={workspaceFolderStatus}
        onSubmit={onSendMessage}
        onStop={onStopRun}
        onRetry={onRetryRun}
        onRegenerate={onRegenerateRun}
        onChooseWorkspaceFolder={onChooseWorkspaceFolder}
      />
    </section>
  )
}
