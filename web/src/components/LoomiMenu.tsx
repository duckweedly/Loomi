import { useEffect } from 'react'
import type { CSSProperties, ReactNode } from 'react'
import { createPortal } from 'react-dom'

type FloatingMenuProps = {
  open: boolean
  className?: string
  style?: CSSProperties
  ignoreSelector?: string
  onClose: () => void
  children: ReactNode
}

export function LoomiFloatingMenu({ open, className = '', style, ignoreSelector, onClose, children }: FloatingMenuProps) {
  useEffect(() => {
    if (!open) return

    const closeMenu = (event: PointerEvent) => {
      const target = event.target
      if (!(target instanceof Element)) return
      if (target.closest('.loomi-floating-menu')) return
      if (ignoreSelector && target.closest(ignoreSelector)) return
      onClose()
    }

    const closeMenuWithKeyboard = (event: KeyboardEvent) => {
      if (event.key === 'Escape') onClose()
    }

    document.addEventListener('pointerdown', closeMenu, true)
    document.addEventListener('keydown', closeMenuWithKeyboard)
    return () => {
      document.removeEventListener('pointerdown', closeMenu, true)
      document.removeEventListener('keydown', closeMenuWithKeyboard)
    }
  }, [ignoreSelector, onClose, open])

  if (!open) return null

  const menu = (
    <div className={`loomi-floating-menu ${className}`.trim()} role="menu" style={style}>
      {children}
    </div>
  )

  return typeof document === 'undefined' ? menu : createPortal(menu, document.body)
}

type MenuItemProps = {
  children: ReactNode
  className?: string
  disabled?: boolean
  onClick?: () => void
}

export function LoomiMenuItem({ children, className = '', disabled = false, onClick }: MenuItemProps) {
  return (
    <button aria-disabled={disabled} className={`loomi-menu-item ${className}`.trim()} disabled={disabled} type="button" role="menuitem" onClick={onClick}>
      {children}
    </button>
  )
}

export function LoomiMenuSeparator() {
  return <div className="loomi-menu-separator" role="separator" />
}
