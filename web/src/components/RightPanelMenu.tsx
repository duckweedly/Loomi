import { rightPanelItems, type RightPanelItemId } from '../rightPanelItems'

type Props = {
  open: boolean
  selectedPanelId: RightPanelItemId
  onSelectPanel: (panelId: RightPanelItemId) => void
}

export function RightPanelMenu({ open, selectedPanelId, onSelectPanel }: Props) {
  return (
    <div className={open ? 'right-panel-menu open' : 'right-panel-menu'}>
      {rightPanelItems.map((item) => {
        const Icon = item.Icon
        return (
          <button
            className={item.id === selectedPanelId ? 'right-panel-menu-item selected' : 'right-panel-menu-item'}
            key={item.id}
            onClick={() => onSelectPanel(item.id)}
          >
            <span className="right-panel-menu-label">
              <Icon size={15} strokeWidth={1.8} />
              {item.label}
            </span>
            {item.shortcut && <span className="right-panel-shortcut">{item.shortcut}</span>}
          </button>
        )
      })}
    </div>
  )
}
