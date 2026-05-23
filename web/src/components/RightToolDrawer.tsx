import { Terminal } from 'lucide-react'
import { rightPanelItems, type RightPanelItemId } from '../rightPanelItems'

type Props = {
  open: boolean
  selectedPanelId: RightPanelItemId
}

function PreviewPanel() {
  return (
    <div className="right-panel-preview">
      <pre className="artifact-code">{`$ loomi preview artifact

Panel: workspace shell
Mode: mock
Status: ready

Runtime after M1.`}</pre>
      <div className="artifact-note">
        <Terminal size={15} /> Terminal / browser placeholder
      </div>
    </div>
  )
}

export function RightToolDrawer({ open, selectedPanelId }: Props) {
  const selectedPanel = rightPanelItems.find((item) => item.id === selectedPanelId) ?? rightPanelItems[0]
  const SelectedIcon = selectedPanel.Icon
  const isPreview = selectedPanel.id === 'preview'

  return (
    <aside className={open ? 'right-tool-drawer open' : 'right-tool-drawer'}>
      <div className="right-panel-head">
        <div>
          <strong>{selectedPanel.title}</strong>
          <span>{isPreview ? 'Artifact' : 'Placeholder'}</span>
        </div>
      </div>
      {isPreview ? (
        <PreviewPanel />
      ) : (
        <div className="right-panel-empty">
          <div className="right-panel-empty-icon">
            <SelectedIcon size={24} strokeWidth={1.7} />
          </div>
          <strong>{selectedPanel.title}</strong>
          <p>{selectedPanel.description}</p>
          <span>Coming soon</span>
        </div>
      )}
    </aside>
  )
}
