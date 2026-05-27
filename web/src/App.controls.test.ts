import { describe, expect, test } from 'bun:test'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { createWorkspaceShellState } from './useWorkspaceShellState'

describe('App titlebar controls', () => {
  test('uses animal-island-ui Button for titlebar chrome controls', () => {
    const source = readFileSync(resolve(import.meta.dir, 'App.tsx'), 'utf8')

    expect(source).toContain("import { Button } from 'animal-island-ui'")
    expect(source).toContain('<Button className="titlebar-button"')
    expect(source).not.toContain('<button className="titlebar-button"')
  })

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
