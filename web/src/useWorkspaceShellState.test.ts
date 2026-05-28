import { describe, expect, test } from 'bun:test'
import { readFileSync } from 'node:fs'
import type { RightPanelItemId } from './rightPanelItems'
import { createWorkspaceShellState } from './useWorkspaceShellState'

describe('createWorkspaceShellState', () => {
  test('does not expose the old run details surface in shell state', () => {
    const shell = createWorkspaceShellState()

    expect(shell.snapshot()).not.toHaveProperty('runDetailsOpen')
    expect(shell).not.toHaveProperty('toggleRunDetails')
  })

  test('opens artifact previews in the narrow right panel', () => {
    const shell = createWorkspaceShellState()

    shell.openArtifact('art-a')

    expect(shell.snapshot()).toMatchObject({
      previewArtifactId: 'art-a',
      rightPanelOpen: true,
      selectedRightPanelId: 'preview' satisfies RightPanelItemId,
    })
  })

  test('keeps right panel selection scoped to preview', () => {
    const shell = createWorkspaceShellState()

    shell.openArtifact()
    shell.openRightPanel('preview')

    expect(shell.snapshot()).toMatchObject({
      rightPanelMenuOpen: false,
      rightPanelOpen: true,
      selectedRightPanelId: 'preview' satisfies RightPanelItemId,
    })
  })

  test('toggles the preview drawer directly without opening a menu', () => {
    const shell = createWorkspaceShellState()

    shell.togglePreviewPanel()

    expect(shell.snapshot()).toMatchObject({
      rightPanelMenuOpen: false,
      rightPanelOpen: true,
      selectedRightPanelId: 'preview' satisfies RightPanelItemId,
    })
  })

  test('uses the titlebar preview button to close the preview drawer', () => {
    const shell = createWorkspaceShellState()

    shell.openRightPanel('preview')
    shell.togglePreviewPanel()

    expect(shell.snapshot()).toMatchObject({
      rightPanelMenuOpen: false,
      rightPanelOpen: false,
      selectedRightPanelId: 'preview' satisfies RightPanelItemId,
    })
  })

  test('opens and closes Settings without losing workspace context', () => {
    const shell = createWorkspaceShellState()

    shell.openRightPanel('preview')
    shell.openSettings()
    shell.setSettingsCategory('providers')
    shell.closeSettings()

    expect(shell.snapshot()).toMatchObject({
      settingsOpen: false,
      settingsCategoryId: 'providers',
      rightPanelOpen: true,
      selectedRightPanelId: 'preview' satisfies RightPanelItemId,
    })
  })

  test('opens Settings with General selected without mutating transient shell surfaces', () => {
    const shell = createWorkspaceShellState({ settingsCategoryId: 'tools' })

    shell.openSettings()

    expect(shell.snapshot()).toMatchObject({
      settingsOpen: true,
      settingsCategoryId: 'general',
      rightPanelMenuOpen: false,
    })
  })

  test('stores current-session settings without persistence dependencies', () => {
    const shell = createWorkspaceShellState()

    shell.setDefaultWorkspaceMode('work')
    shell.setLocale('en')
    shell.setProviderDraftSettings({ baseUrl: 'https://gateway.example.test/v1', model: 'gpt-5.5', apiKeySet: true })

    expect(shell.snapshot().defaultWorkspaceMode).toBe('work')
    expect(shell.snapshot().locale).toBe('en')
    expect(shell.snapshot().providerDraftSettings).toMatchObject({ model: 'gpt-5.5', apiKeySet: true })
    expect(createWorkspaceShellState().snapshot()).toMatchObject({ defaultWorkspaceMode: 'chat', locale: 'zh', providerDraftSettings: { baseUrl: '', model: '', apiKeySet: false } })
  })

  test('resolves the default theme from the system before a manual override exists', () => {
    const source = readFileSync(new URL('./useWorkspaceShellState.ts', import.meta.url), 'utf8')

    expect(source).toContain("window.matchMedia('(prefers-color-scheme: dark)')")
    expect(source).toContain("const themeStorageKey = 'loomi.theme'")
    expect(source).toContain('readStoredTheme() ?? resolveSystemTheme()')
    expect(source).toContain('writeStoredTheme(next.theme)')
  })
})

describe('Settings category routing', () => {
  test('opens Providers directly when requested', () => {
    const shell = createWorkspaceShellState({ settingsCategoryId: 'general' })

    shell.openSettings('providers')

    expect(shell.snapshot()).toMatchObject({
      settingsOpen: true,
      settingsCategoryId: 'providers',
    })
  })
})
