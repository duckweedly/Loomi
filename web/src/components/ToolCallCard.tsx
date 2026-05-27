import { useState } from 'react'
import { Tag } from '@lobehub/ui'
import { AlertTriangle, CheckCircle2, ChevronDown, ChevronRight, Clock3, Search, ShieldCheck, Terminal, XCircle } from 'lucide-react'
import type { ToolCall } from '../domain'
import type { Locale } from '../i18n'
import { formatSafeToolPreview, humanToolName, redactPreviewText } from '../runtime/toolPreview'

type ToolAction = (toolCall: ToolCall) => Promise<void> | void
type PhaseState = 'done' | 'active' | 'pending' | 'failed'

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

export function ToolCallCard({ toolCall, onApprove, onDeny, locale = 'en' }: { toolCall: ToolCall; onApprove?: ToolAction; onDeny?: ToolAction; locale?: Locale }) {
  const [pendingAction, setPendingAction] = useState<'approve' | 'deny' | null>(null)
  const [actionError, setActionError] = useState<string | null>(null)
  const [expanded, setExpanded] = useState(false)
  const text = copy[locale]
  const argumentsSummary = formatSafeToolPreview(toolCall.argumentsSummary, locale)
  const resultSummary = formatSafeToolPreview(toolCall.resultSummary, locale)
  const inputPreview = argumentsSummary || redactPreviewText(toolCall.input)
  const outputPreview = resultSummary || redactPreviewText(toolCall.output)
  const displayName = humanToolName(toolCall.name, locale)
  const approvalRequired = toolCall.status === 'approval_required'
  const hasDetails = Boolean(inputPreview || outputPreview || toolCall.errorMessage)
  const compactPreview = approvalRequired
    ? inputPreview
    : outputPreview || inputPreview
  const sources = webSearchSources(toolCall)
  const phases = phaseStates(toolCall)
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
  return (
    <div className={`tool-card status-${toolCall.status}${expanded ? ' expanded' : ''}${approvalRequired ? ' needs-approval' : ''}`}>
      <button className="tool-card-header" type="button" disabled={!hasDetails} aria-expanded={expanded} onClick={() => setExpanded((value) => !value)}>
        <span>{toolCall.name === 'web.search' ? <Search size={14} /> : <Terminal size={14} />} {displayName}</span>
        <span className="tool-card-meta">
          <Tag variant="filled">{statusLabel(toolCall.status, locale)}</Tag>
          {hasDetails && (expanded ? <ChevronDown size={13} /> : <ChevronRight size={13} />)}
        </span>
      </button>
      <div className="tool-summary">
        <StatusIcon status={toolCall.status} />
        <span>{summary}</span>
        {compactPreview && <em>{compactPreview}</em>}
      </div>
      <div className="tool-phase-strip" aria-label={locale === 'zh' ? '工具阶段' : 'Tool phases'}>
        {text.phases.map((phase, index) => <span className={`tool-phase ${phases[index]}`} key={phase}>{phase}</span>)}
      </div>
      {sources.length > 0 && (
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
          <button className="primary" disabled={actionsDisabled} onClick={() => void runAction('approve', onApprove)}>{pendingAction === 'approve' ? text.approving : text.approve}</button>
          <button disabled={actionsDisabled} onClick={() => void runAction('deny', onDeny)}>{pendingAction === 'deny' ? text.denying : text.deny}</button>
        </div>
      )}
      {expanded && (
        <div className="tool-grid">
          <div><span>{text.input}</span>{inputPreview || '-'}</div>
          <div><span>{text.output}</span>{outputPreview || '-'}</div>
        </div>
      )}
      {expanded && toolCall.errorMessage && <div className="tool-error">{toolCall.errorCode ? `${toolCall.errorCode}: ` : ''}{redactPreviewText(toolCall.errorMessage)}</div>}
      {actionError && <div className="tool-error" role="alert">{actionError}</div>}
    </div>
  )
}
