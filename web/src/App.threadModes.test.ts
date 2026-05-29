import { describe, expect, test } from 'bun:test'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

describe('App thread mode sidebar wiring', () => {
  test('passes the unified thread list into ThreadSidebar', () => {
    const source = readFileSync(resolve(import.meta.dir, 'App.tsx'), 'utf8')

    expect(source).not.toContain('filterThreadsByMode')
    expect(source).not.toContain('const selectedMode')
    expect(source).toContain('threads={threads}')
    expect(source).toContain('useWorkspaceState(shell.defaultWorkspaceMode)')
    expect(source).toContain('const handleCreateThread = useCallback(() => {')
    expect(source).toContain('onCreateThread={handleCreateThread}')
  })

  test('wires the collapsed titlebar compose button to create one default thread kind', () => {
    const source = readFileSync(resolve(import.meta.dir, 'App.tsx'), 'utf8')

    expect(source).toContain('titlebar-create-thread')
    expect(source).toContain('aria-label={dictionary.sidebar.newChat}')
    expect(source).toContain('onClick={() => void createThread()}')
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
