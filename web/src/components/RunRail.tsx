import { useState } from 'react'
import { Check, ChevronDown, FileText, Folder, Globe2, Minus } from 'lucide-react'
import type { Run, RuntimeScriptId } from '../domain'
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

function getEventDetail(event: Run['events'][number]) {
  const usage = event.usage
  const usageParts = usage ? [
    usage.inputTokens !== undefined ? `${usage.inputTokens} in` : null,
    usage.outputTokens !== undefined ? `${usage.outputTokens} out` : null,
    usage.totalTokens !== undefined ? `${usage.totalTokens} total` : null,
  ].filter(Boolean) : []
  const detail = event.type === 'error.provider_error' || event.type === 'error.provider_timeout' || event.type === 'error.provider_rate_limited'
    ? `Provider failure · ${event.detail}`
    : event.type === 'progress.tool_call_blocked'
      ? `Tool request blocked · ${event.detail}`
      : event.detail
  return usageParts.length > 0 ? `${detail} · ${usageParts.join(' / ')}` : detail
}

export function RunRail({ run, open, onOpenArtifact, onStopRun, selectedRuntimeScript = 'success', capabilityStatus, onSelectRuntimeScript }: Props) {
  const [collapsedSections, setCollapsedSections] = useState<Set<string>>(new Set())
  const toggleSection = (section: string) => {
    setCollapsedSections((current) => {
      const next = new Set(current)
      if (next.has(section)) next.delete(section)
      else next.add(section)
      return next
    })
  }

  const eventGroups = groupRuntimeEvents(run?.events ?? [])
  const capabilityCopy = capabilityStatus ? getBackendCapabilityCopy(capabilityStatus) : null

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
          {(run?.status === 'queued' || run?.status === 'running' || run?.status === 'retrying' || run?.status === 'recovering') && onStopRun && <button className="runtime-stop-button ghost" onClick={onStopRun}>Stop run</button>}
          {eventGroups.map((group) => (
            <section key={group.id} className={`runtime-event-group ${group.id}`}>
              <h3>{group.title}</h3>
              {group.events.length === 0 ? <p className="runtime-event-empty">No events yet</p> : group.events.map((event, index) => (
                <div key={event.id} className={getEventClassName(event)}>
                  <span className="progress-mark">{getEventMark(event, index)}</span>
                  <span>{getEventDetail(event)}</span>
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
