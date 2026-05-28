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

  test('keeps the old run-details entry out of the titlebar shell', () => {
    const source = readFileSync(resolve(import.meta.dir, 'App.tsx'), 'utf8')

    expect(source).not.toContain('openRunDetails')
    expect(source).not.toContain('toggleRunDetails')
    expect(source).not.toContain('AlertCircle')
  })

  test('uses the titlebar utility as a direct preview toggle', () => {
    const shell = createWorkspaceShellState()

    shell.togglePreviewPanel()
    expect(shell.snapshot()).toMatchObject({
      rightPanelMenuOpen: false,
      rightPanelOpen: true,
      selectedRightPanelId: 'preview',
    })
  })

  test('clicking the preview utility again closes the drawer without opening a menu', () => {
    const shell = createWorkspaceShellState()

    shell.togglePreviewPanel()
    shell.togglePreviewPanel()
    expect(shell.snapshot()).toMatchObject({
      rightPanelMenuOpen: false,
      rightPanelOpen: false,
      selectedRightPanelId: 'preview',
    })
  })
})
