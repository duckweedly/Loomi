import { useState } from 'react'
import { Check, ChevronDown, FileText, Folder, Globe2, Minus } from 'lucide-react'
import type { Run, RuntimeScriptId } from '../domain'
import { AgentStateMotion } from './AgentStateMotion'

type Props = {
  run: Run | null
  open: boolean
  onOpenArtifact: () => void
  onStopRun?: () => void
  selectedRuntimeScript?: RuntimeScriptId
  onSelectRuntimeScript?: (scriptId: RuntimeScriptId) => void
}

function getEventClassName(status: Run['events'][number]['status']) {
  if (status === 'running') return 'progress-row active'
  if (status === 'failed' || status === 'stopped') return 'progress-row failed'
  return 'progress-row done'
}

function getEventMark(event: Run['events'][number], index: number) {
  if (event.status === 'running') return index + 1
  if (event.status === 'failed' || event.status === 'stopped') return <Minus size={10} />
  return <Check size={11} />
}

function formatRuntimeEventDetail(event: Run['events'][number]) {
  if (event.type === 'error.provider_error' || event.type === 'error.provider_timeout' || event.type === 'error.provider_rate_limited') return `Provider failure · ${event.detail}`
  if (event.type === 'progress.tool_call_blocked') return `Tool request blocked · ${event.detail}`
  if (event.type === 'message.model_output_delta') return 'Model gateway streaming output'
  if (event.type === 'message.model_output_completed') return 'Model gateway response completed'
  return event.detail
}

export function RunRail({ run, open, onOpenArtifact, onStopRun, selectedRuntimeScript = 'success', onSelectRuntimeScript }: Props) {
  const [collapsedSections, setCollapsedSections] = useState<Set<string>>(new Set())
  const toggleSection = (section: string) => {
    setCollapsedSections((current) => {
      const next = new Set(current)
      if (next.has(section)) next.delete(section)
      else next.add(section)
      return next
    })
  }

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
          {run?.status === 'running' && onStopRun && <button className="runtime-stop-button ghost" onClick={onStopRun}>Stop run</button>}
          {run?.events.map((event, index) => (
            <div key={event.id} className={getEventClassName(event.status)}>
              <span className="progress-mark">{getEventMark(event, index)}</span>
              <span>{formatRuntimeEventDetail(event)}</span>
              <small>{event.time}</small>
            </div>
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
