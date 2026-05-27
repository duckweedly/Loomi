import type { Locale } from '../i18n'
import { getRightPanelItemCopy, rightPanelItems, type RightPanelItemId } from '../rightPanelItems'

type Props = {
  open: boolean
  selectedPanelId: RightPanelItemId
  onSelectPanel: (panelId: RightPanelItemId) => void
  locale?: Locale
}

export function RightPanelMenu({ open, selectedPanelId, onSelectPanel, locale = 'en' }: Props) {
  return (
    <div className={open ? 'right-panel-menu open' : 'right-panel-menu'}>
      {rightPanelItems.map((item) => {
        const Icon = item.Icon
        const itemCopy = getRightPanelItemCopy(item, locale)
        return (
          <button
            className={item.id === selectedPanelId ? 'right-panel-menu-item selected' : 'right-panel-menu-item'}
            key={item.id}
            onClick={() => onSelectPanel(item.id)}
          >
            <span className="right-panel-menu-label">
              <Icon size={15} strokeWidth={1.8} />
              {itemCopy.label}
            </span>
            {item.shortcut && <span className="right-panel-shortcut">{item.shortcut}</span>}
          </button>
        )
      })}
    </div>
  )
}
