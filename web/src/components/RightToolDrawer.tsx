import { Button } from '@lobehub/ui'
import { X } from 'lucide-react'
import { rightPanelItems, type RightPanelItemId } from '../rightPanelItems'

type Props = {
  open: boolean
  selectedPanelId: RightPanelItemId
  onClose: () => void
}

export function RightToolDrawer({ open, selectedPanelId, onClose }: Props) {
  const selectedPanel = rightPanelItems.find((item) => item.id === selectedPanelId) ?? rightPanelItems[0]
  const SelectedIcon = selectedPanel.Icon

  return (
    <aside className={open ? 'right-tool-drawer open' : 'right-tool-drawer'}>
      <div className="artifact-drawer-head">
        <div>
          <strong>{selectedPanel.title}</strong>
          <span>Placeholder</span>
        </div>
        <Button aria-label="Close right panel" icon={<X size={14} />} onClick={onClose} size="small" />
      </div>
      <div className="right-panel-empty">
        <div className="right-panel-empty-icon">
          <SelectedIcon size={24} strokeWidth={1.7} />
        </div>
        <strong>{selectedPanel.title}</strong>
        <p>{selectedPanel.description}</p>
        <span>Coming soon</span>
      </div>
    </aside>
  )
}
