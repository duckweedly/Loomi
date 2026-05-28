import type { ReactNode } from 'react'
import { Fragment, useEffect, useState } from 'react'
import { Copy, ChevronDown, ChevronRight, RefreshCcw, RotateCcw } from 'lucide-react'
import { Divider, Typewriter } from 'animal-island-ui'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import type { Components } from 'react-markdown'
import type { AssistantDraft as AssistantDraftState, BackendCapabilityState, ChatCanvasState, Message, Persona, ProviderCapability, Run, RunEvent, StreamState, Thread, ToolCall, WorkspaceRootConfig } from '../domain'
import type { Locale } from '../i18n'
import { getDictionary } from '../i18n'
import { getProviderUnavailableWarning, shouldShowProviderUnavailableWarning } from '../runtime/backendCapabilityStatus'
import { deriveChatCanvasState } from '../runtime/chatCanvasState'
import type { DesktopReadiness } from '../runtime/desktopReadiness'
import { normalizeMarkdownContent } from '../runtime/markdownNormalize'
import type { PreviewArtifact } from '../runtime/artifactPreview'
import { extractMessageArtifact, stripMessageArtifactSource } from '../runtime/messageArtifactPreview'
import { thinkingHintForRun } from '../runtime/thinkingHint'
import { humanToolName } from '../runtime/toolPreview'
import type { ProviderSaveResult } from '../state'
import { Composer } from './Composer'
import type { ComposerAttachment, ComposerModelOption } from './Composer'
import { ToolCallCard } from './ToolCallCard'

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
  desktopReadiness?: DesktopReadiness
  personas?: Persona[]
  selectedPersonaId?: string
  onSelectPersona?: (personaId: string) => void
  onOpenProviderSettings?: () => void
  onRetryReadiness?: () => void
  onDetectLocalProviders?: () => void
  onEnableLocalProvider?: (providerId: string) => void
  onOpenSkillsSettings?: () => void
  onOpenConnectorsSettings?: () => void
  onOpenPluginsSettings?: () => void
  onChooseWorkspaceFolder?: () => void
  onSendMessage: (content: string, options?: { providerId?: string; model?: string; attachments?: ComposerAttachment[] }) => void
  onStopRun: () => void
  onRetryRun?: () => void
  onRegenerateRun?: () => void
  onApproveToolCall?: (toolCall: NonNullable<Run['toolCalls']>[number]) => Promise<void> | void
  onDenyToolCall?: (toolCall: NonNullable<Run['toolCalls']>[number]) => Promise<void> | void
  onOpenArtifact?: (artifact: PreviewArtifact) => void
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

function textFromChildren(children: ReactNode): string {
  if (typeof children === 'string') return children
  if (typeof children === 'number') return String(children)
  if (Array.isArray(children)) return children.map(textFromChildren).join('')
  return ''
}

