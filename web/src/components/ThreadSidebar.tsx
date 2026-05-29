import { useCallback, useMemo, useState } from 'react'
import { Button } from 'animal-island-ui'
import { Check, ChevronRight, MessageSquarePlus, MoreHorizontal, Pencil, Search, Settings, Trash2, X } from 'lucide-react'
import type { Thread } from '../domain'
import { LoomiFloatingMenu, LoomiMenuItem } from './LoomiMenu'
import { createSidebarFooterItems } from './sidebarFooterItems'

type SidebarCopy = {
  newChat: string
  threads: string
  settings: string
  theme: string
  update: string
  open: string
  threadActions: string
  light: string
  dark: string
  archiveThread: string
  renameThread: string
  loadingThreads: string
  retry: string
  searchThreads: string
  emptyThreads: (mode: Thread['mode']) => string
}

type Props = {
  collapsed: boolean
  threads: Thread[]
  selectedThreadId: string
  theme: 'dark' | 'light'
  loading?: boolean
  error?: string | null
  copy: SidebarCopy
  onRefresh: () => void
  onSelectThread: (threadId: string) => void
  onCreateThread: () => void
  onRenameThread: (threadId: string, title: string) => void
  onArchiveThread: (threadId: string) => void
  onToggleTheme: () => void
  onOpenSettings: () => void
}

function getThreadMenuPosition(rect: DOMRect) {
  const threadMenuWidth = 156
  const threadMenuHeight = 96
  const maxTop = Math.max(12, window.innerHeight - threadMenuHeight - 12)
  const top = Math.max(12, Math.min(rect.bottom + 6, maxTop))
  const left = Math.max(12, Math.min(rect.right - threadMenuWidth, window.innerWidth - threadMenuWidth - 12))
  return { top, left }
}

type SidebarLocale = 'zh' | 'en'

export function filterThreadsForSidebar(threads: Thread[], query: string) {
  const normalizedQuery = query.trim().toLowerCase()
  if (!normalizedQuery) return threads
  return threads.filter((thread) => (
    thread.title.toLowerCase().includes(normalizedQuery) ||
    thread.project.toLowerCase().includes(normalizedQuery) ||
    thread.mode.toLowerCase().includes(normalizedQuery) ||
    thread.runStatus.toLowerCase().includes(normalizedQuery)
  ))
}

function startOfDay(date: Date) {
  return new Date(date.getFullYear(), date.getMonth(), date.getDate()).getTime()
}

function parseThreadTime(updatedAt: string) {
  const time = new Date(updatedAt).getTime()
  return Number.isNaN(time) ? Number.POSITIVE_INFINITY : time
}

function groupLabelForThread(thread: Thread, now: Date, locale: SidebarLocale) {
  const updatedAt = new Date(thread.updatedAt)
  if (Number.isNaN(updatedAt.getTime())) return locale === 'zh' ? '最近' : 'Recent'
  const diffDays = Math.floor((startOfDay(now) - startOfDay(updatedAt)) / 86_400_000)
  if (diffDays <= 0) return locale === 'zh' ? '今天' : 'Today'
  if (diffDays === 1) return locale === 'zh' ? '昨天' : 'Yesterday'
  if (diffDays < 7) return locale === 'zh' ? '本周' : 'This week'
  return locale === 'zh' ? '更早' : 'Earlier'
}

export function groupThreadsForSidebar(threads: Thread[], now = new Date(), locale: SidebarLocale = 'en') {
  const groups = new Map<string, Thread[]>()
  for (const thread of [...threads].sort((a, b) => parseThreadTime(b.updatedAt) - parseThreadTime(a.updatedAt))) {
    const label = groupLabelForThread(thread, now, locale)
    groups.set(label, [...(groups.get(label) ?? []), thread])
  }
  return [...groups.entries()].map(([label, groupedThreads]) => ({ label, threads: groupedThreads }))
}

export function formatThreadUpdatedAt(value: string, locale: SidebarLocale) {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return date.toLocaleTimeString(locale === 'zh' ? 'zh-CN' : 'en-US', { hour: '2-digit', minute: '2-digit' })
}

