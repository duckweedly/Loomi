import { useState } from 'react'
import { Button } from 'animal-island-ui'
import { AlertTriangle, CheckCircle2, ChevronDown, ChevronRight, Clock3, FileText, FolderSearch, Search, ShieldCheck, Terminal, XCircle } from 'lucide-react'
import type { ToolCall } from '../domain'
import type { Locale } from '../i18n'
import { getToolCallArtifact, type PreviewArtifact } from '../runtime/artifactPreview'
import { formatSafeToolPreview, humanToolName, redactPreviewText } from '../runtime/toolPreview'

type ToolAction = (toolCall: ToolCall) => Promise<void> | void
type PhaseState = 'done' | 'active' | 'pending' | 'failed'
type Props = {
  toolCall: ToolCall
  onApprove?: ToolAction
  onDeny?: ToolAction
  locale?: Locale
  defaultExpanded?: boolean
  onOpenArtifact?: (artifact: PreviewArtifact) => void
}

const copy = {
  zh: {
    input: '请求',
    output: '结果',
    result: '结果',
    waiting: '等待确认',
    approve: '允许',
    deny: '拒绝',
    approving: '正在允许',
    denying: '正在拒绝',
    failed: '操作失败',
    completedSummary: '工具已完成。',
    runningSummary: '工具正在执行。',
    failedSummary: '工具执行失败。',
    deniedSummary: '已拒绝执行。',
    phases: ['请求', '确认', '执行', '结果'],
    sources: '来源',
  },
  en: {
    input: 'Request',
    output: 'Result',
    result: 'Result',
    waiting: 'Awaiting approval',
    approve: 'Approve',
    deny: 'Deny',
    approving: 'Approving',
    denying: 'Denying',
    failed: 'Tool action failed',
    completedSummary: 'Tool completed.',
    runningSummary: 'Tool is running.',
    failedSummary: 'Tool failed.',
    deniedSummary: 'Tool denied.',
    phases: ['Request', 'Approval', 'Execution', 'Result'],
    sources: 'Sources',
  },
}

function ArtifactResourceCard({ artifact, locale, onOpen }: { artifact: PreviewArtifact; locale: Locale; onOpen?: (artifact: PreviewArtifact) => void }) {
  const kind = artifact.kind === 'markdown'
    ? locale === 'zh' ? 'Markdown 文档' : 'Markdown document'
    : locale === 'zh' ? '文档' : 'Document'
  return (
    <button
      type="button"
      className="artifact-resource-card"
      aria-label={locale === 'zh' ? `打开 ${artifact.title}` : `Open ${artifact.title}`}
      onClick={() => onOpen?.(artifact)}
    >
      <span className="artifact-resource-icon"><FileText size={17} /></span>
      <span className="artifact-resource-copy">
        <strong>{artifact.title}</strong>
        <small>{kind}</small>
      </span>
    </button>
  )
}

function statusLabel(status: ToolCall['status'], locale: Locale) {
  if (locale === 'zh') {
    if (status === 'approval_required') return '等待确认'
    if (status === 'completed') return '已完成'
    if (status === 'running') return '执行中'
    if (status === 'succeeded') return '已完成'
    if (status === 'failed') return '失败'
    if (status === 'denied') return '已拒绝'
    if (status === 'executing') return '执行中'
    return status
  }
  if (status === 'approval_required') return 'Awaiting approval'
  if (status === 'succeeded') return 'Completed'
  if (status === 'failed') return 'Failed'
  if (status === 'denied') return 'Denied'
  if (status === 'executing') return 'Running'
  return status
}

function StatusIcon({ status }: { status: ToolCall['status'] }) {
  if (status === 'approval_required') return <ShieldCheck size={14} />
  if (status === 'failed') return <AlertTriangle size={14} />
  if (status === 'denied' || status === 'cancelled') return <XCircle size={14} />
  if (status === 'running' || status === 'executing') return <Clock3 size={14} />
  return <CheckCircle2 size={14} />
}

