import { describe, expect, test } from 'bun:test'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { createElement } from 'react'
import { renderToStaticMarkup } from 'react-dom/server'
import { ThreadSidebar, filterThreadsForSidebar, groupThreadsForSidebar, formatThreadUpdatedAt } from './ThreadSidebar'
import { getDictionary } from '../i18n'

describe('ThreadSidebar action menu items', () => {
  test('filters threads locally by title and project before rendering groups', () => {
    const threads = [
      { id: 'thread-a', title: 'Daily notes', project: 'Loomi', mode: 'chat' as const, updatedAt: '2026-05-29T10:00:00Z', lifecycleStatus: 'active' as const, runStatus: 'completed' as const },
      { id: 'thread-b', title: 'Research plan', project: 'Craft', mode: 'work' as const, updatedAt: '2026-05-29T09:00:00Z', lifecycleStatus: 'active' as const, runStatus: 'running' as const },
    ]

    expect(filterThreadsForSidebar(threads, 'craft').map((thread) => thread.id)).toEqual(['thread-b'])
    expect(filterThreadsForSidebar(threads, 'daily').map((thread) => thread.id)).toEqual(['thread-a'])
    expect(filterThreadsForSidebar(threads, 'missing')).toEqual([])
  })

  test('groups threads into scan-friendly recency buckets', () => {
    const now = new Date('2026-05-29T12:00:00Z')
    const groups = groupThreadsForSidebar([
      { id: 'thread-a', title: 'Today', project: 'Loomi', mode: 'chat' as const, updatedAt: '2026-05-29T10:00:00Z', lifecycleStatus: 'active' as const, runStatus: 'completed' as const },
      { id: 'thread-b', title: 'Yesterday', project: 'Loomi', mode: 'chat' as const, updatedAt: '2026-05-28T10:00:00Z', lifecycleStatus: 'active' as const, runStatus: 'completed' as const },
      { id: 'thread-c', title: 'Earlier', project: 'Loomi', mode: 'chat' as const, updatedAt: '2026-05-20T10:00:00Z', lifecycleStatus: 'active' as const, runStatus: 'completed' as const },
    ], now)

    expect(groups.map((group) => [group.label, group.threads.map((thread) => thread.id)])).toEqual([
      ['Today', ['thread-a']],
      ['Yesterday', ['thread-b']],
      ['Earlier', ['thread-c']],
    ])
  })

  test('renders sidebar search grouping and thread metadata like compact workspace rows', () => {
    const html = renderToStaticMarkup(createElement(ThreadSidebar, {
      collapsed: false,
      threads: [{
        id: 'thread-a',
        title: 'Craft polish pass',
        project: 'Loomi',
        mode: 'work',
        updatedAt: '2026-05-29T10:00:00Z',
        lifecycleStatus: 'active',
        runStatus: 'running',
      }],
      selectedThreadId: 'thread-a',
      theme: 'dark',
      copy: getDictionary('en').sidebar,
      onRefresh: () => {},
      onSelectThread: () => {},
      onCreateThread: () => {},
      onRenameThread: () => {},
      onArchiveThread: () => {},
      onToggleTheme: () => {},
      onOpenSettings: () => {},
    }))

    expect(html).toContain('sidebar-thread-search')
    expect(html).toContain('placeholder="Search threads"')
    expect(html).toContain('thread-group-label')
    expect(html).toContain('Today')
    expect(html).toContain('thread-meta')
    expect(html).toContain('Running')
    expect(html).toContain('<time')
  })

  test('formats thread updated timestamps for compact sidebar rows', () => {
    expect(formatThreadUpdatedAt('2026-05-29T10:05:00Z', 'en')).toMatch(/10:05|18:05|6:05/)
    expect(formatThreadUpdatedAt('Now', 'en')).toBe('Now')
  })

  test('renders one explicit thread create icon beside the Threads label', () => {
    const source = readFileSync(resolve(import.meta.dir, 'ThreadSidebar.tsx'), 'utf8')

    expect(source).toContain("import { Button } from 'animal-island-ui'")
    expect(source).toContain('<span>{copy.threads}</span>')
    expect(source).toContain('className="thread-create-button"')
    expect(source).toContain('<Button className="thread-create-button"')
    expect(source).toContain('aria-label={copy.newChat}')
    expect(source).toContain('onClick={onCreateThread}')
    expect(source).toContain('MessageSquarePlus')
  })

  test('keeps rename and delete in a sibling row menu instead of the row selector', () => {
    const source = readFileSync(resolve(import.meta.dir, 'ThreadSidebar.tsx'), 'utf8')

    expect(source).toContain('threadRowClassName(thread)')
    expect(source).toContain('className={thread.id === selectedThreadId ? \'thread-card selected\' : \'thread-card\'}')
    expect(source).toContain('<Button')
    expect(source).not.toContain('className="nav-item"')
    expect(source).toContain('className="thread-action"')
    expect(source).toContain('className="thread-action"')
    expect(source).toContain('aria-label={copy.threadActions}')
    expect(source).toContain('className="thread-menu"')
    expect(source).toContain('threadMenuPosition')
    expect(source).toContain('getBoundingClientRect()')
    expect(source).toContain("import { LoomiFloatingMenu, LoomiMenuItem } from './LoomiMenu'")
    expect(source).toContain('data-loomi-menu-trigger="thread-menu"')
    expect(source).toContain('<LoomiFloatingMenu')
    expect(source).toContain('ignoreSelector="[data-loomi-menu-trigger=\'thread-menu\']"')
    expect(source).toContain('setThreadMenuPosition(null)')
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

  test('keeps the row visually stable while menus open or selection changes', () => {
    const source = readFileSync(resolve(import.meta.dir, 'ThreadSidebar.tsx'), 'utf8')
    const css = readFileSync(resolve(import.meta.dir, '../styles/92-unified-workspace.css'), 'utf8')

    expect(source).toContain("if (threadMenuId === thread.id) classNames.push('menu-open')")
    expect(source).toContain('const selectThread = (threadId: string) => {')
    expect(source).toContain('setThreadMenuId(null)')
    expect(source).toContain('onSelectThread(threadId)')
    expect(source).toContain('onClick={() => selectThread(thread.id)}')
    expect(css).toContain('contain: layout paint style !important;')
    expect(css).toContain('animation: none !important;')
    expect(css).toContain('transition: none !important;')
    expect(css).toContain("grid-template-columns: 10px minmax(0, 1fr) !important;")
    expect(css).toContain('.thread-row:hover > .thread-action,')
    expect(css).toContain('.thread-row.menu-open > .thread-action')
    expect(css).not.toContain('.thread-row.selected > .thread-action,\n.thread-row.menu-open > .thread-action')
    expect(css).toContain('opacity: 0 !important;')
    expect(css).toContain('opacity: 1 !important;')
  })

  test('keeps row menus inside the visible viewport like a collision-aware menu', () => {
    const source = readFileSync(resolve(import.meta.dir, 'ThreadSidebar.tsx'), 'utf8')

    expect(source).toContain('function getThreadMenuPosition')
    expect(source).toContain('window.innerHeight')
    expect(source).toContain('threadMenuHeight')
    expect(source).toContain('Math.min(rect.bottom + 6, maxTop)')
    expect(source).toContain('Math.max(12, Math.min(rect.right - threadMenuWidth, window.innerWidth - threadMenuWidth - 12))')
  })
})


describe('ThreadSidebar loading and retry states', () => {
  test('renders loading error retry and empty affordances without clearing selection', () => {
    const html = renderToStaticMarkup(createElement(ThreadSidebar, {
      collapsed: false,
      threads: [],
      selectedThreadId: 'thread-a',
      theme: 'dark',
      loading: true,
      error: 'load failed',
      copy: getDictionary('en').sidebar,
      onRefresh: () => {},
      onSelectThread: () => {},
      onCreateThread: () => {},
      onRenameThread: () => {},
      onArchiveThread: () => {},
      onToggleTheme: () => {},
      onOpenSettings: () => {},
    }))

    expect(html).toContain('Loading threads')
    expect(html).toContain('load failed')
    expect(html).toContain('Retry')
    expect(html).toContain('No threads')
  })

  test('keeps existing thread rows stable during background refreshes', () => {
    const html = renderToStaticMarkup(createElement(ThreadSidebar, {
      collapsed: false,
      threads: [{
        id: 'thread-a',
        title: 'Thread A',
        project: 'Loomi',
        mode: 'chat',
        updatedAt: 'Now',
        lifecycleStatus: 'active',
        runStatus: 'completed',
      }],
      selectedThreadId: 'thread-a',
      theme: 'dark',
      loading: true,
      copy: getDictionary('en').sidebar,
      onRefresh: () => {},
      onSelectThread: () => {},
      onCreateThread: () => {},
      onRenameThread: () => {},
      onArchiveThread: () => {},
      onToggleTheme: () => {},
      onOpenSettings: () => {},
    }))

    expect(html).toContain('Thread A')
    expect(html).not.toContain('Loading threads')
  })

  test('keeps settings as the fixed sidebar footer action', () => {
    const source = readFileSync(resolve(import.meta.dir, 'ThreadSidebar.tsx'), 'utf8')

    expect(source).not.toContain('sidebar-search-field')
    expect(source).toContain('className="sidebar-footer"')
    expect(source).toContain('className="sidebar-settings-button"')
    expect(source).toContain('className="sidebar-settings-icon"')
    expect(source).toContain('className="sidebar-settings-label"')
    expect(source).toContain('className="sidebar-settings-chevron"')
    expect(source).toContain('<Button className="sidebar-settings-button"')
    expect(source).not.toContain('title={copy.settings}')
    expect(source).not.toContain('className="sidebar-theme-toggle"')
    expect(source).not.toContain('settings-button-sky')
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
      theme: 'dark',
      copy: getDictionary('en').sidebar,
      onRefresh: () => {},
      onSelectThread: () => {},
      onCreateThread: () => {},
      onRenameThread: () => {},
      onArchiveThread: () => {},
      onToggleTheme: () => {},
      onOpenSettings: () => {},
    }))

    expect(chatHtml).not.toContain('sidebar-bottom-actions')
    expect(chatHtml).toContain('aria-label="New thread"')
  })
})
