import { describe, expect, test } from 'bun:test'

describe('App settings surface wiring', () => {
  test('opens Settings from the sidebar and renders it in the main workspace region', async () => {
    const source = await Bun.file(new URL('./App.tsx', import.meta.url)).text()

    expect(source).toContain('import { SettingsView }')
    expect(source).toContain('onOpenSettings={shell.openSettings}')
    expect(source).toContain('shell.settingsOpen ?')
    expect(source).toContain('<SettingsView')
  })

  test('passes session locale and provider draft settings into Settings', async () => {
    const source = await Bun.file(new URL('./App.tsx', import.meta.url)).text()

    expect(source).toContain('locale={shell.locale}')
    expect(source).toContain('onSelectLocale={shell.setLocale}')
    expect(source).toContain('providerDraftSettings={shell.providerDraftSettings}')
    expect(source).toContain('onProviderDraftSettingsChange={shell.setProviderDraftSettings}')
    expect(source).toContain('getDictionary(shell.locale)')
  })

  test('keeps a back path from Settings to the workspace', async () => {
    const source = await Bun.file(new URL('./App.tsx', import.meta.url)).text()

    expect(source).toContain('onBack={shell.closeSettings}')
  })
})