function headingText(children: ReactNode) {
  return textFromChildren(children).replace(/^#{1,6}\s+/, '')
}

function codeLanguageLabel(className?: string) {
  return className?.match(/language-([\w+-]+)/)?.[1]
}

function MarkdownCodeBlock({ className, children }: { className?: string; children?: ReactNode }) {
  const text = textFromChildren(children).replace(/\n$/, '')
  const language = codeLanguageLabel(className)
  return (
    <div className="message-code-block">
      <div className="message-code-block-head">
        {language && <span className="message-code-block-lang">{language}</span>}
        <button type="button" onClick={() => void navigator.clipboard?.writeText(text)} aria-label="Copy code">
          <Copy size={13} />
        </button>
      </div>
      <pre><code className={className}>{text}</code></pre>
    </div>
  )
}

const markdownComponents: Components = {
  a: ({ href, children }) => <a href={safeHref(href ?? '')} rel="noreferrer" target="_blank">{children}</a>,
  h1: ({ children }) => <h1>{headingText(children)}</h1>,
  h2: ({ children }) => <h2>{headingText(children)}</h2>,
  h3: ({ children }) => <h3>{headingText(children)}</h3>,
  pre: ({ children }) => {
    const child = Array.isArray(children) ? children[0] : children
    if (typeof child === 'object' && child && 'props' in child) {
      const props = child.props as { className?: string; children?: ReactNode }
      return <MarkdownCodeBlock className={props.className}>{props.children}</MarkdownCodeBlock>
    }
    return <MarkdownCodeBlock>{children}</MarkdownCodeBlock>
  },
  code: ({ className, children }) => <code className={className}>{children}</code>,
  table: ({ children }) => <div className="message-table-wrap"><table>{children}</table></div>,
}

const assistantTypewriterSpeed = 28
const completedTypewriterStorageKey = 'loomi.completedTypewriterMessages'
const streamedAssistantRunsStorageKey = 'loomi.streamedAssistantRuns'
const completedTypewriterMessages = new Set<string>()
const streamedAssistantRuns = new Set<string>()

function shouldAutoPlayTypewriter() {
  return typeof window !== 'undefined' && !window.matchMedia('(prefers-reduced-motion: reduce)').matches
}

function hasCompletedTypewriter(trigger: string) {
  if (completedTypewriterMessages.has(trigger)) return true
  if (typeof window === 'undefined') return false
  try {
    const completed = JSON.parse(window.sessionStorage.getItem(completedTypewriterStorageKey) ?? '[]') as string[]
    completed.forEach((item) => completedTypewriterMessages.add(item))
    return completedTypewriterMessages.has(trigger)
  } catch {
    return false
  }
}

function markCompletedTypewriter(trigger: string) {
  completedTypewriterMessages.add(trigger)
  if (typeof window === 'undefined') return
  try {
    window.sessionStorage.setItem(completedTypewriterStorageKey, JSON.stringify([...completedTypewriterMessages].slice(-100)))
  } catch {
    // Ignore storage failures; the in-memory set still prevents replay in this app session.
  }
}

function hasStreamedAssistantRun(runId: string) {
  if (streamedAssistantRuns.has(runId)) return true
  if (typeof window === 'undefined') return false
  try {
    const streamed = JSON.parse(window.sessionStorage.getItem(streamedAssistantRunsStorageKey) ?? '[]') as string[]
    streamed.forEach((item) => streamedAssistantRuns.add(item))
    return streamedAssistantRuns.has(runId)
  } catch {
    return false
  }
}

function markStreamedAssistantRun(runId: string) {
  streamedAssistantRuns.add(runId)
  if (typeof window === 'undefined') return
  try {
    window.sessionStorage.setItem(streamedAssistantRunsStorageKey, JSON.stringify([...streamedAssistantRuns].slice(-100)))
  } catch {
    // Best effort only; the current app session still knows this run streamed.
  }
}

function MarkdownMessage({ content, typewriterTrigger }: { content: string; typewriterTrigger?: string }) {
  const markdown = (
    <ReactMarkdown remarkPlugins={[remarkGfm]} components={markdownComponents}>
      {normalizeMarkdownContent(content)}
    </ReactMarkdown>
  )
  const [shouldPlayTypewriter, setShouldPlayTypewriter] = useState(() => Boolean(typewriterTrigger && shouldAutoPlayTypewriter() && !hasCompletedTypewriter(typewriterTrigger)))
  if (!typewriterTrigger || !shouldPlayTypewriter) {
    return <div className="message-markdown">{markdown}</div>
  }
  return (
    <div className="message-markdown message-markdown-typewriter">
      <Typewriter
        speed={assistantTypewriterSpeed}
        trigger={typewriterTrigger}
        onDone={() => {
          markCompletedTypewriter(typewriterTrigger)
          setShouldPlayTypewriter(false)
        }}
      >
        {markdown}
      </Typewriter>
    </div>
  )
}

function MessageArtifactCard({ artifact, locale, onOpenArtifact }: { artifact: PreviewArtifact; locale: Locale; onOpenArtifact?: (artifact: PreviewArtifact) => void }) {
  return (
    <button
      type="button"
      className="artifact-resource-card message-artifact-card"
      aria-label={locale === 'zh' ? `预览 ${artifact.title}` : `Preview ${artifact.title}`}
      onClick={() => onOpenArtifact?.(artifact)}
    >
      <span className="artifact-resource-icon"><Copy size={16} /></span>
      <span className="artifact-resource-copy">
        <strong>{artifact.title}</strong>
        <small>{locale === 'zh' ? 'Markdown 文档' : 'Markdown document'}</small>
      </span>
    </button>
  )
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

function shouldTypewriteHistoryMessage(message: Message, run: Run | null, isLastAssistant: boolean) {
  if (!isLastAssistant || !run || run.status !== 'completed' || message.role !== 'assistant') return false
  if (hasStreamedAssistantRun(run.id)) return false
  if (message.runId && message.runId === run.id) return true
  return Boolean(run.assistantDraft?.status === 'completed' && message.content === run.assistantDraft.content)
}

function ConversationDivider() {
  return (
    <div className="conversation-divider" aria-hidden="true">
      <Divider type="wave-yellow" />
    </div>
  )
}

function MessageHistory({ messages, run, locale, canRegenerate, onRegenerate, onOpenArtifact }: { messages: Message[]; run: Run | null; locale: Locale; canRegenerate?: boolean; onRegenerate?: () => void; onOpenArtifact?: (artifact: PreviewArtifact) => void }) {
  const copy = getDictionary(locale).chatCanvas
  const lastAssistant = [...messages].reverse().find((message) => message.role === 'assistant')
  return messages.map((message, index) => {
    const typewriterTrigger = shouldTypewriteHistoryMessage(message, run, message.id === lastAssistant?.id) ? `${message.id}:${run?.id}:${message.content.length}` : undefined
    const showTurnDivider = index > 0 && message.role === 'user'
    const artifact = message.role === 'assistant' ? extractMessageArtifact(message) : null
    const visibleContent = artifact ? stripMessageArtifactSource(message.content) : message.content
    return (
      <Fragment key={`${message.id}-${index}`}>
        {showTurnDivider && <ConversationDivider />}
        <article className={`message-row ${message.role}`}>
          <div className="message-avatar">{message.role === 'assistant' ? 'L' : 'U'}</div>
          <div className="message-bubble">
            <div className="message-meta">{message.role === 'assistant' ? copy.assistant : copy.user} · {displayMessageTime(message.createdAt, locale)}</div>
            <MarkdownMessage content={visibleContent} typewriterTrigger={typewriterTrigger} />
            {artifact && <MessageArtifactCard artifact={artifact} locale={locale} onOpenArtifact={onOpenArtifact} />}
            {message.toolCalls?.length ? <ToolCallList toolCalls={message.toolCalls} locale={locale} onOpenArtifact={onOpenArtifact} /> : null}
            <MessageActions message={message} locale={locale} canRegenerate={canRegenerate && message.id === lastAssistant?.id} onRegenerate={onRegenerate} />
          </div>
        </article>
      </Fragment>
    )
  })
}

function draftFallback(status: AssistantDraftState['status'], locale: Locale) {
  const copy = getDictionary(locale).chatCanvas
  if (status === 'failed') return copy.failedDetail
  if (status === 'stopped') return copy.stoppedDraft
  if (status === 'recovering') return copy.recoveringDraft
  if (status === 'stopping') return copy.stoppingDetail
  return copy.modelDrafting
}

function draftPendingText(run: Run, status: AssistantDraftState['status'], locale: Locale) {
  if (status === 'recovering' || status === 'stopping') return draftFallback(status, locale)
  return thinkingHintForRun(run.id, locale)
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

type TranscriptBlock =
  | { id: string; kind: 'assistant'; content: string; status: AssistantDraftState['status']; time: string }
  | { id: string; kind: 'tools'; toolCalls: ToolCall[] }
  | { id: string; kind: 'waiting' }

function eventText(event: RunEvent) {
  if (event.type !== 'model.delta' && event.type !== 'message.model_output_delta' && event.type !== 'assistant.drafting') return ''
  return event.assistantDelta ?? event.content ?? ''
}

function toolStatusForEvent(event: RunEvent): ToolCall['status'] | null {
  switch (event.type) {
    case 'tool.call.requested':
      return 'requested'
    case 'tool.call.approval_required':
      return 'approval_required'
    case 'tool.call.approved':
      return 'approved'
    case 'tool.call.executing':
      return 'executing'
    case 'tool.call.succeeded':
      return 'succeeded'
    case 'tool.call.failed':
      return 'failed'
    case 'tool.call.denied':
      return 'denied'
    case 'tool.call.cancelled':
      return 'cancelled'
    default:
      return null
  }
}

function metadataRecord(value: unknown): Record<string, unknown> | undefined {
  return typeof value === 'object' && value !== null && !Array.isArray(value) ? value as Record<string, unknown> : undefined
}

function metadataString(value: unknown): string | undefined {
  return typeof value === 'string' && value.trim() ? value : undefined
}

function toolCallFromEvent(event: RunEvent, previous?: ToolCall): ToolCall | null {
  const status = toolStatusForEvent(event)
  if (!status) return null
  const toolCallId = metadataString(event.metadata?.tool_call_id) ?? previous?.toolCallId ?? event.id
  return {
    id: previous?.id ?? event.id,
    toolCallId,
    name: metadataString(event.metadata?.tool_name) ?? previous?.name ?? event.label,
    status,
    approvalStatus: metadataString(event.metadata?.approval_status) as ToolCall['approvalStatus'] ?? previous?.approvalStatus,
    executionStatus: metadataString(event.metadata?.execution_status) as ToolCall['executionStatus'] ?? previous?.executionStatus,
    summary: event.detail || previous?.summary || '',
    input: previous?.input ?? '',
    output: event.content ?? previous?.output ?? '',
    argumentsSummary: metadataRecord(event.metadata?.arguments_summary) ?? previous?.argumentsSummary,
    resultSummary: metadataRecord(event.metadata?.result_summary) ?? previous?.resultSummary,
    errorCode: metadataString(event.metadata?.error_code) ?? previous?.errorCode,
    errorMessage: metadataString(event.metadata?.error_message) ?? previous?.errorMessage,
  }
}

function isTerminalToolStatus(status: ToolCall['status']) {
  return status === 'succeeded' || status === 'failed' || status === 'denied' || status === 'cancelled'
}

function isActiveRunStatus(status: Run['status']) {
  return status === 'running' || status === 'queued' || status === 'blocked_on_tool_approval'
}

function buildRunTranscriptBlocks(run: Run | null): TranscriptBlock[] {
  if (!run?.events.length) return []
  const blocks: TranscriptBlock[] = []
  const toolBlockIndex = new Map<string, { blockIndex: number; toolIndex: number }>()
  let currentToolGroupIndex: number | null = null
  let text = ''
  let textStartId = ''
  let textTime = run.createdAt ?? new Date().toISOString()

  const flushText = (status: AssistantDraftState['status'] = 'streaming') => {
    if (!text.trim()) return
    blocks.push({ id: textStartId || `assistant-${blocks.length}`, kind: 'assistant', content: text, status, time: textTime })
    text = ''
    textStartId = ''
    currentToolGroupIndex = null
  }

  for (const event of [...run.events].sort((a, b) => (a.sequence ?? 0) - (b.sequence ?? 0))) {
    const delta = eventText(event)
    if (delta) {
      if (!text) {
        textStartId = event.id
        textTime = event.time
      }
      text += delta
      continue
    }

    const toolCall = toolCallFromEvent(event)
    if (toolCall) {
      flushText('paused_for_tool')
      const blockKey = toolCall.toolCallId ?? toolCall.id
      const existingLocation = toolBlockIndex.get(blockKey)
      if (existingLocation === undefined) {
        let targetIndex: number
        if (currentToolGroupIndex !== null && blocks[currentToolGroupIndex]?.kind === 'tools') {
          targetIndex = currentToolGroupIndex
        } else {
          targetIndex = blocks.length
          currentToolGroupIndex = targetIndex
          blocks.push({ id: blockKey, kind: 'tools', toolCalls: [] })
        }
        const target = blocks[targetIndex] as Extract<TranscriptBlock, { kind: 'tools' }>
        toolBlockIndex.set(blockKey, { blockIndex: targetIndex, toolIndex: target.toolCalls.length })
        target.toolCalls.push(toolCall)
      } else {
        const target = blocks[existingLocation.blockIndex] as Extract<TranscriptBlock, { kind: 'tools' }>
        target.toolCalls[existingLocation.toolIndex] = { ...target.toolCalls[existingLocation.toolIndex], ...toolCall }
      }
    }
  }

  flushText(run.assistantDraft?.status ?? 'streaming')
  const lastBlock = blocks.at(-1)
  if (lastBlock?.kind === 'tools' && isActiveRunStatus(run.status) && lastBlock.toolCalls.length > 0 && lastBlock.toolCalls.every((toolCall) => isTerminalToolStatus(toolCall.status)) && !run.assistantDraft?.content?.trim()) {
    blocks.push({ id: `${lastBlock.id}-waiting`, kind: 'waiting' })
  }
  return blocks
}

function RunTranscript({ run, locale, onApproveToolCall, onDenyToolCall, onOpenArtifact }: { run: Run | null; locale: Locale; onApproveToolCall?: (toolCall: NonNullable<Run['toolCalls']>[number]) => Promise<void> | void; onDenyToolCall?: (toolCall: NonNullable<Run['toolCalls']>[number]) => Promise<void> | void; onOpenArtifact?: (artifact: PreviewArtifact) => void }) {
  const blocks = buildRunTranscriptBlocks(run)
  if (!run || blocks.length === 0) return null
  const copy = getDictionary(locale).chatCanvas
  return (
    <>
      {blocks.map((block) => (
        <article className={block.kind === 'assistant' ? `message-row assistant draft ${block.status}` : block.kind === 'waiting' ? 'message-row assistant draft drafting' : 'message-row assistant draft tool-call-draft tool-transcript-group'} key={block.id}>
          <div className="message-avatar">L</div>
          <div className="message-bubble">
            {block.kind === 'assistant' ? (
              <>
                <div className="message-meta">{copy.assistant} · {draftStatusLabel(block.status, locale)}</div>
                <MarkdownMessage content={block.content} />
              </>
            ) : block.kind === 'waiting' ? (
              <>
                <div className="message-meta">{copy.assistant} · {draftStatusLabel('drafting', locale)}</div>
                <div className="message-draft-status" role="status">
                  <span aria-hidden="true" />
                  <p className="thinking-shimmer">{draftPendingText(run, 'drafting', locale)}</p>
                </div>
              </>
            ) : <ToolCallList toolCalls={block.toolCalls} locale={locale} onApproveToolCall={onApproveToolCall} onDenyToolCall={onDenyToolCall} onOpenArtifact={onOpenArtifact} />}
          </div>
        </article>
      ))}
    </>
  )
}

function shouldRenderDraftContent(status: AssistantDraftState['status'], content: string) {
  return status === 'completed' || status === 'failed' || status === 'stopped' || Boolean(content.trim())
}

function AssistantDraft({ run, locale, onRetry, onOpenArtifact }: { run: Run | null; locale: Locale; onRetry?: () => void; onOpenArtifact?: (artifact: PreviewArtifact) => void }) {
  const copy = getDictionary(locale).chatCanvas
  const draft = run?.assistantDraft

  useEffect(() => {
    if (!run || !draft) return
    if (draft.content.trim() && (draft.status === 'streaming' || draft.status === 'drafting')) {
      markStreamedAssistantRun(run.id)
    }
  }, [draft?.content, draft?.status, run?.id])

  if (!run || !draft || draft.status === 'empty') return null

  const shouldRenderContent = shouldRenderDraftContent(draft.status, draft.content)
  const draftMessage: Message = { id: draft.messageId ?? run.id, threadId: run.threadId, role: 'assistant', content: draft.content || draftFallback(draft.status, locale), createdAt: run.completedAt ?? run.createdAt ?? new Date().toISOString(), runId: run.id }
  const artifact = extractMessageArtifact(draftMessage)
  const visibleContent = artifact ? stripMessageArtifactSource(draftMessage.content) : draftMessage.content
  const typewriterTrigger = draft.status === 'completed' && !hasStreamedAssistantRun(run.id) ? `${draftMessage.id}:${draftMessage.content.length}` : undefined

  return (
    <article className={`message-row assistant draft ${draft.status}`}>
      <div className="message-avatar">L</div>
      <div className="message-bubble">
        <div className="message-meta">{copy.assistant} · {draftStatusLabel(draft.status, locale)}</div>
        {shouldRenderContent ? (
          <>
            <MarkdownMessage content={visibleContent} typewriterTrigger={typewriterTrigger} />
            {artifact && <MessageArtifactCard artifact={artifact} locale={locale} onOpenArtifact={onOpenArtifact} />}
          </>
        ) : (
          <div className="message-draft-status" role="status">
            <span aria-hidden="true" />
            <p className="thinking-shimmer">{draftPendingText(run, draft.status, locale)}</p>
          </div>
        )}
        {draft.status === 'failed' && <MessageActions message={draftMessage} locale={locale} onRetry={onRetry} />}
      </div>
    </article>
  )
}

function ToolCallGroup({ toolCalls, locale, onOpenArtifact }: { toolCalls: NonNullable<Run['toolCalls']>; locale: Locale; onOpenArtifact?: (artifact: PreviewArtifact) => void }) {
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
          {toolCalls.map((toolCall, index) => <ToolCallCard key={`${toolCall.toolCallId ?? toolCall.id}-${index}`} toolCall={toolCall} locale={locale} onOpenArtifact={onOpenArtifact} />)}
        </div>
      )}
    </div>
  )
}

