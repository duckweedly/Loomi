import { useState } from 'react'
import { Tag } from '@lobehub/ui'
import { CheckCircle2, ChevronDown, ChevronRight, Search, Terminal } from 'lucide-react'
import type { ToolCall } from '../domain'
import type { Locale } from '../i18n'
import { formatSafeToolPreview, humanToolName, redactPreviewText } from '../runtime/toolPreview'

type ToolAction = (toolCall: ToolCall) => Promise<void> | void

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
    <div className={`tool-card${expanded ? ' expanded' : ''}${approvalRequired ? ' needs-approval' : ''}`}>
      <button className="tool-card-header" type="button" disabled={!hasDetails} aria-expanded={expanded} onClick={() => setExpanded((value) => !value)}>
        <span>{toolCall.name === 'web.search' ? <Search size={14} /> : <Terminal size={14} />} {displayName}</span>
        <span className="tool-card-meta">
          <Tag variant="filled">{statusLabel(toolCall.status, locale)}</Tag>
          {hasDetails && (expanded ? <ChevronDown size={13} /> : <ChevronRight size={13} />)}
        </span>
      </button>
      <div className="tool-summary">
        <CheckCircle2 size={14} />
        <span>{summary}</span>
        {compactPreview && <em>{compactPreview}</em>}
      </div>
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
