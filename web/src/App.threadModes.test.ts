import { describe, expect, test } from 'bun:test'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

describe('App thread mode sidebar wiring', () => {
  test('passes only current-mode threads into ThreadSidebar', () => {
    const source = readFileSync(resolve(import.meta.dir, 'App.tsx'), 'utf8')

    expect(source).toContain('const selectedMode = selectedThread?.mode ?? \'chat\'')
    expect(source).toContain('const visibleThreads = filterThreadsByMode(threads, selectedMode)')
    expect(source).toContain('threads={visibleThreads}')
    expect(source).not.toContain('threads={threads}')
    expect(source).toContain('useWorkspaceState(shell.defaultWorkspaceMode)')
  })
})