function ToolIcon({ name }: { name: string }) {
  if (name === 'web.search' || name.startsWith('web.')) return <Search size={14} />
  if (name.startsWith('workspace.grep') || name.startsWith('workspace.glob') || name.startsWith('lsp.')) return <FolderSearch size={14} />
  if (name.startsWith('workspace.') || name.startsWith('artifact.')) return <FileText size={14} />
  return <Terminal size={14} />
}

function phaseStates(toolCall: ToolCall): PhaseState[] {
  if (toolCall.status === 'approval_required') return ['done', 'active', 'pending', 'pending']
  if (toolCall.status === 'denied' || toolCall.status === 'cancelled') return ['done', 'failed', 'pending', 'failed']
  if (toolCall.status === 'failed') return ['done', 'done', 'failed', 'failed']
  if (toolCall.status === 'running' || toolCall.status === 'executing') return ['done', 'done', 'active', 'pending']
  return ['done', 'done', 'done', 'done']
}

function webSearchSources(toolCall: ToolCall) {
  if (toolCall.name !== 'web.search' || !Array.isArray(toolCall.resultSummary?.items)) return []
  return toolCall.resultSummary.items
    .filter((item): item is Record<string, unknown> => typeof item === 'object' && item !== null && !Array.isArray(item))
    .map((item) => {
      const title = typeof item.title === 'string' ? redactPreviewText(item.title).trim() : ''
      const snippet = typeof item.snippet === 'string' ? redactPreviewText(item.snippet).replace(/\s+/g, ' ').trim() : ''
      let host = ''
      if (typeof item.url === 'string') {
        try {
          host = new URL(item.url).hostname.replace(/^www\./, '')
        } catch {
          host = ''
        }
      }
      return { title: title || host, host, snippet }
    })
    .filter((item) => item.title || item.host)
    .slice(0, 3)
}

function workspaceLabel(toolCall: ToolCall) {
  if (!toolCall.name.startsWith('workspace.')) return ''
  const result = toolCall.resultSummary
  if (!result || typeof result !== 'object' || Array.isArray(result)) return ''
  const label = (result as Record<string, unknown>).workspace_label
  return typeof label === 'string' ? redactPreviewText(label).trim() : ''
}