function ToolCallList({ toolCalls, locale, onApproveToolCall, onDenyToolCall, onOpenArtifact }: { toolCalls: NonNullable<Run['toolCalls']>; locale: Locale; onApproveToolCall?: (toolCall: NonNullable<Run['toolCalls']>[number]) => Promise<void> | void; onDenyToolCall?: (toolCall: NonNullable<Run['toolCalls']>[number]) => Promise<void> | void; onOpenArtifact?: (artifact: PreviewArtifact) => void }) {
  const approvalCalls = toolCalls.filter((toolCall) => toolCall.status === 'approval_required')
  const completedCalls = toolCalls.filter((toolCall) => toolCall.status !== 'approval_required')
  return (
    <>
      {completedCalls.length > 1 ? (
        <ToolCallGroup toolCalls={completedCalls} locale={locale} onOpenArtifact={onOpenArtifact} />
      ) : completedCalls.map((toolCall, index) => <ToolCallCard key={`${toolCall.toolCallId ?? toolCall.id}-${index}`} toolCall={toolCall} locale={locale} onOpenArtifact={onOpenArtifact} />)}
      {approvalCalls.map((toolCall, index) => <ToolCallCard key={`${toolCall.toolCallId ?? toolCall.id}-approval-${index}`} toolCall={toolCall} locale={locale} onApprove={onApproveToolCall} onDeny={onDenyToolCall} onOpenArtifact={onOpenArtifact} />)}
    </>
  )
}

