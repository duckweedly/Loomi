import { describe, expect, test } from 'bun:test'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

describe('animal-island-ui integration', () => {
  test('loads the component library stylesheet at the app entry', () => {
    const source = readFileSync(resolve(import.meta.dir, 'main.tsx'), 'utf8')

    expect(source).toContain("import 'animal-island-ui/style'")
  })

  test('imports ac-site-template day night theme token layer', () => {
    const styles = readFileSync(resolve(import.meta.dir, 'styles.css'), 'utf8')
    const theme = readFileSync(resolve(import.meta.dir, 'styles/87-ac-site-theme.css'), 'utf8')

    expect(styles).toContain("@import './styles/87-ac-site-theme.css'")
    expect(styles.indexOf("@import './styles/87-ac-site-theme.css'")).toBeGreaterThan(styles.indexOf("@import './styles/91-commercial-product.css'"))
    expect(theme).toContain('--ac-cream: #f7f3e7')
    expect(theme).toContain('--ac-cream: #252b46')
    expect(theme).toContain(".app-shell[data-theme='dark']")
    expect(theme).toContain('.composer.animal-command-bar')
    expect(theme).toContain('.composer-model-select > div > div')
    expect(theme).toContain('.animal-send-button')
    expect(theme).toContain('.main-titlebar')
    expect(theme).not.toContain('.sidebar-mode-row')
    expect(theme).toContain('.thread-card.selected')
    expect(theme).toContain('.settings-content-header')
    expect(theme).toContain('.provider-filter-tabs button.selected')
    expect(theme).toContain('.thread-state.error')
    expect(theme).toContain('.chat-shell.glass-panel')
    expect(theme).toContain('.composer.glass-panel.animal-command-bar')
    expect(theme).toContain('.composer.glass-panel.animal-command-bar::before')
    expect(theme).toContain('.composer-model-select.disabled > div > div')
    expect(theme).toContain("background: #20263f !important")
    expect(theme).toContain("display: none !important")
  })
})
