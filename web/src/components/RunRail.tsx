import { useState } from 'react'
import { Check, ChevronDown, FileText, Folder, Globe2, Minus } from 'lucide-react'
import type { Run, RuntimeScriptId } from '../domain'
import type { Locale } from '../i18n'
import { getDictionary } from '../i18n'
import type { BackendCapabilityStatus } from '../runtime/backendCapabilityStatus'
import { getBackendCapabilityCopy } from '../runtime/backendCapabilityStatus'
import { groupRuntimeEvents } from '../runtime/runtimeEventGroups'
import { AgentStateMotion } from './AgentStateMotion'

type Props = {
  run: Run | null
  open: boolean
  onOpenArtifact: () => void
  onStopRun?: () => void
  selectedRuntimeScript?: RuntimeScriptId
  capabilityStatus?: BackendCapabilityStatus
  locale?: Locale
  onSelectRuntimeScript?: (scriptId: RuntimeScriptId) => void
}

function getEventClassName(event: Run['events'][number]) {
  if (event.severity === 'error' || event.group === 'error' || event.status === 'failed') return 'progress-row failed'
  if (event.status === 'stopped' || event.status === 'cancelled' || event.severity === 'warning') return 'progress-row warning'
  if (event.status === 'queued' || event.status === 'running' || event.status === 'retrying' || event.status === 'recovering' || event.status === 'stopping') return 'progress-row active'
  return 'progress-row done'
}

function getEventMark(event: Run['events'][number], index: number) {
  if (event.severity === 'error' || event.group === 'error' || event.status === 'failed') return <Minus size={10} />
  if (event.status === 'stopped' || event.status === 'cancelled' || event.severity === 'warning') return <Minus size={10} />
  if (event.status === 'queued' || event.status === 'running' || event.status === 'retrying' || event.status === 'recovering' || event.status === 'stopping') return index + 1
  return <Check size={11} />
}

function isWorkspaceMutationTool(toolName: unknown) {
  return toolName === 'workspace.write_file' || toolName === 'workspace.edit'
}

