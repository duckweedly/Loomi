import { Button, Tag } from '@lobehub/ui'
import { Box, Braces, FileText, GitBranch, Layers3, X } from 'lucide-react'
import type { Run } from '../domain'

type Props = {
  run: Run | null
  open: boolean
  onClose: () => void
  onOpenArtifact: () => void
}

export function RunRail({ run, open, onClose, onOpenArtifact }: Props) {
  return (
    <aside className={open ? 'floating-rail open' : 'floating-rail'}>
      <div className="floating-head">
        <div>
          <strong>Run</strong>
          <span>{run?.model ?? '-'}</span>
        </div>
        <Button aria-label="Close run details" icon={<X size={14} />} onClick={onClose} size="small" />
      </div>

      <div className="rail-strip">
        <span>Status</span>
        <strong>{run?.status ?? 'idle'}</strong>
        <span>Context</span>
        <strong>{run?.context ?? '-'}</strong>
      </div>

      <div className="floating-section">
        <div className="rail-title">Timeline</div>
        <div className="timeline-list">
          {run?.events.map((event) => (
            <div key={event.id} className={`timeline-item ${event.status}`}>
              <span className="timeline-pin" />
              <div>
                <div className="timeline-head">
                  <strong>{event.label}</strong>
                  <span>{event.time}</span>
                </div>
                <p>{event.detail}</p>
                <code>{event.type}</code>
              </div>
            </div>
          ))}
        </div>
      </div>

      <div className="floating-section">
        <div className="rail-title">Panel</div>
        <div className="panel-tabs">
          <Tag variant="filled"><Layers3 size={12} /> Context</Tag>
          <Tag variant="filled"><FileText size={12} /> Files</Tag>
          <Tag variant="filled"><Braces size={12} /> Events</Tag>
        </div>
        <button className="artifact-preview" onClick={onOpenArtifact}>
          <div className="artifact-icon"><Box size={17} /></div>
          <div>
            <strong>Workspace artifact</strong>
            <span>UI shell · mock</span>
          </div>
        </button>
        <div className="dispatch-preview">
          <GitBranch size={14} /> Dispatch placeholder
        </div>
      </div>
    </aside>
  )
}
