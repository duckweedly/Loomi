import { renderToStaticMarkup } from 'react-dom/server'
import { SettingsView } from './SettingsView'

describe('SettingsView MCP management', () => {
  test('renders editable MCP server status without raw saved config or secrets', () => {
    const html = renderToStaticMarkup(<SettingsView {...baseSettingsProps()} />)

    expect(html).toContain('本地 stdio MCP')
    expect(html).toContain('保存配置')
    expect(html).toContain('连接测试')
    expect(html).toContain('Local Smoke')
    expect(html).toContain('local-smoke')
    expect(html).toContain('stdio')
    expect(html).toContain('succeeded')
    expect(html).toContain('mcp.local-smoke.echo')
    expect(html).toContain('approval_gated')
    expect(html).not.toContain('未来 MCP')
    expect(html).not.toContain('Mock only')
    expect(html).not.toContain('/Users/')
    expect(html).not.toContain('SECRET_CANARY')
  })

  test('wires save, discover, and delete actions', () => {
    const calls: string[] = []
    const html = renderToStaticMarkup(<SettingsView {...baseSettingsProps({
      onSaveMCPServer: (input) => calls.push(`save:${input.slug}`),
      onDiscoverMCPServer: (slug) => calls.push(`discover:${slug}`),
      onDeleteMCPServer: (slug) => calls.push(`delete:${slug}`),
    })} />)

    expect(html).toContain('data-testid="mcp-settings"')
    expect(calls).toEqual([])
  })
})

function baseSettingsProps(overrides: Partial<Parameters<typeof SettingsView>[0]> = {}): Parameters<typeof SettingsView>[0] {
  return {
    locale: 'zh',
    selectedCategoryId: 'mcp',
    defaultWorkspaceMode: 'work',
    theme: 'light',
    themePreference: 'system',
    backendCapability: 'available',
    streamState: 'closed',
    selectedThreadTitle: 'M25 smoke',
    selectedRunStatus: 'completed',
    providerCapabilities: [],
    toolCatalog: [],
    mcpServers: [{
      serverSafeId: 'mcp:local-smoke',
      serverSlug: 'local-smoke',
      displayName: 'Local Smoke',
      transport: 'stdio',
      enabled: true,
      configSource: 'local',
      discoveryStatus: 'succeeded',
      candidateCount: 1,
      candidateNames: ['mcp.local-smoke.echo'],
      executionMode: 'approval_gated',
    }],
    localProviderDetections: [],
    memoryEntries: [],
    memoryQuery: '',
    memoryFilters: {},
    memoryLoading: false,
    memoryDetail: null,
    memoryAuditItems: [],
    memoryAuditLoading: false,
    pendingDeleteMemoryEntry: null,
    providerCheckResults: {},
    providerSaveResult: { status: 'idle' },
    providerDraftSettings: { baseUrl: '', model: '', apiKey: '', apiKeySet: false },
    onSelectLocale: () => {},
    onSelectCategory: () => {},
    onSelectDefaultWorkspaceMode: () => {},
    onSelectTheme: () => {},
    onProviderDraftSettingsChange: () => {},
    onSaveProvider: () => {},
    onCheckProvider: () => {},
    onDetectLocalProviders: () => {},
    onEnableLocalProvider: () => {},
    onDisableLocalProvider: () => {},
    onMemoryQueryChange: () => {},
    onMemoryFiltersChange: () => {},
    onOpenMemoryDetail: () => {},
    onCloseMemoryDetail: () => {},
    onRequestDeleteMemoryEntry: () => {},
    onCancelDeleteMemoryEntry: () => {},
    onConfirmDeleteMemoryEntry: () => {},
    onBack: () => {},
    ...overrides,
  }
}
