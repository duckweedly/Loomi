import { describe, expect, test } from 'bun:test'

describe('SettingsView runtime rows', () => {
  test('renders read-only runtime and provider status rows', async () => {
    const source = await Bun.file(new URL('./SettingsView.tsx', import.meta.url)).text()

    expect(source).toContain('t.dataSourceMode')
    expect(source).toContain('t.backendCapability')
    expect(source).toContain('t.providerCapability')
    expect(source).toContain('t.streamState')
    expect(source).toContain('t.readOnly')
  })

  test('masks provider key entry and does not embed credentials', async () => {
    const source = await Bun.file(new URL('./SettingsView.tsx', import.meta.url)).text()

    expect(source).toContain('type="password"')
    expect(source).toContain('apiKeySet')
    expect(source).not.toMatch(/authorization/i)
    expect(source).not.toMatch(/sk-[a-z0-9]/i)
  })

  test('gives Providers and About their own contract panels', async () => {
    const source = await Bun.file(new URL('./SettingsView.tsx', import.meta.url)).text()

    expect(source).toContain('isProviders')
    expect(source).toContain('isAbout')
    expect(source).toContain('t.providerSummaryTitle')
    expect(source).toContain('t.aboutLocalApp')
  })
})
