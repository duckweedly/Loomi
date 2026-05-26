import { Tag } from '@lobehub/ui'
import { AlertTriangle, CheckCircle2, CircleStop, Loader2, Terminal } from 'lucide-react'
import type { ToolCall } from '../domain'

function formatSummary(value: Record<string, unknown> | null | undefined) {
  if (!value) return ''
  return Object.entries(value).map(([key, item]) => `${key}: ${formatSummaryValue(item)}`).join(' · ')
}

function formatSummaryValue(value: unknown): string {
  if (Array.isArray(value)) {
    return value.map((item) => {
      if (typeof item === 'object' && item !== null && 'path' in item && 'line' in item) {
        const record = item as { path?: unknown; line?: unknown; preview?: unknown }
        return `${String(record.path)}:${String(record.line)} ${String(record.preview ?? '')}`.trim()
      }
      return formatSummaryValue(item)
    }).join(', ')
  }
  if (typeof value === 'object' && value !== null) {
    return Object.entries(value as Record<string, unknown>).map(([key, item]) => `${key}: ${formatSummaryValue(item)}`).join(', ')
  }
  return String(value)
}

type ToolCallCardProps = {
  toolCall: ToolCall
  onApprove?: (toolCall: ToolCall) => void
  onDeny?: (toolCall: ToolCall) => void
}

function statusLabel(status: ToolCall['status']) {
  if (status === 'approved') return status
  return status.split('_').map((part) => `${part[0]?.toUpperCase() ?? ''}${part.slice(1)}`).join(' ')
}

export function ToolCallCard({ toolCall, onApprove, onDeny }: ToolCallCardProps) {
	const argumentsSummary = formatSummary(toolCall.argumentsSummary)
	const resultSummary = formatSummary(toolCall.resultSummary)
	const approvalRequired = toolCall.status === 'approval_required'
	const canDecide = approvalRequired && Boolean(onApprove) && Boolean(onDeny)
	const status = statusLabel(toolCall.status)
	return (
		<div className="tool-card">
      <div className="tool-card-header">
        <span><Terminal size={14} /> {toolCall.name}</span>
        <Tag variant="filled">{status}</Tag>
      </div>
      <div className="tool-summary"><CheckCircle2 size={14} /> {toolCall.summary}</div>
      <div className="tool-grid">
        <div><span>Input</span>{argumentsSummary || toolCall.input}</div>
        <div><span>Output</span>{resultSummary || toolCall.output}</div>
      </div>
      {toolCall.status === 'executing' && <div className="tool-summary"><Loader2 size={14} /> Executing</div>}
      {toolCall.status === 'succeeded' && <div className="tool-summary"><CheckCircle2 size={14} /> Result {resultSummary || toolCall.output}</div>}
      {toolCall.status === 'failed' && <div className="tool-summary"><AlertTriangle size={14} /> Redacted error {toolCall.errorCode}{toolCall.errorMessage ? ` · ${toolCall.errorMessage}` : ''}</div>}
      {toolCall.status === 'denied' && <div className="tool-summary"><CircleStop size={14} /> Denied</div>}
      {toolCall.status === 'cancelled' && <div className="tool-summary"><CircleStop size={14} /> Cancelled</div>}
			{approvalRequired && (
				<div className="tool-actions">
					<button disabled={!canDecide} onClick={() => onApprove?.(toolCall)}>Approve</button>
					<button disabled={!canDecide} onClick={() => onDeny?.(toolCall)}>Deny</button>
				</div>
			)}
		</div>
	)
}
