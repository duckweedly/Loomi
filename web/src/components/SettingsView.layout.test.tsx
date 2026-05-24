import { describe, expect, test } from 'bun:test'

const source = Bun.file(new URL('./SettingsView.tsx', import.meta.url)).text()

describe('SettingsView layout contract', () => {
  test('renders the required desktop-style landmarks', async () => {
    const text = await source

    expect(text).toContain('className="settings-shell"')
    expect(text).toContain('className="settings-sidebar"')
    expect(text).toContain('className="settings-content"')
    expect(text).toContain('className="settings-card"')
    expect(text).toContain('t.back')
  })

  test('distinguishes rows with status badges and right-aligned controls', async () => {
    const text = await source

    expect(text).toContain('setting-status-badge')
    expect(text).toContain('setting-row-control')
    expect(text).toContain('t.working')
    expect(text).toContain('t.previewOnly')
  })
})