export function ThreadSidebar({
  collapsed,
  threads,
  selectedThreadId,
  loading = false,
  error = null,
  copy,
  onRefresh,
  onSelectThread,
  onCreateThread,
  onRenameThread,
  onArchiveThread,
  onOpenSettings,
}: Props) {
  const [threadMenuId, setThreadMenuId] = useState<string | null>(null)
  const [threadMenuPosition, setThreadMenuPosition] = useState<{ top: number; left: number } | null>(null)
  const [editingThreadId, setEditingThreadId] = useState<string | null>(null)
  const [renameDraft, setRenameDraft] = useState('')
  const [threadSearch, setThreadSearch] = useState('')
  const footerItems = createSidebarFooterItems()
  const showThreadLoading = loading && threads.length === 0
  const sidebarLocale: SidebarLocale = copy.searchThreads === '搜索会话' ? 'zh' : 'en'
  const visibleThreads = useMemo(() => filterThreadsForSidebar(threads, threadSearch), [threadSearch, threads])
  const threadGroups = useMemo(() => groupThreadsForSidebar(visibleThreads, new Date(), sidebarLocale), [sidebarLocale, visibleThreads])
  const runStatusCopy: Record<Thread['runStatus'], string> = {
    pending: 'Pending',
    queued: 'Queued',
    running: 'Running',
    recovering: 'Recovering',
    blocked_on_tool_approval: 'Waiting for approval',
    stopping: 'Stopping',
    completed: 'Completed',
    failed: 'Failed',
    stopped: 'Stopped',
    cancelled: 'Cancelled',
    retrying: 'Retrying',
  }

  const renameThread = (thread: Thread) => {
    setThreadMenuId(null)
    setThreadMenuPosition(null)
    setEditingThreadId(thread.id)
    setRenameDraft(thread.title)
  }

  const submitRename = (thread: Thread) => {
    const title = renameDraft.trim()
    setEditingThreadId(null)
    setRenameDraft('')
    if (title && title !== thread.title) onRenameThread(thread.id, title)
  }

  const cancelRename = () => {
    setEditingThreadId(null)
    setRenameDraft('')
  }

  const deleteThread = (thread: Thread) => {
    setThreadMenuId(null)
    setThreadMenuPosition(null)
    if (window.confirm(copy.archiveThread)) onArchiveThread(thread.id)
  }

  const closeThreadMenu = useCallback(() => {
    setThreadMenuId(null)
    setThreadMenuPosition(null)
  }, [])

  const selectThread = (threadId: string) => {
    setThreadMenuId(null)
    setThreadMenuPosition(null)
    onSelectThread(threadId)
  }

  const threadRowClassName = (thread: Thread) => {
    const classNames = ['thread-row']
    if (thread.id === selectedThreadId) classNames.push('selected')
    if (threadMenuId === thread.id) classNames.push('menu-open')
    return classNames.join(' ')
  }

  if (collapsed) return null

  return (
    <aside className="sidebar">
      <div className="sidebar-section thread-list-section">
        <div className="section-label">
          <span>{copy.threads}</span>
          <Button className="thread-create-button" htmlType="button" aria-label={copy.newChat} title={copy.newChat} onClick={onCreateThread}>
            <MessageSquarePlus size={14} />
          </Button>
        </div>
        <label className="sidebar-thread-search">
          <Search size={13} aria-hidden="true" />
          <input
            type="search"
            value={threadSearch}
            placeholder={copy.searchThreads}
            aria-label={copy.searchThreads}
            onChange={(event) => setThreadSearch(event.target.value)}
          />
        </label>
        <div className="thread-list">
          {showThreadLoading && <div className="thread-state">{copy.loadingThreads}</div>}
          {error && <div className="thread-state error"><span>{error}</span><button type="button" onClick={onRefresh}>{copy.retry}</button></div>}
          {threads.length === 0 && <div className="thread-state empty">{copy.emptyThreads('chat')}</div>}
          {threads.length > 0 && visibleThreads.length === 0 && <div className="thread-state empty">{copy.emptyThreads('chat')}</div>}
          {threadGroups.map((group) => (
            <section className="thread-group" key={group.label}>
              <div className="thread-group-label">{group.label}</div>
              {group.threads.map((thread) => (
                <div className={threadRowClassName(thread)} key={thread.id}>
                  {editingThreadId === thread.id ? (
                    <form
                      className={thread.id === selectedThreadId ? 'thread-rename-form selected' : 'thread-rename-form'}
                      onSubmit={(event) => {
                        event.preventDefault()
                        submitRename(thread)
                      }}
                    >
                      <span className={`run-dot ${thread.runStatus}`} aria-label={runStatusCopy[thread.runStatus]} title={runStatusCopy[thread.runStatus]} />
                      <input
                        aria-label={copy.renameThread}
                        autoFocus
                        value={renameDraft}
                        onChange={(event) => setRenameDraft(event.target.value)}
                        onKeyDown={(event) => {
                          if (event.key === 'Escape') cancelRename()
                        }}
                      />
                      <button type="submit" aria-label={copy.renameThread}><Check size={13} /></button>
                      <button type="button" aria-label="Cancel rename" onClick={cancelRename}><X size={13} /></button>
                    </form>
                  ) : (
                    <>
                      <Button
                        className={thread.id === selectedThreadId ? 'thread-card selected' : 'thread-card'}
                        onClick={() => selectThread(thread.id)}
                      >
                        <span className={`run-dot ${thread.runStatus}`} aria-label={runStatusCopy[thread.runStatus]} title={runStatusCopy[thread.runStatus]} />
                        <span className="thread-copy">
                          <span className="thread-title">{thread.title}</span>
                          <span className="thread-meta">
                            <span>{runStatusCopy[thread.runStatus]}</span>
                            <span aria-hidden="true">·</span>
                            <time dateTime={thread.updatedAt}>{formatThreadUpdatedAt(thread.updatedAt, sidebarLocale)}</time>
                          </span>
                        </span>
                      </Button>
                      <Button
                        className="thread-action"
                        data-loomi-menu-trigger="thread-menu"
                        aria-expanded={threadMenuId === thread.id}
                        aria-label={copy.threadActions}
                        onClick={(event) => {
                          event.stopPropagation()
                          const rect = event.currentTarget.getBoundingClientRect()
                          setThreadMenuId((current) => {
                            const nextId = current === thread.id ? null : thread.id
                            setThreadMenuPosition(nextId ? getThreadMenuPosition(rect) : null)
                            return nextId
                          })
                        }}
                      >
                        <MoreHorizontal size={14} />
                      </Button>
                    </>
                  )}
                  {threadMenuId === thread.id && (
                    <LoomiFloatingMenu
                      open
                      className="thread-menu"
                      ignoreSelector="[data-loomi-menu-trigger='thread-menu']"
                      onClose={closeThreadMenu}
                      style={threadMenuPosition ? { top: threadMenuPosition.top, left: threadMenuPosition.left, width: 156 } : undefined}
                    >
                      <LoomiMenuItem onClick={() => renameThread(thread)}><Pencil size={13} /> <span>{copy.renameThread}</span></LoomiMenuItem>
                      <LoomiMenuItem className="danger" onClick={() => deleteThread(thread)}><Trash2 size={13} /> <span>{copy.archiveThread}</span></LoomiMenuItem>
                    </LoomiFloatingMenu>
                  )}
                </div>
              ))}
            </section>
          ))}
        </div>
      </div>
      <div className="sidebar-footer">
        {footerItems.map((item) => (
          <Button className="sidebar-settings-button" key={item.id} htmlType="button" aria-label={copy.settings} onClick={onOpenSettings}>
            <span className="sidebar-settings-icon" aria-hidden="true"><Settings size={17} /></span>
            <span className="sidebar-settings-label">{copy.settings}</span>
            <ChevronRight className="sidebar-settings-chevron" size={15} aria-hidden="true" />
          </Button>
        ))}
      </div>
    </aside>
  )
}
