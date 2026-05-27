import { describe, expect, test } from 'bun:test'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

describe('ThreadSidebar layout CSS', () => {
  test('allocates a dedicated grid column for the archive action', () => {
    const css = [
      readFileSync(resolve(import.meta.dir, '../styles.css'), 'utf8'),
      readFileSync(resolve(import.meta.dir, '../styles/10-sidebar.css'), 'utf8'),
      readFileSync(resolve(import.meta.dir, '../styles/90-motion-icon-fixes.css'), 'utf8'),
    ].join('\n')

    expect(css).toContain('grid-template-columns: 10px minmax(0, 1fr) 24px;')
    expect(css).toContain('position: absolute;')
    expect(css).toContain('right: 4px;')
    expect(css).toContain('.thread-row.selected > .thread-action')
    expect(css).toContain('.thread-rename-form')
    expect(css).toContain('.thread-rename-form input')
    expect(css).toContain('.thread-list {\n  display: flex;\n  min-height: 0;\n  flex-direction: column;\n  gap: 3px;\n  overflow: visible;')
    expect(css).toContain('.thread-list-section {\n  display: flex;\n  min-height: 0;\n  flex: 1 1 auto;\n  flex-direction: column;\n  overflow: visible;')
    expect(css).toContain('.thread-row > .thread-action')
    expect(css).toContain(".thread-row > .thread-action[aria-expanded='true']")
    expect(css).toContain('width: 136px;')
    expect(css).toContain('font-size: 12px;')
  })

  test('aligns electron titlebar icons with the native window controls', () => {
    const css = [
      readFileSync(resolve(import.meta.dir, '../styles.css'), 'utf8'),
      readFileSync(resolve(import.meta.dir, '../styles/00-base-shell.css'), 'utf8'),
      readFileSync(resolve(import.meta.dir, '../styles/90-motion-icon-fixes.css'), 'utf8'),
    ].join('\n')

    expect(css).toContain(".app-shell[data-runtime='electron'] .titlebar-button")
    expect(css).toContain('transform: translateY(-3px);')
  })

  test('applies the visible animal island shell to navigation chrome', () => {
    const css = readFileSync(resolve(import.meta.dir, '../styles/87-ac-site-theme.css'), 'utf8')

    expect(css).toContain('.sidebar-shell')
    expect(css).toContain('.main-titlebar')
    expect(css).toContain('.titlebar-button')
    expect(css).not.toContain('.sidebar-mode-row')
    expect(css).toContain('.thread-create-button')
    expect(css).toContain('.sidebar-settings-button')
    expect(css).toContain('.thread-card > span')
    expect(css).toContain('box-shadow: 0 5px 0')
  })
})
