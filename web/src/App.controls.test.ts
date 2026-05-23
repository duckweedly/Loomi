import { describe, expect, test } from 'bun:test'
import { createWorkspaceShellState } from './useWorkspaceShellState'

describe('App titlebar controls', () => {
  test('keeps run details and right tools as separate controls', () => {
    const shell = createWorkspaceShellState()

    shell.toggleRunDetails()
    expect(shell.snapshot()).toMatchObject({
      runDetailsOpen: true,
      rightPanelMenuOpen: false,
      rightPanelOpen: false,
    })

    shell.toggleRightPanelMenu()
    expect(shell.snapshot()).toMatchObject({
      runDetailsOpen: false,
      rightPanelMenuOpen: true,
      rightPanelOpen: false,
    })
  })

  test('opens the right tools menu separately from the expanded right panel', () => {
    const shell = createWorkspaceShellState()

    shell.toggleRightPanelMenu()
    expect(shell.snapshot().rightPanelOpen).toBe(false)

    shell.openRightPanel('preview')
    expect(shell.snapshot()).toMatchObject({
      rightPanelMenuOpen: false,
      rightPanelOpen: true,
      selectedRightPanelId: 'preview',
    })
  })
})
