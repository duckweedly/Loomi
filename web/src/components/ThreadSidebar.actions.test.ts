import { describe, expect, test } from 'bun:test'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { createElement } from 'react'
import { renderToStaticMarkup } from 'react-dom/server'
import { ThreadSidebar } from './ThreadSidebar'
import { getDictionary } from '../i18n'
import { createSidebarModeMenuItems } from './sidebarModeMenuItems'

describe('ThreadSidebar action menu items', () => {
  test('marks New Chat as the thread creation action', () => {
    expect(createSidebarModeMenuItems('chat')).toContainEqual({
      id: 'new-chat',
      label: 'New Chat',
      action: 'create-thread',
    })
  })

  test('renders one explicit thread create icon beside the Threads label', () => {
    const source = readFileSync(resolve(import.meta.dir, 'ThreadSidebar.tsx'), 'utf8')

    expect(source).toContain('<span>{copy.threads}</span>')
    expect(source).toContain('className="thread-create-button"')
    expect(source).toContain('aria-label={copy.newChat}')
    expect(source).toContain('onClick={onCreateThread}')
    expect(source).toContain('MessageSquarePlus')
  })

  test('keeps rename and delete in a sibling row menu instead of the row selector', () => {
    const source = readFileSync(resolve(import.meta.dir, 'ThreadSidebar.tsx'), 'utf8')

    expect(source).toContain("className={thread.id === selectedThreadId ? 'thread-row selected' : 'thread-row'}")
    expect(source).toContain('className={thread.id === selectedThreadId ? \'thread-card selected\' : \'thread-card\'}')
    expect(source).toContain('className="thread-action"')
    expect(source).toContain('aria-label={copy.threadActions}')
    expect(source).toContain('className="thread-menu"')
    expect(source).toContain('renameThread(thread)')
    expect(source).not.toContain('window.prompt')
    expect(source).toContain('editingThreadId === thread.id')
    expect(source).toContain('className={thread.id === selectedThreadId ? \'thread-rename-form selected\' : \'thread-rename-form\'}')
    expect(source).toContain('submitRename(thread)')
    expect(source).toContain('deleteThread(thread)')
    expect(source).toContain('Pencil')
    expect(source).toContain('Trash2')
    expect(source).not.toContain('role="button" aria-label={copy.archiveThread}')
  })
})


describe('ThreadSidebar loading and retry states', () => {
  test('renders loading error retry and empty affordances without clearing selection', () => {
    const html = renderToStaticMarkup(createElement(ThreadSidebar, {
      collapsed: false,
      threads: [],
      selectedThreadId: 'thread-a',
      selectedMode: 'chat',
      modeCopy: { chat: 'Chat', work: 'Work' },
      theme: 'dark',
      loading: true,
      error: 'load failed',
      copy: getDictionary('en').sidebar,
      onRefresh: () => {},
      onSelectThread: () => {},
      onSelectMode: () => {},
      onCreateThread: () => {},
      onRenameThread: () => {},
      onArchiveThread: () => {},
      onToggleTheme: () => {},
      onOpenSettings: () => {},
    }))

    expect(html).toContain('Loading threads')
    expect(html).toContain('load failed')
    expect(html).toContain('Retry')
    expect(html).toContain('No chat threads')
  })

  test('keeps settings as the only fixed sidebar footer action', () => {
    const source = readFileSync(resolve(import.meta.dir, 'ThreadSidebar.tsx'), 'utf8')

    expect(source).not.toContain('sidebar-search-field')
    expect(source).toContain('className="sidebar-footer"')
    expect(source).toContain('className="sidebar-settings-button"')
    expect(source).toContain('createSidebarFooterItems')
    expect(source).not.toContain('sidebar-bottom-actions')
    expect(source).not.toContain('sidebar-search-button')
  })

  test('gives run dots readable status text', () => {
    const source = readFileSync(resolve(import.meta.dir, 'ThreadSidebar.tsx'), 'utf8')

    expect(source).toContain('runStatusCopy')
    expect(source).toContain('aria-label={runStatusCopy[thread.runStatus]}')
  })

  test('does not render the old bottom create action surface', () => {
    const chatHtml = renderToStaticMarkup(createElement(ThreadSidebar, {
      collapsed: false,
      threads: [],
      selectedThreadId: 'thread-a',
      selectedMode: 'chat',
      modeCopy: { chat: 'Chat', work: 'Work' },
      theme: 'dark',
      copy: getDictionary('en').sidebar,
      onRefresh: () => {},
      onSelectThread: () => {},
      onSelectMode: () => {},
      onCreateThread: () => {},
      onRenameThread: () => {},
      onArchiveThread: () => {},
      onToggleTheme: () => {},
      onOpenSettings: () => {},
    }))

    expect(chatHtml).not.toContain('sidebar-bottom-actions')
    expect(chatHtml).toContain('aria-label="New Chat"')
  })
})
