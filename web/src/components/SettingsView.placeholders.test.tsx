import { describe, expect, test } from 'bun:test'
import { settingsCategories } from './settingsCatalog'

describe('SettingsView placeholder categories', () => {
  test('does not expose placeholder-only categories', () => {
    expect(settingsCategories.filter((category) => category.status === 'mock')).toEqual([])
    expect(settingsCategories.map((category) => category.label)).not.toContain('Plugins')
  })

  test('removes the generic placeholder panel from SettingsView', async () => {
    const source = await Bun.file(new URL('./SettingsView.tsx', import.meta.url)).text()

    expect(source).not.toContain('function PlaceholderPanel')
    expect(source).not.toContain('placeholderSettingRows')
  })

  test('keeps SettingsView category selection controlled by shell state', async () => {
    const source = await Bun.file(new URL('./SettingsView.tsx', import.meta.url)).text()

    expect(source).toContain('selectedCategoryId')
    expect(source).toContain('onSelectCategory')
  })
})
