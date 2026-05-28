import { describe, expect, test } from 'bun:test'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

describe('Loomi menu primitive', () => {
  test('keeps floating menus in a shared portal with unified dismissal', () => {
    const source = readFileSync(resolve(import.meta.dir, 'LoomiMenu.tsx'), 'utf8')

    expect(source).toContain("import { createPortal } from 'react-dom'")
    expect(source).toContain('className={`loomi-floating-menu ${className}`.trim()}')
    expect(source).toContain('createPortal(menu, document.body)')
    expect(source).toContain("target.closest('.loomi-floating-menu')")
    expect(source).toContain('ignoreSelector && target.closest(ignoreSelector)')
    expect(source).toContain("event.key === 'Escape'")
    expect(source).toContain('role="menu"')
    expect(source).toContain('role="menuitem"')
    expect(source).toContain('role="separator"')
  })
})