function ActiveToolCalls({ run, locale, onApproveToolCall, onDenyToolCall, onOpenArtifact }: { run: Run | null; locale: Locale; onApproveToolCall?: (toolCall: NonNullable<Run['toolCalls']>[number]) => Promise<void> | void; onDenyToolCall?: (toolCall: NonNullable<Run['toolCalls']>[number]) => Promise<void> | void; onOpenArtifact?: (artifact: PreviewArtifact) => void }) {
  if (!run?.toolCalls?.length) return null
  return (
    <article className="message-row assistant draft tool-call-draft">
      <div className="message-avatar">L</div>
      <div className="message-bubble">
        <ToolCallList toolCalls={run.toolCalls} locale={locale} onApproveToolCall={onApproveToolCall} onDenyToolCall={onDenyToolCall} onOpenArtifact={onOpenArtifact} />
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

function DesktopReadinessPanel({ readiness, onRetry, onOpenSettings, onDetectLocalProviders, onEnableLocalProvider, onChooseWorkspaceFolder }: { readiness?: DesktopReadiness; onRetry?: () => void; onOpenSettings?: () => void; onDetectLocalProviders?: () => void; onEnableLocalProvider?: (providerId: string) => void; onChooseWorkspaceFolder?: () => void }) {
  if (!readiness || readiness.primary.code === 'ready') return null
  const issue = readiness.primary
  const enableLocalCodex = () => onEnableLocalProvider?.(issue.providerId ?? 'local_codex')
  const labels = issue.title.includes('未') || issue.title.includes('不可') || issue.title.includes('连接') || issue.title.includes('配置')
    ? { retry: '重试', settings: '打开设置', detect: '检测 Local Provider', enable: '启用 Local Codex', folder: '选择目录' }
    : { retry: 'Retry', settings: 'Open Settings', detect: 'Detect Local Provider', enable: 'Enable Local Codex', folder: 'Choose folder' }
  return (
    <div className="api-error desktop-readiness" role="status">
      <strong>{issue.title}</strong>
      <span>{issue.detail}</span>
      <div className="desktop-readiness-actions">
        <button type="button" onClick={onRetry}>{labels.retry}</button>
        <button type="button" onClick={onOpenSettings}>{labels.settings}</button>
        <button type="button" onClick={onDetectLocalProviders}>{labels.detect}</button>
        {issue.code === 'local_codex_detected_disabled' && <button type="button" onClick={enableLocalCodex}>{labels.enable}</button>}
        {issue.code === 'workspace_unselected' && <button type="button" onClick={onChooseWorkspaceFolder}>{labels.folder}</button>}
      </div>
    </div>
  )
}

function MissingFinalNotice({ locale }: { locale: Locale }) {
  const copy = getDictionary(locale).chatCanvas
  return (
    <div className="runtime-final-warning" role="status">
      <strong>{copy.missingFinalTitle}</strong>
      <span>{copy.missingFinalDetail}</span>
    </div>
  )
}

export function ChatCanvas({ thread, messages, run, loading, error, dataSourceMode, backendCapability = 'available', backendUnavailableAttempted = false, providerCapabilities = [], workspaceRootConfig, workspaceRootSaveResult, desktopReadiness, onOpenProviderSettings, onRetryReadiness, onDetectLocalProviders, onEnableLocalProvider, onOpenSkillsSettings, onOpenConnectorsSettings, onOpenPluginsSettings, onChooseWorkspaceFolder, onSendMessage, onStopRun, onRetryRun, onRegenerateRun, onApproveToolCall, onDenyToolCall, onOpenArtifact, locale }: Props) {
  const readinessError = desktopReadiness && desktopReadiness.primary.code !== 'ready'
  const visibleError = readinessError ? null : error
  const visibleRun = visibleRunForTranscript(run, messages)
  const state = deriveChatCanvasState({
    loading,
    error: visibleError,
    backendCapability,
    backendUnavailableAttempted,
    selectedThreadId: thread?.id ?? null,
    messageCount: messages.length,
    run: visibleRun,
  })
  const copy = getDictionary(locale).chatCanvas
  const stateCopy = createStateCopy(locale)
  const composerDisabled = state === 'loading' || state === 'error' || state === 'no-thread' || state === 'backend-unavailable' || state === 'waiting-run' || state === 'running' || state === 'recovering' || state === 'stopping'
  const composerPlaceholder = state === 'history' || !composerDisabled ? copy.messageLoomi : stateCopy[state].title
  const providerUnavailableBeforeSend = shouldShowProviderUnavailableWarning(dataSourceMode, providerCapabilities)
  const providerUnavailableWarning = getProviderUnavailableWarning(providerCapabilities, locale)
  const workspaceFolderStatus = workspaceRootSaveResult?.message || (workspaceRootConfig?.configured ? copy.workspaceRootSelected(workspaceRootConfig.displayName) : copy.workspaceRootHome)
  const hasPersistedCompletedDraftMessage = visibleRun?.assistantDraft?.status === 'completed' && messages.some((message) => (
    message.role === 'assistant' && (
      message.id === visibleRun.assistantDraft?.messageId ||
      message.runId === visibleRun.id ||
      (message.threadId === visibleRun.threadId && message.content === visibleRun.assistantDraft?.content)
    )
  ))
  const hasAssistantFinalContent = Boolean(visibleRun?.status === 'completed' && (
    messages.some((message) => message.role === 'assistant' && message.content.trim() && (!message.runId || message.runId === visibleRun.id)) ||
    (visibleRun.assistantDraft?.status === 'completed' && visibleRun.assistantDraft.content.trim())
  ))
  const missingFinalContent = dataSourceMode === 'real_api' && visibleRun?.status === 'completed' && !hasAssistantFinalContent
  const hasRunTranscript = buildRunTranscriptBlocks(visibleRun).length > 0
  const shouldShowAssistantDraft = Boolean(visibleRun && !hasPersistedCompletedDraftMessage && !hasRunTranscript)
  const shouldShowHistory = state === 'history' || state === 'waiting-run' || state === 'running' || state === 'completed' || state === 'failed' || state === 'stopped' || state === 'recovering' || state === 'stopping'
  const composerModelOptions: ComposerModelOption[] = providerCapabilities
    .filter((provider) => ['available', 'configured', 'reachable', 'completion-ok'].includes(provider.status) && provider.executionState !== 'unsupported')
    .map((provider) => ({
      key: `${provider.id}:${provider.model}`,
      providerId: provider.id,
      model: provider.model,
      label: `${provider.model} · ${provider.localProvider ? 'Local' : provider.family}`,
    }))
  const canRegenerateAnswer = Boolean(thread && visibleRun && !activeRunStatuses.has(visibleRun.status) && messages.some((message) => message.role === 'assistant'))

  return (
    <section className="chat-shell glass-panel" data-chat-state={state}>
      <DesktopReadinessPanel readiness={desktopReadiness} onRetry={onRetryReadiness} onOpenSettings={onOpenProviderSettings} onDetectLocalProviders={onDetectLocalProviders} onEnableLocalProvider={onEnableLocalProvider} onChooseWorkspaceFolder={onChooseWorkspaceFolder} />
      {visibleError && <div className="api-error">{visibleError}</div>}
      <ToolBoundaryNotice run={visibleRun} locale={locale} />
      <ApprovalNotice run={visibleRun} locale={locale} onStopRun={onStopRun} />
      {missingFinalContent && <MissingFinalNotice locale={locale} />}
      {providerUnavailableBeforeSend && (
        <div className="provider-warning" role="status">
          <span>{providerUnavailableWarning}</span>
          <button type="button" onClick={onOpenProviderSettings}>{copy.openProviderSettings}</button>
        </div>
      )}

      <div className="message-list">
        {state === 'history' ? (
          <>
            <MessageHistory messages={messages} run={visibleRun} locale={locale} canRegenerate={canRegenerateAnswer} onRegenerate={onRegenerateRun} onOpenArtifact={onOpenArtifact} />
            <RunTranscript run={visibleRun} locale={locale} onApproveToolCall={onApproveToolCall} onDenyToolCall={onDenyToolCall} onOpenArtifact={onOpenArtifact} />
            {shouldShowAssistantDraft && <AssistantDraft run={visibleRun} locale={locale} onRetry={onRetryRun} onOpenArtifact={onOpenArtifact} />}
            {!hasRunTranscript && <ActiveToolCalls run={visibleRun} locale={locale} onApproveToolCall={onApproveToolCall} onDenyToolCall={onDenyToolCall} onOpenArtifact={onOpenArtifact} />}
          </>
        ) : (
          <>
            {shouldShowHistory && <MessageHistory messages={messages} run={visibleRun} locale={locale} canRegenerate={canRegenerateAnswer} onRegenerate={onRegenerateRun} onOpenArtifact={onOpenArtifact} />}
            <RunTranscript run={visibleRun} locale={locale} onApproveToolCall={onApproveToolCall} onDenyToolCall={onDenyToolCall} onOpenArtifact={onOpenArtifact} />
            {shouldShowAssistantDraft && <AssistantDraft run={visibleRun} locale={locale} onRetry={onRetryRun} onOpenArtifact={onOpenArtifact} />}
            {!hasRunTranscript && <ActiveToolCalls run={visibleRun} locale={locale} onApproveToolCall={onApproveToolCall} onDenyToolCall={onDenyToolCall} onOpenArtifact={onOpenArtifact} />}
            {(state === 'no-thread' || state === 'empty-thread' || state === 'loading' || state === 'error' || state === 'backend-unavailable') && <StatePanel state={state} error={state === 'error' ? visibleError : null} locale={locale} />}
          </>
        )}
      </div>

      <Composer
        disabled={composerDisabled}
        providerUnavailable={providerUnavailableBeforeSend}
        placeholder={composerPlaceholder}
        threadSelected={Boolean(thread)}
        run={visibleRun}
        messages={messages}
        modelOptions={composerModelOptions}
        stopLabel={copy.stop}
        modelLabel={copy.model}
        modelUnavailableLabel={copy.modelUnavailable}
        attachLabel={copy.attach}
        addFilesAndPhotosLabel={copy.addFilesAndPhotos}
        addFolderLabel={copy.addFolder}
        skillsLabel={copy.skills}
        connectorsLabel={copy.connectors}
        addPluginsLabel={copy.addPlugins}
        contextMenuLabel={copy.contextMenu}
        pasteImageLabel={copy.pasteImage}
        attachmentPendingLabel={copy.attachmentPending}
        workspaceFolderLabel={copy.chooseWorkspaceFolder}
        workspaceFolderStatus={workspaceFolderStatus}
        onSubmit={onSendMessage}
        onStop={onStopRun}
        onRetry={onRetryRun}
        onRegenerate={onRegenerateRun}
        onChooseWorkspaceFolder={onChooseWorkspaceFolder}
        onOpenSkills={onOpenSkillsSettings}
        onOpenConnectors={onOpenConnectorsSettings}
        onOpenPlugins={onOpenPluginsSettings}
      />
    </section>
  )
}
