import { Tag } from '@lobehub/ui'
import { CheckCircle2, Terminal } from 'lucide-react'
import type { ToolCall } from '../domain'

function formatSummary(value: Record<string, unknown> | null | undefined) {
  if (!value) return ''
  return Object.entries(value).map(([key, item]) => `${key}: ${String(item)}`).join(' · ')
}

export function ToolCallCard({ toolCall }: { toolCall: ToolCall }) {
  const argumentsSummary = formatSummary(toolCall.argumentsSummary)
  const approvalRequired = toolCall.status === 'approval_required'
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
      {approvalRequired && <div className="tool-actions"><button disabled>Approve</button><button disabled>Deny</button></div>}
    </div>
  )
}
