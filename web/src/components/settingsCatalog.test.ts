import { describe, expect, test } from 'bun:test'
import { generalSettingSections, getLocalizedSettingsCategories, placeholderCategoryIds, settingsCategories, settingsCategoryGroups } from './settingsCatalog'

describe('settings catalog', () => {
  test('defines every required M5.5 settings category', () => {
    expect(settingsCategories.map((category) => category.label)).toEqual([
      'General',
      'Appearance',
      'Providers',
      'Connectors',
      'Plugins',
      'Skill',
      'MCP',
      'Notebook',
      'Memory',
      'Activity Recorder',
      'Context',
      'Safety',
      'Tools',
      'Routes',
      'About',
      'Advanced',
    ])
  })

  test('keeps implemented areas distinct from future placeholders', () => {
    expect(settingsCategories.find((category) => category.id === 'general')).toMatchObject({ group: 'primary', status: 'working' })
    expect(settingsCategories.find((category) => category.id === 'providers')).toMatchObject({ group: 'agent_core', status: 'mixed' })
    expect(settingsCategories.find((category) => category.id === 'tools')).toMatchObject({ group: 'management', status: 'read_only' })
    expect(settingsCategories.find((category) => category.id === 'about')).toMatchObject({ group: 'management', status: 'mixed' })
    expect(placeholderCategoryIds).not.toContain('general')
    expect(placeholderCategoryIds).not.toContain('tools')
    expect(settingsCategories.filter((category) => category.status === 'mock').length).toBeGreaterThan(10)
  })

  test('keeps required navigation groups and working row vocabulary', () => {
    expect(settingsCategoryGroups.primary).toEqual({ zh: '工作区', en: 'Workspace' })
    expect(getLocalizedSettingsCategories('zh').find((category) => category.id === 'general')?.label).toBe('通用')
    expect(generalSettingSections.flatMap((section) => section.rows.map((row) => row.id))).toEqual([
      'default-workspace-mode',
      'mock-runtime-scenario',
      'data-source-mode',
      'backend-capability',
      'provider-capability',
    ])
  })
})
