import { describe, expect, test } from 'bun:test'
import type { RightPanelItemId } from './rightPanelItems'
import { createWorkspaceShellState } from './useWorkspaceShellState'

describe('createWorkspaceShellState', () => {
  test('opens run details while closing right-side tool surfaces', () => {
    const shell = createWorkspaceShellState()

    shell.openRightPanel('terminal')
    shell.openArtifact()
    shell.toggleRunDetails()

    expect(shell.snapshot()).toMatchObject({
      runDetailsOpen: true,
      rightPanelMenuOpen: false,
      rightPanelOpen: false,
      selectedRightPanelId: 'preview' satisfies RightPanelItemId,
    })
  })

  test('opens artifact previews in the narrow right panel', () => {
    const shell = createWorkspaceShellState()

    shell.toggleRunDetails()
    shell.openArtifact()

    expect(shell.snapshot()).toMatchObject({
      runDetailsOpen: false,
      rightPanelOpen: true,
      selectedRightPanelId: 'preview' satisfies RightPanelItemId,
    })
  })

  test('selects a right panel while closing the menu and run details', () => {
    const shell = createWorkspaceShellState()

    shell.toggleRightPanelMenu()
    shell.toggleRunDetails()
    shell.openArtifact()
    shell.openRightPanel('files')

    expect(shell.snapshot()).toMatchObject({
      runDetailsOpen: false,
      rightPanelMenuOpen: false,
      rightPanelOpen: true,
      selectedRightPanelId: 'files' satisfies RightPanelItemId,
    })
  })

  test('toggles the right panel menu without opening the right drawer', () => {
    const shell = createWorkspaceShellState()

    shell.toggleRightPanelMenu()

    expect(shell.snapshot()).toMatchObject({
      runDetailsOpen: false,
      rightPanelMenuOpen: true,
      rightPanelOpen: false,
    })
  })

  test('uses the titlebar right panel button to close an open detail panel', () => {
    const shell = createWorkspaceShellState()

    shell.openRightPanel('preview')
    shell.toggleRightPanelMenu()

    expect(shell.snapshot()).toMatchObject({
      rightPanelMenuOpen: false,
      rightPanelOpen: false,
      selectedRightPanelId: 'preview' satisfies RightPanelItemId,
    })
  })

  test('opens and closes Settings without losing workspace context', () => {
    const shell = createWorkspaceShellState()

    shell.openRightPanel('files')
    shell.openSettings()
    shell.setSettingsCategory('providers')
    shell.closeSettings()

    expect(shell.snapshot()).toMatchObject({
      settingsOpen: false,
      settingsCategoryId: 'providers',
      rightPanelOpen: true,
      selectedRightPanelId: 'files' satisfies RightPanelItemId,
    })
  })

  test('opens Settings with General selected without mutating transient shell surfaces', () => {
    const shell = createWorkspaceShellState({ settingsCategoryId: 'tools' })

    shell.toggleRunDetails()
    shell.openSettings()

    expect(shell.snapshot()).toMatchObject({
      settingsOpen: true,
      settingsCategoryId: 'general',
      rightPanelMenuOpen: false,
      runDetailsOpen: true,
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
})
