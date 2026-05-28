import { describe, expect, test } from 'bun:test'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

describe('animal-island-ui integration', () => {
  test('loads the component library stylesheet at the app entry', () => {
    const source = readFileSync(resolve(import.meta.dir, 'main.tsx'), 'utf8')

    expect(source).toContain("import 'animal-island-ui/style'")
  })

  test('imports the unified workspace theme as the terminal visual layer', () => {
    const styles = readFileSync(resolve(import.meta.dir, 'styles.css'), 'utf8')
    const theme = readFileSync(resolve(import.meta.dir, 'styles/92-unified-workspace.css'), 'utf8')

    expect(styles).toContain("@import './styles/92-unified-workspace.css'")
    expect(styles).not.toContain("@import './styles/82-brand-refinement.css'")
    expect(styles).not.toContain("@import './styles/83-dark-compact.css'")
    expect(styles).not.toContain("@import './styles/84-pastel-compact.css'")
    expect(styles).not.toContain("@import './styles/86-pastel-green-finish.css'")
    expect(styles).not.toContain("@import './styles/87-ac-site-theme.css'")
    expect(styles).not.toContain("@import './styles/91-commercial-product.css'")
    expect(theme).toContain('--loomi-bg: #ffffff')
    expect(theme).toContain('--loomi-bg: #181818')
    expect(theme).toContain(".app-shell[data-theme='dark']")
    expect(theme).toContain('.composer.animal-command-bar')
    expect(theme).toContain('.composer-model-select > div > div')
    expect(theme).toContain('.animal-send-button')
    expect(theme).toContain('.main-titlebar')
    expect(theme).not.toContain('.sidebar-mode-row')
    expect(theme).toContain('.thread-card.selected')
    expect(theme).toContain('.settings-content-header')
    expect(theme).toContain('.settings-card')
    expect(theme).toContain('.thread-state.error')
    expect(theme).toContain('.chat-shell.glass-panel')
    expect(theme).toContain('--composer-occlusion-height')
    expect(theme).toContain('.chat-shell[data-chat-state]::after')
    expect(theme).toContain('.composer.glass-panel.animal-command-bar')
    expect(theme).toContain('.composer.animal-command-bar::before')
    expect(theme).toContain('.tool-card.needs-approval')
    expect(theme).toContain('max-width: min(560px, 100%) !important')
    expect(theme).toContain('.tool-stack-toggle small')
    expect(theme).toContain(".app-shell[data-theme='dark'] .thread-card.selected")
    expect(theme).toContain(".app-shell[data-theme='dark'] .message-avatar")
    expect(theme).toContain(".app-shell[data-theme='dark'] .composer.glass-panel.animal-command-bar")
    expect(theme).toContain("background: var(--loomi-panel) !important")
  })
})