function getEventDetail(event: Run['events'][number], locale: Locale) {
  const workerCopy = getDictionary(locale).runtime.workerJob
  const usage = event.usage
  const usageParts = usage ? [
    usage.inputTokens !== undefined ? `${usage.inputTokens} in` : null,
    usage.outputTokens !== undefined ? `${usage.outputTokens} out` : null,
    usage.totalTokens !== undefined ? `${usage.totalTokens} total` : null,
  ].filter(Boolean) : []
  const eventLabels: Record<string, string> = {
    job_claimed: workerCopy.jobClaimed,
    'job.claimed': workerCopy.jobClaimed,
    lease_renewed: workerCopy.leaseRenewed,
    'worker.lease_renewed': workerCopy.leaseRenewed,
    job_recovering: workerCopy.jobRecovering,
    'job.recovering': workerCopy.jobRecovering,
    job_retry_scheduled: workerCopy.retryScheduled,
    'job.retry_scheduled': workerCopy.retryScheduled,
    job_attempt_failed: workerCopy.attemptFailed,
    'job.attempt_failed': workerCopy.attemptFailed,
    job_retry_exhausted: workerCopy.retryExhausted,
    'job.retry_exhausted': workerCopy.retryExhausted,
    cancellation: workerCopy.cancellationRequested,
    'run.cancelled': workerCopy.cancellationRequested,
    worker_diagnostics: workerCopy.diagnostics,
  }
  const modelPhase = event.metadata?.model_phase === 'continuation'
    ? 'Continuation model phase'
    : event.metadata?.model_phase === 'initial'
      ? 'Initial model phase'
      : ''
  const loopIndex = typeof event.metadata?.loop_index === 'number' ? event.metadata.loop_index : undefined
  const loopMax = typeof event.metadata?.loop_max === 'number' ? event.metadata.loop_max : undefined
  const loopCopy = loopIndex !== undefined
    ? loopMax !== undefined
      ? `Loop ${loopIndex}/${loopMax}`
      : `Loop ${loopIndex}`
    : ''
  const isWorkspaceTool = event.metadata?.tool_group === 'workspace'
    || (typeof event.metadata?.tool_name === 'string' && event.metadata.tool_name.startsWith('workspace.'))
  const isSandboxTool = event.metadata?.tool_group === 'sandbox'
    || (typeof event.metadata?.tool_name === 'string' && event.metadata.tool_name.startsWith('sandbox.'))
  const isLSPTool = event.metadata?.tool_group === 'lsp'
    || (typeof event.metadata?.tool_name === 'string' && event.metadata.tool_name.startsWith('lsp.'))
  const isWebTool = event.metadata?.tool_group === 'web'
    || (typeof event.metadata?.tool_name === 'string' && event.metadata.tool_name.startsWith('web.'))
  const isBrowserTool = event.metadata?.tool_group === 'browser'
    || (typeof event.metadata?.tool_name === 'string' && event.metadata.tool_name.startsWith('browser.'))
  const isArtifactTool = event.metadata?.tool_group === 'artifact'
    || (typeof event.metadata?.tool_name === 'string' && event.metadata.tool_name.startsWith('artifact.'))
  const isAgentTool = event.metadata?.tool_group === 'agent'
    || (typeof event.metadata?.tool_name === 'string' && event.metadata.tool_name.startsWith('agent.'))
  const workspaceToolLabel = isWorkspaceMutationTool(event.metadata?.tool_name)
    ? 'Workspace mutation tool · high risk · write-capable'
    : 'Workspace tool'
  const sandboxToolLabel = event.metadata?.tool_name === 'sandbox.exec_command'
    ? 'Sandbox exec tool · high risk · exec-capable'
    : 'Sandbox tool'
  const webToolLabel = event.metadata?.tool_name === 'web.fetch'
    ? 'Web fetch tool · medium risk · public HTTP only'
    : 'Web tool'
  const browserToolLabel = isBrowserTool
    ? 'Browser automation tool · medium risk · public HTTP only'
    : ''
  const artifactToolLabel = isArtifactTool
    ? 'Artifact runtime tool · medium risk · non-executable'
    : ''
  const agentToolLabel = isAgentTool
    ? 'Agent coordination tool · medium risk · no autonomous execution'
    : ''
  const detail = modelPhase
    ? `${modelPhase} · ${event.detail}`
    : event.type.startsWith('tool.call.') && isWorkspaceTool
      ? [workspaceToolLabel, loopCopy, event.detail].filter(Boolean).join(' · ')
    : event.type.startsWith('tool.call.') && isSandboxTool
      ? [sandboxToolLabel, loopCopy, event.detail].filter(Boolean).join(' · ')
    : event.type.startsWith('tool.call.') && isLSPTool
      ? ['LSP read-only tool · low risk · workspace-scoped', loopCopy, event.detail].filter(Boolean).join(' · ')
    : event.type.startsWith('tool.call.') && isWebTool
      ? [webToolLabel, loopCopy, event.detail].filter(Boolean).join(' · ')
    : event.type.startsWith('tool.call.') && isBrowserTool
      ? [browserToolLabel, loopCopy, event.detail].filter(Boolean).join(' · ')
    : event.type.startsWith('tool.call.') && isArtifactTool
      ? [artifactToolLabel, loopCopy, event.detail].filter(Boolean).join(' · ')
    : event.type.startsWith('tool.call.') && isAgentTool
      ? [agentToolLabel, loopCopy, event.detail].filter(Boolean).join(' · ')
    : event.type === 'error.provider_error' || event.type === 'error.provider_timeout' || event.type === 'error.provider_rate_limited'
    ? `Provider failure · ${event.detail}`
    : event.type === 'progress.tool_call_blocked'
      ? `Tool request blocked · ${event.detail}`
      : eventLabels[event.type]
        ? `${eventLabels[event.type]} · ${event.detail}`
        : event.type.includes('worker') || event.type.includes('job')
          ? `${workerCopy.unknownWorkerEvent} · ${event.type} · ${event.detail}`
          : event.detail
  return usageParts.length > 0 ? `${detail} · ${usageParts.join(' / ')}` : detail
}

