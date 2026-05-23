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
      artifactOpen: false,
      selectedRightPanelId: 'terminal' satisfies RightPanelItemId,
    })
  })

  test('opens artifact while closing run details and the right drawer', () => {
    const shell = createWorkspaceShellState()

    shell.toggleRunDetails()
    shell.openRightPanel('preview')
    shell.openArtifact()

    expect(shell.snapshot()).toMatchObject({
      runDetailsOpen: false,
      rightPanelOpen: false,
      artifactOpen: true,
      selectedRightPanelId: 'preview' satisfies RightPanelItemId,
    })
  })

  test('selects a right panel while closing the menu, run details, and artifact drawer', () => {
    const shell = createWorkspaceShellState()

    shell.toggleRightPanelMenu()
    shell.toggleRunDetails()
    shell.openArtifact()
    shell.openRightPanel('files')

    expect(shell.snapshot()).toMatchObject({
      runDetailsOpen: false,
      rightPanelMenuOpen: false,
      rightPanelOpen: true,
      artifactOpen: false,
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
      artifactOpen: false,
    })
  })
})
