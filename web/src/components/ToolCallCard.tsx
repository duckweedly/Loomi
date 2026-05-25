import { useState } from 'react'
import { Tag } from '@lobehub/ui'
import { CheckCircle2, Terminal } from 'lucide-react'
import type { ToolCall } from '../domain'

function formatSummary(value: Record<string, unknown> | null | undefined) {
  if (!value) return ''
  return Object.entries(value).map(([key, item]) => `${key}: ${String(item)}`).join(' · ')
}

type ToolAction = (toolCall: ToolCall) => Promise<void> | void

export function ToolCallCard({ toolCall, onApprove, onDeny }: { toolCall: ToolCall; onApprove?: ToolAction; onDeny?: ToolAction }) {
  const [pendingAction, setPendingAction] = useState<'approve' | 'deny' | null>(null)
  const [actionError, setActionError] = useState<string | null>(null)
  const argumentsSummary = formatSummary(toolCall.argumentsSummary)
  const approvalRequired = toolCall.status === 'approval_required'
  const terminal = toolCall.status === 'denied' || toolCall.status === 'succeeded' || toolCall.status === 'failed' || toolCall.status === 'cancelled'
  const actionsDisabled = !approvalRequired || terminal || pendingAction !== null || !onApprove || !onDeny
  const runAction = async (action: 'approve' | 'deny', handler?: ToolAction) => {
    if (!handler || actionsDisabled) return
    setActionError(null)
    setPendingAction(action)
    try {
      await handler(toolCall)
    } catch (err) {
      setActionError(err instanceof Error ? err.message : 'Tool action failed')
    } finally {
      setPendingAction(null)
    }
  }
  return (
    <div className="tool-card">
      <div className="tool-card-header">
        <span><Terminal size={14} /> {toolCall.name}</span>
        <Tag variant="filled">{toolCall.status}</Tag>
      </div>
      <div className="tool-summary"><CheckCircle2 size={14} /> {toolCall.summary}</div>
      <div className="tool-grid">
        <div><span>Input</span>{argumentsSummary || toolCall.input}</div>
        <div><span>Output</span>{toolCall.output}</div>
      </div>
      {toolCall.resultSummary && <div className="tool-summary"><span>Result</span>{formatSummary(toolCall.resultSummary)}</div>}
      {toolCall.errorMessage && <div className="tool-error">{toolCall.errorCode ? `${toolCall.errorCode}: ` : ''}{toolCall.errorMessage}</div>}
      {approvalRequired && (
        <div className="tool-actions">
          <button disabled={actionsDisabled} onClick={() => void runAction('approve', onApprove)}>{pendingAction === 'approve' ? 'Approving' : 'Approve'}</button>
          <button disabled={actionsDisabled} onClick={() => void runAction('deny', onDeny)}>{pendingAction === 'deny' ? 'Denying' : 'Deny'}</button>
        </div>
      )}
      {actionError && <div className="tool-error" role="alert">{actionError}</div>}
    </div>
  )
}
