import { describe, expect, test } from 'bun:test'
import { placeholderCategoryIds, placeholderSettingRows } from './settingsCatalog'

describe('SettingsView placeholder categories', () => {
  test('covers every future placeholder category', () => {
    expect(placeholderCategoryIds).toEqual([
      'appearance',
      'connectors',
      'plugins',
      'skill',
      'mcp',
      'notebook',
      'activity-recorder',
      'context',
      'safety',
      'tools',
      'routes',
      'advanced',
    ])
  })

  test('uses disabled mock controls with safe copy', () => {
    expect(placeholderSettingRows.every((row) => row.status === 'disabled' || row.status === 'mock')).toBe(true)
    expect(placeholderSettingRows.map((row) => row.helperText).join(' ')).toContain('not connected')
    expect(placeholderSettingRows.map((row) => row.helperText).join(' ')).toContain('not connected to providers, tools, files, or backend writes')
  })

  test('preserves selected placeholder category through controlled props', async () => {
    const source = await Bun.file(new URL('./SettingsView.tsx', import.meta.url)).text()

    expect(source).toContain('selectedCategoryId')
    expect(source).toContain('onSelectCategory')
    expect(source).not.toContain('useState(')
  })

  test('does not call provider, tool, connector, file, or backend write paths from placeholders', async () => {
    const source = await Bun.file(new URL('./SettingsView.tsx', import.meta.url)).text()

    expect(source).not.toContain('fetch(')
    expect(source).not.toContain('requestJSON')
    expect(source).not.toContain('checkModelProvider')
    expect(source).not.toContain('startRun')
    expect(source).not.toContain('sendMessage(')
    expect(source).not.toContain('createThread(')
    expect(source).not.toContain('updateThread(')
    expect(source).not.toContain('archiveThread(')
  })
})
