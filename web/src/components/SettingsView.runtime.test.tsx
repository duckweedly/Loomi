import { describe, expect, test } from 'bun:test'

describe('SettingsView runtime rows', () => {
  test('keeps runtime diagnostics out of General settings', async () => {
    const source = await Bun.file(new URL('./SettingsView.tsx', import.meta.url)).text()

    expect(source).not.toContain('t.dataSourceMode')
    expect(source).not.toContain('t.mockRuntimeScenario')
    expect(source).toContain('t.theme')
    expect(source).not.toContain('RuntimeStatusRows')
    expect(source).not.toContain('t.streamState')
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

describe('SettingsView memory provider foundation', () => {
  test('renders backend-derived memory provider state and actions', async () => {
    const source = await Bun.file(new URL('./SettingsView.tsx', import.meta.url)).text()

    expect(source).toContain('MemoryProviderFoundationPanel')
    expect(source).toContain('memoryProviderStatus')
    expect(source).toContain('onRefreshMemoryProviderStatus')
    expect(source).toContain('onUpdateMemoryProvider')
    expect(source).toContain('current.diagnostic.message')
    expect(source).toContain('MemorySnapshotPanel')
    expect(source).toContain('memoryOverviewSnapshot')
    expect(source).toContain('memoryImpressionSnapshot')
    expect(source).toContain('onRebuildMemoryOverviewSnapshot')
    expect(source).toContain('onRebuildMemoryImpressionSnapshot')
    expect(source).toContain('onGetMemoryContent')
    expect(source).toContain('memory-content-modal')
    expect(source).toContain('memoryErrors')
    expect(source).toContain('memoryProviderErrorSummary')
    expect(source).toContain('item.eventType')
    expect(source).toContain('item.runId')
    expect(source).toContain('近期异常')
    expect(source).toContain('onDetectNowledgeMemoryProvider')
    expect(source).toContain('onDetectOpenVikingMemoryProvider')
    expect(source).toContain('检测本地实例')
    expect(source).toContain('记忆画像')
    expect(source).toContain('记忆快照')
    expect(source).toContain('待审批提案')
    expect(source).toContain('OpenViking')
    expect(source).toContain('Nowledge')
    expect(source).toContain('rootApiKeySet')
    expect(source).not.toMatch(/Arkloop/i)
    expect(source).not.toMatch(/sk-[a-z0-9]/i)
  })
})
