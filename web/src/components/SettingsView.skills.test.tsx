import { describe, expect, test } from 'bun:test'
import { renderToStaticMarkup } from 'react-dom/server'
import { SettingsView } from './SettingsView'

describe('SettingsView skills', () => {
  test('renders installed skills and persona summaries as a real read-only page', () => {
    const html = renderToStaticMarkup(
      <SettingsView
        {...baseSettingsProps()}
        selectedCategoryId="skill"
        personas={[{ id: 'persona_1', slug: 'default', name: 'Default', description: 'Default persona.', activeVersion: '1', isDefault: true }]}
        installedSkills={[{ id: 'skill_1', name: 'speckit-implement', description: 'Execute tasks.', source: 'project', sourceLabel: 'Project .agents', package: 'speckit', path: '/repo/.agents/skills/speckit-implement/SKILL.md', installed: true }]}
      />,
    )

    expect(html).toContain('data-testid="skill-settings"')
    expect(html).toContain('speckit-implement')
    expect(html).toContain('Project .agents')
    expect(html).toContain('Default')
    expect(html).not.toContain('Mock only')
  })
})

function baseSettingsProps(): Parameters<typeof SettingsView>[0] {
  return {
    locale: 'zh',
    selectedCategoryId: 'skill',
    defaultWorkspaceMode: 'work',
    theme: 'light',
    backendCapability: 'available',
    streamState: 'closed',
    providerCapabilities: [],
    toolCatalog: [],
    localProviderDetections: [],
    memoryEntries: [],
    memoryQuery: '',
    memoryFilters: {},
    memoryLoading: false,
    memoryAuditItems: [],
    memoryAuditLoading: false,
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
  }
}
