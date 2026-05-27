import { describe, expect, test } from 'bun:test'
import { getSettingsCategoriesByGroup, getSettingsCategory } from './settingsCatalog'

describe('SettingsView navigation contract', () => {
  test('uses General as the default category from shell state', () => {
    expect(getSettingsCategory('general')).toMatchObject({ label: 'General', status: 'working' })
  })

  test('includes a back affordance in the settings view source', async () => {
    const source = await Bun.file(new URL('./SettingsView.tsx', import.meta.url)).text()

    expect(source).toContain('onBack')
    expect(source).toContain('t.back')
  })

  test('groups categories into productized settings sections', () => {
    expect(getSettingsCategoriesByGroup('primary').map((category) => category.label)).toEqual(['General'])
    expect(getSettingsCategoriesByGroup('agent_core').map((category) => category.label)).toEqual(['Model Connections', 'Web Search', 'Skill', 'MCP', 'Memory & Context'])
    expect(getSettingsCategoriesByGroup('management').map((category) => category.label)).toEqual(['Tool Permissions', 'About'])
  })
})
