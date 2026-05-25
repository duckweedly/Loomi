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
    expect(source).toContain('t.providerConsoleTitle')
    expect(source).toContain('t.providerLocalDraftTitle')
    expect(source).toContain('t.aboutLocalApp')
  })
})

describe('SettingsView provider test console', () => {
  test('renders configured providers, test action, and check states', async () => {
    const source = await Bun.file(new URL('./SettingsView.tsx', import.meta.url)).text()
    const copy = await Bun.file(new URL('../i18n.ts', import.meta.url)).text()

    expect(source).toContain('ProviderCheckConsole')
    expect(source).toContain('t.providerConfiguredProviders')
    expect(source).toContain('t.providerTestConnection')
    expect(source).toContain('t.providerChecking')
    expect(source).toContain('providerCheckResults[provider.id]')
    expect(source).toContain('disabled={checking}')
    expect(source).toContain('onCheckProvider(provider.id)')
    expect(source).toContain('provider.baseUrl')
    expect(copy).toContain("providerConsoleTitle: 'Provider Test Console'")
  })

  test('labels provider draft as local and separate from real provider calls', async () => {
    const source = await Bun.file(new URL('./SettingsView.tsx', import.meta.url)).text()

    expect(source).toContain('t.providerLocalDraftTitle')
    expect(source).toContain('t.providerLocalDraftDescription')
    expect(source).toContain('providerDraftSettings.baseUrl')
    expect(source).toContain('providerDraftSettings.model')
    expect(source).toContain('apiKeySet')
  })

  test('renders local provider autodetect opt-in status without secrets', async () => {
    const source = await Bun.file(new URL('./SettingsView.tsx', import.meta.url)).text()
    const copy = await Bun.file(new URL('../i18n.ts', import.meta.url)).text()

    expect(source).toContain('LocalProviderDetectionList')
    expect(source).toContain('localProviderDetections')
    expect(source).toContain('onDetectLocalProviders')
    expect(source).toContain('t.localProviderDetectAction')
    expect(source).toContain('t.localProviderDetectionIdle')
    expect(source).toContain('t.localProviderAutodetectTitle')
    expect(source).toContain('t.localProviderExplicitOptIn')
    expect(source).toContain('t.localProviderNoSecrets')
    expect(source).toContain('onEnableLocalProvider(provider.providerId)')
    expect(source).toContain('onDisableLocalProvider(provider.providerId)')
    expect(source).toContain('t.localProviderEnableForSession')
    expect(source).toContain('t.localProviderDisableForSession')
    expect(source).toContain('t.localProviderExecutionUnsupported')
    expect(source).toContain('t.localProviderExecutionSupported')
    expect(source).not.toContain('access_token')
    expect(source).not.toContain('refresh_token')
    expect(source).not.toContain('sk-')
    expect(copy).toContain("localProviderAutodetectTitle: 'Local provider autodetect'")
    expect(copy).toContain("localProviderExecutionUnsupported: 'Local Codex is enabled but execution is not supported yet'")
    expect(copy).toContain("localProviderExecutionSupported: 'execution supported'")
  })
})
