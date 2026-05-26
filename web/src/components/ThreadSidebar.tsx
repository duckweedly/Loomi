import { useState } from 'react'
import { BriefcaseBusiness, Check, Clock3, FolderKanban, MessageCircle, MessageSquarePlus, MoreHorizontal, Pencil, Settings, Trash2, X } from 'lucide-react'
import type { Thread } from '../domain'
import { createSidebarFooterItems } from './sidebarFooterItems'
import { createSidebarModeMenuItems, type SidebarMode } from './sidebarModeMenuItems'

type SidebarCopy = {
  newChat: string
  newWork: string
  projects: string
  scheduled: string
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
  emptyThreads: (mode: SidebarMode) => string
}

type Props = {
  collapsed: boolean
  threads: Thread[]
  selectedThreadId: string
  selectedMode: SidebarMode
  modeCopy: {
    chat: string
    work: string
  }
  theme: 'dark' | 'light'
  loading?: boolean
  error?: string | null
  copy: SidebarCopy
  onRefresh: () => void
  onSelectThread: (threadId: string) => void
  onSelectMode: (mode: SidebarMode) => void
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
  selectedMode,
  modeCopy,
  loading = false,
  error = null,
  copy,
  onRefresh,
  onSelectThread,
  onSelectMode,
  onCreateThread,
  onRenameThread,
  onArchiveThread,
  onOpenSettings,
}: Props) {
  const [threadMenuId, setThreadMenuId] = useState<string | null>(null)
  const [editingThreadId, setEditingThreadId] = useState<string | null>(null)
  const [renameDraft, setRenameDraft] = useState('')
  const footerItems = createSidebarFooterItems()
  const modeMenuItems = createSidebarModeMenuItems(selectedMode, copy)
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
    if (window.confirm(copy.archiveThread)) onArchiveThread(thread.id)
  }

  if (collapsed) return null

  return (
    <aside className="sidebar">
      <div className="sidebar-mode-row" aria-label="Workspace mode">
        <button className={selectedMode === 'chat' ? 'selected' : undefined} type="button" onClick={() => onSelectMode('chat')} title={modeCopy.chat}>
          <MessageCircle size={17} />
          <span>{modeCopy.chat}</span>
        </button>
        <button className={selectedMode === 'work' ? 'selected' : undefined} type="button" onClick={() => onSelectMode('work')} title={modeCopy.work}>
          <BriefcaseBusiness size={17} />
          <span>{modeCopy.work}</span>
        </button>
      </div>

      <div className="sidebar-section nav-stack compact-nav">
        {modeMenuItems.filter((item) => item.id !== 'new-chat').map((item) => (
          <button className="nav-item" key={item.id} onClick={item.action === 'create-thread' ? onCreateThread : undefined}>
            {item.id === 'projects' && <FolderKanban size={15} />}
            {item.id === 'scheduled' && <Clock3 size={15} />}
            {item.label}
          </button>
        ))}
      </div>

      <div className="sidebar-divider" />

      <div className="sidebar-section thread-list-section">
        <div className="section-label">
          <span>{copy.threads}</span>
          <button className="thread-create-button" type="button" aria-label={copy.newChat} title={copy.newChat} onClick={onCreateThread}>
            <MessageSquarePlus size={14} />
          </button>
        </div>
        <div className="thread-list">
          {loading && <div className="thread-state">{copy.loadingThreads}</div>}
          {error && <div className="thread-state error"><span>{error}</span><button type="button" onClick={onRefresh}>{copy.retry}</button></div>}
          {threads.length === 0 && <div className="thread-state empty">{copy.emptyThreads(selectedMode)}</div>}
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
                  <button
                    className={thread.id === selectedThreadId ? 'thread-card selected' : 'thread-card'}
                    onClick={() => onSelectThread(thread.id)}
                  >
                    <span className={`run-dot ${thread.runStatus}`} aria-label={runStatusCopy[thread.runStatus]} title={runStatusCopy[thread.runStatus]} />
                    <span className="thread-title">{thread.title}</span>
                  </button>
                  <button
                    className="thread-action"
                    aria-expanded={threadMenuId === thread.id}
                    aria-label={copy.threadActions}
                    onClick={(event) => {
                      event.stopPropagation()
                      setThreadMenuId((current) => current === thread.id ? null : thread.id)
                    }}
                  >
                    <MoreHorizontal size={14} />
                  </button>
                </>
              )}
              {threadMenuId === thread.id && (
                <div className="thread-menu">
                  <button type="button" onClick={() => renameThread(thread)}><Pencil size={13} /> <span>{copy.renameThread}</span></button>
                  <button className="danger" type="button" onClick={() => deleteThread(thread)}><Trash2 size={13} /> <span>{copy.archiveThread}</span></button>
                </div>
              )}
            </div>
          ))}
        </div>
      </div>
      <div className="sidebar-footer">
        {footerItems.map((item) => (
          <button className="sidebar-settings-button" key={item.id} type="button" onClick={onOpenSettings}>
            <Settings size={16} />
            <span>{copy.settings}</span>
          </button>
        ))}
      </div>
    </aside>
  )
}
