import { describe, expect, test } from 'bun:test'
import { generalSettingSections, getLocalizedSettingsCategories, settingsCategories, settingsCategoryGroups } from './settingsCatalog'

describe('settings catalog', () => {
  test('defines every required M5.5 settings category', () => {
    expect(settingsCategories.map((category) => category.label)).toEqual([
      'General',
      'Providers',
      'Web Search',
      'Skill',
      'MCP',
      'Memory',
      'Tools',
      'About',
    ])
  })

  test('keeps General working and future areas visibly non-working', () => {
    expect(settingsCategories.find((category) => category.id === 'general')).toMatchObject({ group: 'primary', status: 'working' })
    expect(settingsCategories.find((category) => category.id === 'providers')).toMatchObject({ group: 'agent_core', status: 'mixed' })
    expect(settingsCategories.find((category) => category.id === 'web-search')).toMatchObject({ group: 'agent_core', status: 'mixed' })
    expect(settingsCategories.find((category) => category.id === 'skill')).toMatchObject({ group: 'agent_core', status: 'read_only' })
    expect(settingsCategories.find((category) => category.id === 'mcp')).toMatchObject({ group: 'agent_core', status: 'working' })
    expect(settingsCategories.find((category) => category.id === 'about')).toMatchObject({ group: 'management', status: 'mixed' })
    expect(settingsCategories.filter((category) => category.status === 'mock')).toEqual([])
  })

  test('keeps required navigation groups and working row vocabulary', () => {
    expect(settingsCategoryGroups.primary).toEqual({ zh: '工作区', en: 'Workspace' })
    expect(getLocalizedSettingsCategories('zh').find((category) => category.id === 'general')?.label).toBe('通用')
    expect(generalSettingSections.flatMap((section) => section.rows.map((row) => row.id))).toEqual([
      'default-workspace-mode',
      'theme',
    ])
  })
})
