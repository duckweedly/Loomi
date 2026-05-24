import { describe, expect, test } from 'bun:test'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { createElement } from 'react'
import { renderToStaticMarkup } from 'react-dom/server'
import { ThreadSidebar } from './ThreadSidebar'
import { createSidebarModeMenuItems } from './sidebarModeMenuItems'

describe('ThreadSidebar action menu items', () => {
  test('marks New Chat as the thread creation action', () => {
    expect(createSidebarModeMenuItems('chat')).toContainEqual({
      id: 'new-chat',
      label: 'New Chat',
      action: 'create-thread',
    })
  })

  test('does not render a duplicate create button beside the Threads label', () => {
    const source = readFileSync(resolve(import.meta.dir, 'ThreadSidebar.tsx'), 'utf8')

    expect(source).toContain('<span>Threads</span>')
    expect(source).not.toContain('aria-label="Create thread"')
  })

  test('keeps archive as a sibling button instead of a button-like element inside the row selector', () => {
    const source = readFileSync(resolve(import.meta.dir, 'ThreadSidebar.tsx'), 'utf8')

    expect(source).toContain('className="thread-row"')
    expect(source).toContain('className={thread.id === selectedThreadId ? \'thread-card selected\' : \'thread-card\'}')
    expect(source).toContain('<button className="thread-action"')
    expect(source).not.toContain('role="button" aria-label="Archive thread"')
  })
})


describe('ThreadSidebar loading and retry states', () => {
  test('renders loading error retry and empty affordances without clearing selection', () => {
    const html = renderToStaticMarkup(createElement(ThreadSidebar, {
      collapsed: false,
      threads: [],
      selectedThreadId: 'thread-a',
      selectedMode: 'chat',
      theme: 'dark',
      loading: true,
      error: 'load failed',
      onRefresh: () => {},
      onSelectThread: () => {},
      onCreateThread: () => {},
      onRenameThread: () => {},
      onArchiveThread: () => {},
      onToggleTheme: () => {},
    }))

    expect(html).toContain('Loading threads')
    expect(html).toContain('load failed')
    expect(html).toContain('Retry')
    expect(html).toContain('No chat threads')
  })
})
