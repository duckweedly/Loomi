import { renderToStaticMarkup } from 'react-dom/server'
import { SettingsView } from './SettingsView'
import type { ToolCatalogItem } from '../domain'

describe('SettingsView tools catalog', () => {
  test('mock catalog includes every M21 workspace read tool', async () => {
    const api = await import('../mockApiClient')
    const tools = await api.mockApiClient.listToolCatalog()
    const workspaceTools = tools.filter((tool) => tool.name === 'workspace.glob' || tool.name === 'workspace.grep' || tool.name === 'workspace.read')

    expect(workspaceTools.map((tool) => tool.name).sort()).toEqual(['workspace.glob', 'workspace.grep', 'workspace.read'])
    for (const tool of workspaceTools) {
      expect(tool.safeMetadata?.read_only).toBe(true)
      expect(tool.safeMetadata?.scope).toBe('workspace')
      expect(`${tool.displayName} ${tool.description} ${JSON.stringify(tool.safeMetadata)}`).not.toContain('/Users/')
    }
  })

  test('mock catalog includes M23 workspace mutation tools as write-capable high-risk entries', async () => {
    const api = await import('../mockApiClient')
    const tools = await api.mockApiClient.listToolCatalog()
    const mutationTools = tools.filter((tool) => tool.name === 'workspace.write_file' || tool.name === 'workspace.edit' || tool.name === 'workspace.patch_preview' || tool.name === 'workspace.patch_apply')

    expect(mutationTools.map((tool) => tool.name).sort()).toEqual(['workspace.edit', 'workspace.patch_apply', 'workspace.patch_preview', 'workspace.write_file'])
    for (const tool of mutationTools) {
      expect(tool.group).toBe('workspace')
      expect(tool.riskLevel).toBe('high')
      expect(tool.approvalPolicy).toBe('always_required')
      expect(tool.safeMetadata?.read_only).toBe(tool.name === 'workspace.patch_preview')
      expect(tool.safeMetadata?.write_capable).toBe(tool.name !== 'workspace.patch_preview')
      expect(tool.safeMetadata?.scope).toBe('workspace')
      if (tool.name === 'workspace.edit') {
        expect(tool.safeMetadata?.requires_read_before_edit).toBe(true)
        expect(tool.safeMetadata?.returns_diff).toBe(true)
        expect(tool.safeMetadata?.normalizes_line_endings).toBe(true)
        expect(tool.safeMetadata?.preserves_indentation).toBe(true)
        expect(tool.safeMetadata?.strips_trailing_whitespace_except_markdown).toBe(true)
      }
      if (tool.name === 'workspace.patch_preview') {
        expect(tool.safeMetadata?.requires_read_before_preview).toBe(true)
        expect(tool.safeMetadata?.preview_only).toBe(true)
        expect(tool.safeMetadata?.returns_diff).toBe(true)
      }
      if (tool.name === 'workspace.patch_apply') {
        expect(tool.safeMetadata?.requires_patch_preview).toBe(true)
        expect(tool.safeMetadata?.returns_diff).toBe(true)
      }
      expect(`${tool.displayName} ${tool.description} ${JSON.stringify(tool.safeMetadata)}`).not.toContain('/Users/')
    }
  })

  test('mock catalog includes M24 sandbox exec command as exec-capable high-risk entry', async () => {
    const api = await import('../mockApiClient')
    const tools = await api.mockApiClient.listToolCatalog()
    const tool = tools.find((item) => item.name === 'sandbox.exec_command')

    expect(tool).toBeDefined()
    expect(tool?.group).toBe('sandbox')
    expect(tool?.riskLevel).toBe('high')
    expect(tool?.approvalPolicy).toBe('always_required')
    expect(tool?.safeMetadata?.scope).toBe('bounded_command')
    expect(tool?.safeMetadata?.exec_capable).toBe(true)
    expect(tool?.safeMetadata?.validation_capable).toBe(true)
    expect(tool?.safeMetadata?.isolated_sandbox).toBe(false)
    expect(tool?.safeMetadata?.argv_only).toBe(true)
    expect(`${tool?.displayName} ${tool?.description} ${JSON.stringify(tool?.safeMetadata)}`).not.toContain('/Users/')
  })

  test('mock catalog includes M25 LSP read-only tools', async () => {
    const api = await import('../mockApiClient')
    const tools = await api.mockApiClient.listToolCatalog()
    const lspTools = tools.filter((tool) => tool.group === 'lsp')

    expect(lspTools.map((tool) => tool.name).sort()).toEqual(['lsp.diagnostics', 'lsp.references', 'lsp.symbols'])
    for (const tool of lspTools) {
      expect(tool.source).toBe('builtin')
      expect(tool.riskLevel).toBe('low')
      expect(tool.approvalPolicy).toBe('always_required')
      expect(tool.executionState).toBe('executable')
      expect(tool.safeMetadata?.scope).toBe('lsp')
      expect(tool.safeMetadata?.read_only).toBe(true)
      expect(`${tool.displayName} ${tool.description} ${JSON.stringify(tool.safeMetadata)}`).not.toContain('/Users/')
    }
  })

  test('mock catalog includes M26 web fetch as public read-only network tool', async () => {
    const api = await import('../mockApiClient')
    const tools = await api.mockApiClient.listToolCatalog()
    const tool = tools.find((item) => item.name === 'web.fetch')

    expect(tool).toBeDefined()
    expect(tool?.source).toBe('builtin')
    expect(tool?.group).toBe('web')
    expect(tool?.riskLevel).toBe('medium')
    expect(tool?.approvalPolicy).toBe('read_only')
    expect(tool?.executionState).toBe('executable')
    expect(tool?.safeMetadata?.scope).toBe('web')
    expect(tool?.safeMetadata?.read_only).toBe(true)
    expect(tool?.safeMetadata?.network_access).toBe('public_http_only')
    expect(`${tool?.displayName} ${tool?.description} ${JSON.stringify(tool?.safeMetadata)}`).not.toContain('/Users/')
  })

  test('mock catalog includes web search as Brave/Tavily read-only network tool', async () => {
    const api = await import('../mockApiClient')
    const tools = await api.mockApiClient.listToolCatalog()
    const tool = tools.find((item) => item.name === 'web.search')

    expect(tool).toBeDefined()
    expect(tool?.source).toBe('builtin')
    expect(tool?.group).toBe('web')
    expect(tool?.riskLevel).toBe('medium')
    expect(tool?.approvalPolicy).toBe('read_only')
    expect(tool?.executionState).toBe('executable')
    expect(tool?.safeMetadata?.scope).toBe('web')
    expect(tool?.safeMetadata?.read_only).toBe(true)
    expect(tool?.safeMetadata?.network_access).toBe('search_provider_api')
    expect(tool?.safeMetadata?.providers).toEqual(['tavily', 'brave'])
    expect(`${tool?.displayName} ${tool?.description} ${JSON.stringify(tool?.safeMetadata)}`).not.toContain('tvly-')
    expect(`${tool?.displayName} ${tool?.description} ${JSON.stringify(tool?.safeMetadata)}`).not.toContain('/Users/')
  })

  test('mock catalog includes M27 browser automation as public stateful network tools', async () => {
    const api = await import('../mockApiClient')
    const tools = await api.mockApiClient.listToolCatalog()
    const browserTools = tools.filter((tool) => tool.group === 'browser')

    expect(browserTools.map((tool) => tool.name).sort()).toEqual(['browser.click_link', 'browser.open', 'browser.snapshot'])
    for (const tool of browserTools) {
      expect(tool.source).toBe('builtin')
      expect(tool.riskLevel).toBe('medium')
      expect(tool.approvalPolicy).toBe('always_required')
      expect(tool.executionState).toBe('executable')
      expect(tool.safeMetadata?.scope).toBe('browser')
      expect(tool.safeMetadata?.network_access).toBe('public_http_only')
      expect(tool.safeMetadata?.stateful_session).toBe(true)
      expect(`${tool.displayName} ${tool.description} ${JSON.stringify(tool.safeMetadata)}`).not.toContain('/Users/')
      expect(`${tool.displayName} ${tool.description} ${JSON.stringify(tool.safeMetadata)}`).not.toContain('cookie')
    }
  })

  test('mock catalog includes M28 artifact runtime as non-executable tools', async () => {
    const api = await import('../mockApiClient')
    const tools = await api.mockApiClient.listToolCatalog()
    const artifactTools = tools.filter((tool) => tool.group === 'artifact')

    expect(artifactTools.map((tool) => tool.name).sort()).toEqual(['artifact.create_text', 'artifact.list', 'artifact.read'])
    for (const tool of artifactTools) {
      expect(tool.source).toBe('builtin')
      expect(tool.riskLevel).toBe('medium')
      expect(tool.approvalPolicy).toBe('always_required')
      expect(tool.executionState).toBe('executable')
      expect(tool.safeMetadata?.scope).toBe('artifact')
      expect(tool.safeMetadata?.non_executable).toBe(true)
      expect(`${tool.displayName} ${tool.description} ${JSON.stringify(tool.safeMetadata)}`).not.toContain('/Users/')
      expect(`${tool.displayName} ${tool.description} ${JSON.stringify(tool.safeMetadata)}`).not.toContain('raw_result')
    }
  })

  test('mock catalog includes M29 agent coordination tools with no autonomous execution', async () => {
    const api = await import('../mockApiClient')
    const tools = await api.mockApiClient.listToolCatalog()
    const agentTools = tools.filter((tool) => tool.group === 'agent')

    expect(agentTools.map((tool) => tool.name).sort()).toEqual(['agent.complete', 'agent.list', 'agent.spawn'])
    for (const tool of agentTools) {
      expect(tool.source).toBe('builtin')
      expect(tool.riskLevel).toBe('medium')
      expect(tool.approvalPolicy).toBe('always_required')
      expect(tool.executionState).toBe('executable')
      expect(tool.safeMetadata?.scope).toBe('agent')
      expect(tool.safeMetadata?.coordination_only).toBe(true)
      expect(tool.safeMetadata?.autonomous_execution).toBe(false)
      expect(`${tool.displayName} ${tool.description} ${JSON.stringify(tool.safeMetadata)}`).not.toContain('/Users/')
      expect(`${tool.displayName} ${tool.description} ${JSON.stringify(tool.safeMetadata)}`).not.toContain('raw_result')
    }
  })

  test('mock catalog includes memory tools as approval-gated safe summaries', async () => {
    const api = await import('../mockApiClient')
    const tools = await api.mockApiClient.listToolCatalog()
    const memoryTools = tools.filter((tool) => tool.group === 'memory')

    expect(memoryTools.map((tool) => tool.name).sort()).toEqual(['memory.connections', 'memory.context', 'memory.edit', 'memory.forget', 'memory.list', 'memory.read', 'memory.search', 'memory.status', 'memory.thread_fetch', 'memory.thread_search', 'memory.timeline', 'memory.write'])
    for (const tool of memoryTools) {
      expect(tool.source).toBe('builtin')
      expect(tool.riskLevel).toBe('medium')
      expect(tool.approvalPolicy).toBe('always_required')
      expect(tool.executionState).toBe('executable')
      expect(tool.safeMetadata?.scope).toBe('memory')
      expect(tool.safeMetadata?.approval_gated).toBe(true)
      expect(tool.safeMetadata?.returns_raw_content).toBe(false)
      expect(`${tool.displayName} ${tool.description} ${JSON.stringify(tool.safeMetadata)}`).not.toContain('/Users/')
      expect(`${tool.displayName} ${tool.description} ${JSON.stringify(tool.safeMetadata)}`).not.toContain('sk-')
    }
  })

  test('renders a read-only safe tool catalog surface', async () => {
    const source = await Bun.file(new URL('./SettingsView.tsx', import.meta.url)).text()

    expect(source).toContain('function ToolsPanel')
    expect(source).toContain('data-testid="tools-catalog-list"')
    expect(source).toContain('tool.source')
    expect(source).toContain('tool.group')
    expect(source).toContain('read-only')
    expect(source).toContain('toolScopeLabel(tool, locale)')
    expect(source).toContain('tool.safeMetadata?.scope')
    expect(source).toContain('non-executable')
    expect(source).toContain('coordination-only')
    expect(source).toContain('no autonomous execution')
    expect(source).toContain('approval-gated')
    expect(source).toContain('public HTTP only')
    expect(source).toContain('write-capable')
    expect(source).toContain('exec-capable')
    expect(source).toContain('tool.safeMetadata?.read_only')
    expect(source).toContain('tool.safeMetadata?.write_capable')
    expect(source).toContain('tool.safeMetadata?.exec_capable')
    expect(source).toContain('tool.safeMetadata?.coordination_only')
    expect(source).toContain('tool.riskLevel')
    expect(source).toContain('tool.approvalPolicy')
    expect(source).toContain('tool.executionState')
    expect(source).not.toContain('raw_args')
    expect(source).not.toContain('raw_result')
    const toolsPanelSource = source.slice(source.indexOf('function ToolsPanel'), source.indexOf('function WebSearchPanel'))
    expect(toolsPanelSource).not.toContain('secret')
  })

  test('Tools category is read-only, not a placeholder write surface', async () => {
    const catalog = await Bun.file(new URL('./settingsCatalog.ts', import.meta.url)).text()
    const api = await Bun.file(new URL('../realApiClient.ts', import.meta.url)).text()

    expect(catalog).toContain("{ id: 'tools', group: 'management', status: 'read_only' }")
    expect(api).toContain('/v1/tools/catalog')
    expect(api).toContain('safe_metadata')
    expect(api).not.toContain('/v1/tools/install')
    expect(api).not.toContain('/v1/tools/enable')
  })

  test('rendered workspace tools omit absolute paths from safe metadata', () => {
    const toxicTools: ToolCatalogItem[] = [{
      name: 'workspace.read',
      displayName: 'Workspace read',
      description: 'Read a bounded UTF-8 text slice from one workspace file.',
      source: 'builtin',
      group: 'workspace',
      riskLevel: 'low',
      approvalPolicy: 'read_only',
      enabled: true,
      executionState: 'executable',
      safeMetadata: { read_only: true, scope: 'workspace', root: '/Users/xuean/Repos/personal-projects/Loomi' },
    }, {
      name: 'workspace.glob',
      displayName: 'Workspace glob',
      description: 'Find files under the configured workspace root.',
      source: 'builtin',
      group: 'workspace',
      riskLevel: 'low',
      approvalPolicy: 'read_only',
      enabled: true,
      executionState: 'executable',
      safeMetadata: { read_only: true, scope: 'workspace', example: '/Users/xuean/private/example.ts' },
    }, {
      name: 'workspace.grep',
      displayName: 'Workspace grep',
      description: 'Search text files under the configured workspace root.',
      source: 'builtin',
      group: 'workspace',
      riskLevel: 'low',
      approvalPolicy: 'always_required',
      enabled: true,
      executionState: 'executable',
      safeMetadata: { read_only: true, scope: 'workspace', path: '/Users/xuean/.ssh/id_ed25519' },
    }]
    const html = renderToStaticMarkup(<SettingsView {...baseSettingsProps()} toolCatalog={toxicTools} />)

    expect(html).toContain('workspace.read')
    expect(html).toContain('workspace.glob')
    expect(html).toContain('workspace.grep')
    expect(html).toContain('工作区范围')
    expect(html).not.toContain('/Users/')
    expect(html).not.toContain('.ssh')
  })

  test('renders web search as a dedicated settings menu instead of Tools configuration', () => {
    const webTools: ToolCatalogItem[] = [{
      name: 'web.search',
      displayName: 'Web search',
      description: 'Search the public web.',
      source: 'builtin',
      group: 'web',
      riskLevel: 'medium',
      approvalPolicy: 'always_required',
      enabled: true,
      executionState: 'executable',
      safeMetadata: { read_only: true, scope: 'web', network_access: 'search_provider_api', providers: ['tavily', 'brave'], example_key: 'tvly-secret' },
    }, {
      name: 'web.fetch',
      displayName: 'Web fetch',
      description: 'Fetch a public URL.',
      source: 'builtin',
      group: 'web',
      riskLevel: 'medium',
      approvalPolicy: 'always_required',
      enabled: true,
      executionState: 'executable',
      safeMetadata: { read_only: true, scope: 'web', network_access: 'public_http_only' },
    }]

    const webSearchHtml = renderToStaticMarkup(<SettingsView {...baseSettingsProps()} selectedCategoryId="web-search" toolCatalog={webTools} />)
    const toolsHtml = renderToStaticMarkup(<SettingsView {...baseSettingsProps()} selectedCategoryId="tools" toolCatalog={webTools} />)

    expect(webSearchHtml).toContain('网页搜索')
    expect(webSearchHtml).toContain('Tavily Key')
    expect(webSearchHtml).toContain('Brave Search')
    expect(webSearchHtml).toContain('web.search')
    expect(webSearchHtml).toContain('已可用')
    expect(webSearchHtml).toContain('已保存的 key 不显示。')
    expect(webSearchHtml).not.toContain('tvly-secret')

    expect(toolsHtml).toContain('web.fetch')
    expect(toolsHtml).not.toContain('web.search')
    expect(toolsHtml).not.toContain('Tavily')
  })
})

function baseSettingsProps(): Parameters<typeof SettingsView>[0] {
  return {
    locale: 'zh',
    selectedCategoryId: 'tools',
    defaultWorkspaceMode: 'work',
    theme: 'light',
    backendCapability: 'available',
    streamState: 'closed',
    selectedThreadTitle: 'M21 smoke',
    selectedRunStatus: 'completed',
    providerCapabilities: [],
    toolCatalog: [],
    webSearchConfig: { hasTavilyKey: true, hasBraveKey: false, enabled: true },
    webSearchSaveResult: { status: 'idle' },
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
    onSaveWebSearchKeys: () => {},
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
