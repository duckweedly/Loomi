describe('SettingsView tools catalog', () => {
  test('renders a read-only safe tool catalog surface', async () => {
    const source = await Bun.file(new URL('./SettingsView.tsx', import.meta.url)).text()

    expect(source).toContain('function ToolsPanel')
    expect(source).toContain('data-testid="tools-catalog-list"')
    expect(source).toContain('tool.source')
    expect(source).toContain('tool.group')
    expect(source).toContain('tool.riskLevel')
    expect(source).toContain('tool.approvalPolicy')
    expect(source).toContain('tool.executionState')
    expect(source).not.toContain('raw_args')
    expect(source).not.toContain('raw_result')
    const toolsPanelSource = source.slice(source.indexOf('function ToolsPanel'), source.indexOf('function PlaceholderPanel'))
    expect(toolsPanelSource).not.toContain('secret')
  })

  test('Tools category is read-only, not a placeholder write surface', async () => {
    const catalog = await Bun.file(new URL('./settingsCatalog.ts', import.meta.url)).text()
    const api = await Bun.file(new URL('../realApiClient.ts', import.meta.url)).text()

    expect(catalog).toContain("{ id: 'tools', group: 'management', status: 'read_only' }")
    expect(api).toContain('/v1/tools/catalog')
    expect(api).not.toContain('/v1/tools/install')
    expect(api).not.toContain('/v1/tools/enable')
  })
})
