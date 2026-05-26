import { describe, expect, test } from 'bun:test'

describe('SettingsView runtime rows', () => {
  test('renders read-only runtime and provider status rows', async () => {
    const source = await Bun.file(new URL('./SettingsView.tsx', import.meta.url)).text()

    expect(source).not.toContain('t.dataSourceMode')
    expect(source).not.toContain('t.mockRuntimeScenario')
    expect(source).toContain('t.theme')
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
    expect(source).toContain('ProviderManagementPanel')
    expect(source).toContain('t.providerAdd')
    expect(source).toContain('t.aboutLocalApp')
  })
})

describe('SettingsView provider test console', () => {
  test('renders configured providers, test action, and check states', async () => {
    const source = await Bun.file(new URL('./SettingsView.tsx', import.meta.url)).text()
    const copy = await Bun.file(new URL('../i18n.ts', import.meta.url)).text()

    expect(source).toContain('provider-card-grid')
    expect(source).toContain('t.providerSearchPlaceholder')
    expect(source).toContain('t.providerFilterAll')
    expect(source).toContain('t.providerFilterEnabled')
    expect(source).toContain('t.providerTestConnection')
    expect(source).toContain('t.providerChecking')
    expect(source).toContain('providerCheckResults')
    expect(source).toContain("disabled={result?.status === 'checking'}")
    expect(source).toContain('onCheckProvider(provider.id)')
    expect(source).toContain('provider.baseUrl')
    expect(copy).toContain("providerAdd: 'Add provider'")
  })

  test('labels provider draft as local and separate from real provider calls', async () => {
    const source = await Bun.file(new URL('./SettingsView.tsx', import.meta.url)).text()

    expect(source).toContain('provider-modal')
    expect(source).toContain('t.providerName')
    expect(source).toContain('t.providerType')
    expect(source).toContain('providerDraftSettings.baseUrl')
    expect(source).toContain('providerDraftSettings.model')
    expect(source).toContain('apiKeySet')
  })

  test('renders local provider autodetect opt-in status without secrets', async () => {
    const source = await Bun.file(new URL('./SettingsView.tsx', import.meta.url)).text()
    const copy = await Bun.file(new URL('../i18n.ts', import.meta.url)).text()

    expect(source).toContain('localProviderDetections')
    expect(source).toContain('onDetectLocalProviders')
    expect(source).toContain('onEnableLocalProvider(provider.id)')
    expect(source).toContain('onDisableLocalProvider(provider.id)')
    expect(source).toContain('t.localProviderDetectAction')
    expect(source).toContain('t.localProviderEnableForSession')
    expect(source).toContain('t.localProviderDisableForSession')
    expect(source).toContain('provider.localProvider')
    expect(source).toContain('provider.readOnly')
    expect(source).not.toContain('access_token')
    expect(source).not.toContain('refresh_token')
    expect(copy).toContain("providerFilterLocal: 'Local'")
    expect(copy).toContain("providerFilterCloud: 'Cloud'")
  })
})
