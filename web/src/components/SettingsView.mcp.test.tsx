import { renderToStaticMarkup } from 'react-dom/server'
import { SettingsView } from './SettingsView'

describe('SettingsView MCP management', () => {
  test('renders read-only MCP server status without raw config or secrets', () => {
    const html = renderToStaticMarkup(<SettingsView {...baseSettingsProps()} />)

    expect(html).toContain('Local Smoke')
    expect(html).toContain('local-smoke')
    expect(html).toContain('stdio')
    expect(html).toContain('succeeded')
    expect(html).toContain('mcp.local-smoke.echo')
    expect(html).toContain('approval_gated')
    expect(html).not.toContain('/Users/')
    expect(html).not.toContain('SECRET_CANARY')
    expect(html).not.toContain('command')
    expect(html).not.toContain('env')
  })
})

function baseSettingsProps(): Parameters<typeof SettingsView>[0] {
  return {
    locale: 'zh',
    selectedCategoryId: 'mcp',
    defaultWorkspaceMode: 'work',
    selectedRuntimeScript: 'success',
    dataSourceMode: 'mock',
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
    onSelectRuntimeScript: () => {},
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
  }
}
