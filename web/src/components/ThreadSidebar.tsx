import { useCallback, useState } from 'react'
import { Button } from 'animal-island-ui'
import { Check, ChevronRight, MessageSquarePlus, MoreHorizontal, Pencil, Settings, Trash2, X } from 'lucide-react'
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
  const footerItems = createSidebarFooterItems()
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
        <div className="thread-list">
          {loading && <div className="thread-state">{copy.loadingThreads}</div>}
          {error && <div className="thread-state error"><span>{error}</span><button type="button" onClick={onRefresh}>{copy.retry}</button></div>}
          {threads.length === 0 && <div className="thread-state empty">{copy.emptyThreads('chat')}</div>}
          {threads.map((thread) => (
            <div className={thread.id === selectedThreadId ? 'thread-row selected' : 'thread-row'} key={thread.id}>
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
                    onClick={() => onSelectThread(thread.id)}
                  >
                    <span className={`run-dot ${thread.runStatus}`} aria-label={runStatusCopy[thread.runStatus]} title={runStatusCopy[thread.runStatus]} />
                    <span className="thread-title">{thread.title}</span>
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
                        setThreadMenuPosition(nextId ? { top: rect.bottom + 6, left: Math.max(12, rect.right - 156) } : null)
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
