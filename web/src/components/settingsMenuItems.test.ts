import { describe, expect, test } from 'bun:test'
import { createSettingsMenuItems } from './settingsMenuItems'

describe('createSettingsMenuItems', () => {
  test('keeps the sidebar settings popover to the three compact commands', () => {
    const items = createSettingsMenuItems('light')

    expect(items.map((item) => item.id)).toEqual(['settings', 'theme', 'update'])
    expect(items.map((item) => item.label)).toEqual(['Settings', 'Theme', 'Update'])
    expect(items.find((item) => item.id === 'theme')?.value).toBe('Light')
    expect(items.find((item) => item.id === 'update')?.value).toBeUndefined()
  })

  test('reflects the current theme state without changing the menu scope', () => {
    const items = createSettingsMenuItems('dark')

    expect(items.map((item) => item.id)).toEqual(['settings', 'theme', 'update'])
    expect(items.find((item) => item.id === 'theme')?.value).toBe('Dark')
  })
})