export function ToolCallCard({ toolCall, onApprove, onDeny, locale = 'en', defaultExpanded = false, onOpenArtifact }: Props) {
  const [pendingAction, setPendingAction] = useState<'approve' | 'deny' | null>(null)
  const [actionError, setActionError] = useState<string | null>(null)
  const [expanded, setExpanded] = useState(defaultExpanded)
  const text = copy[locale]
  const argumentsSummary = formatSafeToolPreview(toolCall.argumentsSummary, locale)
  const resultSummary = formatSafeToolPreview(toolCall.resultSummary, locale)
  const inputPreview = argumentsSummary || redactPreviewText(toolCall.input)
  const outputPreview = resultSummary || redactPreviewText(toolCall.output)
  const workspace = workspaceLabel(toolCall)
  const workspacePreview = workspace ? `${locale === 'zh' ? '正在读取' : 'Reading'}：${workspace}` : ''
  const displayName = humanToolName(toolCall.name, locale)
  const approvalRequired = toolCall.status === 'approval_required'
  const hasDetails = Boolean(inputPreview || outputPreview || toolCall.errorMessage)
  const attentionState = approvalRequired || toolCall.status === 'running' || toolCall.status === 'executing' || toolCall.status === 'failed'
  const compactPreview = attentionState
    ? toolCall.status === 'failed' ? outputPreview || inputPreview : inputPreview || outputPreview
    : workspacePreview || inputPreview
  const sources = webSearchSources(toolCall)
  const phases = phaseStates(toolCall)
  const showPhaseStrip = approvalRequired || toolCall.status === 'running' || toolCall.status === 'executing'
  const summary = approvalRequired
    ? text.waiting
    : toolCall.status === 'succeeded'
      ? text.completedSummary
      : toolCall.status === 'failed'
        ? text.failedSummary
        : toolCall.status === 'denied'
          ? text.deniedSummary
          : text.runningSummary
  const terminal = toolCall.status === 'denied' || toolCall.status === 'succeeded' || toolCall.status === 'failed' || toolCall.status === 'cancelled'
  const artifact = terminal ? getToolCallArtifact(toolCall) : null
  const actionsDisabled = !approvalRequired || terminal || pendingAction !== null || !onApprove || !onDeny
  const runAction = async (action: 'approve' | 'deny', handler?: ToolAction) => {
    if (!handler || actionsDisabled) return
    setActionError(null)
    setPendingAction(action)
    try {
      await handler(toolCall)
    } catch (err) {
      setActionError(err instanceof Error ? err.message : text.failed)
    } finally {
      setPendingAction(null)
    }
  }
  if (artifact && toolCall.status !== 'failed') {
    return (
      <div className="tool-artifact-row">
        <ArtifactResourceCard artifact={artifact} locale={locale} onOpen={onOpenArtifact} />
      </div>
    )
  }

  return (
    <div className={`tool-card status-${toolCall.status}${expanded ? ' expanded' : ''}${approvalRequired ? ' needs-approval' : ''}${attentionState ? '' : ' compact'}`}>
      <button className="tool-card-header" type="button" disabled={!hasDetails} aria-expanded={expanded} onClick={() => setExpanded((value) => !value)}>
        <span className="tool-card-title"><ToolIcon name={toolCall.name} /> {displayName}</span>
        {!attentionState && compactPreview && <span className="tool-card-preview">{compactPreview}</span>}
        <span className="tool-card-meta">
          <span className="tool-status-pill">{statusLabel(toolCall.status, locale)}</span>
          {hasDetails && (expanded ? <ChevronDown size={13} /> : <ChevronRight size={13} />)}
        </span>
      </button>
      {attentionState && (
        <div className="tool-summary">
          <StatusIcon status={toolCall.status} />
          {(showPhaseStrip || toolCall.status === 'failed' || toolCall.status === 'denied' || toolCall.status === 'cancelled') && <span>{summary}</span>}
          {compactPreview && <em>{compactPreview}</em>}
        </div>
      )}
      {showPhaseStrip && (
        <div className="tool-phase-strip" aria-label={locale === 'zh' ? '工具阶段' : 'Tool phases'}>
          {text.phases.map((phase, index) => <span className={`tool-phase ${phases[index]}`} key={phase}>{phase}</span>)}
        </div>
      )}
      {expanded && sources.length > 0 && (
        <div className="tool-source-list" aria-label={text.sources}>
          <strong>{text.sources}</strong>
          {sources.map((source, index) => (
            <span className="tool-source-chip" key={`${source.host}-${source.title}-${index}`}>
              <span>{source.title}</span>
              {source.host && <em>{source.host}</em>}
            </span>
          ))}
        </div>
      )}
      {approvalRequired && (
        <div className="tool-actions">
          <Button className="primary" disabled={actionsDisabled} onClick={() => void runAction('approve', onApprove)} type="primary">{pendingAction === 'approve' ? text.approving : text.approve}</Button>
          <Button disabled={actionsDisabled} onClick={() => void runAction('deny', onDeny)}>{pendingAction === 'deny' ? text.denying : text.deny}</Button>
        </div>
      )}
      {expanded && (
        <div className="tool-detail-panel">
          {inputPreview && <div><span>{text.input}</span><p>{inputPreview}</p></div>}
          {outputPreview && <div><span>{text.output}</span><p>{outputPreview}</p></div>}
        </div>
      )}
      {expanded && toolCall.errorMessage && <div className="tool-error">{toolCall.errorCode ? `${toolCall.errorCode}: ` : ''}{redactPreviewText(toolCall.errorMessage)}</div>}
      {actionError && <div className="tool-error" role="alert">{actionError}</div>}
    </div>
  )
}
