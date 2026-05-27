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
    expect(source).toContain('void createThread(mode)')
  })

  test('wires the collapsed titlebar compose button to create the current mode thread', () => {
    const source = readFileSync(resolve(import.meta.dir, 'App.tsx'), 'utf8')

    expect(source).toContain('titlebar-create-thread')
    expect(source).toContain("aria-label={selectedMode === 'work' ? dictionary.sidebar.newWork : dictionary.sidebar.newChat}")
    expect(source).toContain('onClick={() => void createThread(selectedMode)}')
    expect(source).toContain('SquarePen')
    expect(source).toContain('{shell.sidebarCollapsed && (')
  })

  test('passes selected thread loading error and latest run state to canvas sidebar and timeline', () => {
    const source = readFileSync(resolve(import.meta.dir, 'App.tsx'), 'utf8')

    expect(source).toContain('thread={selectedThread}')
    expect(source).toContain('messages={messages}')
    expect(source).toContain('run={run}')
    expect(source).toContain('loading={loading}')
    expect(source).toContain('error={error}')
    expect(source).toContain('RunTimeline')
    expect(source).toContain('ThreadSidebar')
  })
})
