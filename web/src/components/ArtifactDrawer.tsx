import { Button } from '@lobehub/ui'
import { Terminal, X } from 'lucide-react'

type Props = {
  open: boolean
  onClose: () => void
}

export function ArtifactDrawer({ open, onClose }: Props) {
  return (
    <aside className={open ? 'artifact-drawer open' : 'artifact-drawer'}>
      <div className="artifact-drawer-head">
        <div>
          <strong>Workspace artifact</strong>
          <span>Preview</span>
        </div>
        <Button aria-label="Close artifact preview" icon={<X size={14} />} onClick={onClose} size="small" />
      </div>
      <div className="artifact-stage">
        <pre className="artifact-code">{`$ loomi preview artifact

Panel: workspace shell
Mode: mock
Status: ready

The real browser / terminal / artifact runtime lands after M1.`}</pre>
        <div className="artifact-note">
          <Terminal size={15} /> Terminal / browser surface placeholder
        </div>
      </div>
    </aside>
  )
}
