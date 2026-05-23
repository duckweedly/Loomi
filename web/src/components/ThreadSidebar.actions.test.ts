import { describe, expect, test } from 'bun:test'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
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
