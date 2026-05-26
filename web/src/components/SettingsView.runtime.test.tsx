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

  test('gives Tools a read-only catalog panel instead of the placeholder', async () => {
    const source = await Bun.file(new URL('./SettingsView.tsx', import.meta.url)).text()

    expect(source).toContain('isTools')
    expect(source).toContain('ToolCatalogPanel')
    expect(source).toContain('toolCatalog')
    expect(source).toContain('tool.name')
    expect(source).toContain('riskLevel')
    expect(source).toContain('sideEffect')
    expect(source).toContain('!isTools && !isAbout')
  })
})

describe('SettingsView provider test console', () => {
  test('renders configured providers, test action, and check states', async () => {
    const source = await Bun.file(new URL('./SettingsView.tsx', import.meta.url)).text()

    expect(source).toContain('ProviderCheckConsole')
    expect(source).toContain('t.providerConfiguredProviders')
    expect(source).toContain('t.providerTestConnection')
    expect(source).toContain('t.providerChecking')
    expect(source).toContain('providerCheckResults[provider.id]')
    expect(source).toContain('disabled={checking}')
    expect(source).toContain('onCheckProvider(provider.id)')
  })

  test('labels provider draft as local and separate from real provider calls', async () => {
    const source = await Bun.file(new URL('./SettingsView.tsx', import.meta.url)).text()

    expect(source).toContain('t.providerLocalDraftTitle')
    expect(source).toContain('t.providerLocalDraftDescription')
    expect(source).toContain('providerDraftSettings.baseUrl')
    expect(source).toContain('providerDraftSettings.model')
    expect(source).toContain('apiKeySet')
  })
})
