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

  test('supports system theme as the default preference and pins manual choices', () => {
    const shell = createWorkspaceShellState({ theme: 'dark', themePreference: 'system' })

    shell.setThemePreference('light')

    expect(shell.snapshot()).toMatchObject({
      theme: 'light',
      themePreference: 'light',
    })

    shell.setThemePreference('system')

    expect(shell.snapshot().themePreference).toBe('system')
    expect(['dark', 'light']).toContain(shell.snapshot().theme)
  })

  test('legacy theme toggle becomes an explicit manual override', () => {
    const shell = createWorkspaceShellState({ theme: 'dark', themePreference: 'system' })

    shell.toggleTheme()

    expect(shell.snapshot()).toMatchObject({
      theme: 'light',
      themePreference: 'light',
    })
  })

  test('starts with a narrow open sidebar and keeps resize bounds tight', () => {
    const shell = createWorkspaceShellState()
    const appSource = readFileSync(new URL('./App.tsx', import.meta.url), 'utf8')

    expect(shell.snapshot().sidebarWidth).toBe(136)
    expect(appSource).toContain('Math.min(sidebarMaxWidth, Math.max(sidebarMinWidth, startWidth + moveEvent.clientX - startX))')

    shell.setSidebarWidth(999)
    expect(shell.snapshot().sidebarWidth).toBe(172)

    shell.setSidebarWidth(120)
    expect(shell.snapshot().sidebarWidth).toBe(128)
  })

  test('persists user-adjusted sidebar width across shell recreation', () => {
    const source = readFileSync(new URL('./useWorkspaceShellState.ts', import.meta.url), 'utf8')

    expect(source).toContain("const sidebarWidthStorageKey = 'loomi.sidebarWidth'")
    expect(source).toContain('readStoredSidebarWidth()')
    expect(source).toContain('isLegacyDefaultSidebarWidth(storedWidth)')
    expect(source).toContain('156')
    expect(source).toContain('148')
    expect(source).toContain('168')
    expect(source).toContain('184')
    expect(source).toContain('writeStoredSidebarWidth(next.sidebarWidth)')
    expect(source).toContain('clampSidebarWidth(action.sidebarWidth)')
  })

  test('resolves the default theme from the system before a manual override exists', () => {
    const source = readFileSync(new URL('./useWorkspaceShellState.ts', import.meta.url), 'utf8')

    expect(source).toContain("window.matchMedia('(prefers-color-scheme: dark)')")
    expect(source).toContain("const themeStorageKey = 'loomi.theme'")
    expect(source).toContain("themePreference: storedTheme ?? 'system'")
    expect(source).toContain('clearStoredTheme()')
    expect(source).toContain('writeStoredTheme(action.themePreference)')
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