export function RunRail({ run, open, onOpenArtifact, onStopRun, selectedRuntimeScript = 'success', capabilityStatus, locale = 'en', onSelectRuntimeScript }: Props) {
  const [collapsedSections, setCollapsedSections] = useState<Set<string>>(new Set())
  const toggleSection = (section: string) => {
    setCollapsedSections((current) => {
      const next = new Set(current)
      if (next.has(section)) next.delete(section)
      else next.add(section)
      return next
    })
  }

  const eventGroups = groupRuntimeEvents(run?.events ?? [], locale)
  const capabilityCopy = capabilityStatus ? getBackendCapabilityCopy(capabilityStatus, locale) : null

  return (
    <aside className={open ? 'floating-rail open' : 'floating-rail'}>
      <section className={collapsedSections.has('progress') ? 'rail-card progress-card collapsed' : 'rail-card progress-card'}>
        <button className="rail-card-head" onClick={() => toggleSection('progress')}>
          <h2>Progress</h2>
          <ChevronDown size={18} />
        </button>
        <div className="rail-card-body progress-list">
          <AgentStateMotion run={run} compact />
          {onSelectRuntimeScript && (
            <div className="runtime-script-switch compact" aria-label="Mock runtime script">
              <span>Scenario</span>
              <button className={selectedRuntimeScript === 'success' ? 'selected' : undefined} onClick={() => onSelectRuntimeScript('success')}>Success</button>
              <button className={selectedRuntimeScript === 'failure' ? 'selected' : undefined} onClick={() => onSelectRuntimeScript('failure')}>Fail</button>
            </div>
          )}
          {capabilityCopy && <div className={`capability-rail ${capabilityStatus}`}><strong>{capabilityCopy.title}</strong><span>{capabilityCopy.detail}</span></div>}
          {(run?.status === 'queued' || run?.status === 'running' || run?.status === 'retrying' || run?.status === 'recovering' || run?.status === 'blocked_on_tool_approval') && onStopRun && <button className="runtime-stop-button ghost" onClick={onStopRun}>Stop run</button>}
          {eventGroups.map((group) => (
            <section key={group.id} className={`runtime-event-group ${group.id}`}>
              <h3>{group.title}</h3>
              {group.events.length === 0 ? <p className="runtime-event-empty">No events yet</p> : group.events.map((event, index) => (
                <div key={event.id} className={getEventClassName(event)}>
                  <span className="progress-mark">{getEventMark(event, index)}</span>
                  <span>{getEventDetail(event, locale)}</span>
                  <small>{event.time}</small>
                </div>
              ))}
            </section>
          ))}
        </div>
      </section>

      <section className={collapsedSections.has('files') ? 'rail-card files-card collapsed' : 'rail-card files-card'}>
        <button className="rail-card-head" onClick={() => toggleSection('files')}>
          <h2>Loomi</h2>
          <div className="rail-card-actions">
            <Folder size={17} />
            <ChevronDown size={18} />
          </div>
        </button>
        <div className="rail-card-body file-list">
          {['Instructions · CLAUDE.md', 'compose.yaml', 'tasks.md', 'data-model.md', 'spec.md', 'plan.md'].map((file) => (
            <div className="file-row" key={file}>
              <span className="file-icon"><FileText size={16} /></span>
              <span>{file}</span>
            </div>
          ))}
        </div>
      </section>

      <section className={collapsedSections.has('context') ? 'rail-card context-card collapsed' : 'rail-card context-card'}>
        <button className="rail-card-head" onClick={() => toggleSection('context')}>
          <h2>Context</h2>
          <ChevronDown size={18} />
        </button>
        <div className="rail-card-body">
        <span className="rail-card-kicker">Connectors</span>
        <button className="file-row context-row" onClick={onOpenArtifact}>
          <span className="file-icon"><Globe2 size={16} /></span>
          <span>Web search</span>
        </button>
        </div>
      </section>
    </aside>
  )
}
