import { useState } from 'react'
import { Archive, Check, Clock3, FolderKanban, MessageSquarePlus, MonitorCog, RefreshCw, Settings } from 'lucide-react'
import type { Thread } from '../domain'
import { createSettingsMenuItems, type SettingsMenuItemId } from './settingsMenuItems'
import { createSidebarModeMenuItems, type SidebarMode } from './sidebarModeMenuItems'

type Props = {
  collapsed: boolean
  threads: Thread[]
  selectedThreadId: string
  selectedMode: SidebarMode
  theme: 'dark' | 'light'
  loading?: boolean
  error?: string | null
  onRefresh: () => void
  onSelectThread: (threadId: string) => void
  onCreateThread: () => void
  onRenameThread: (threadId: string, title: string) => void
  onArchiveThread: (threadId: string) => void
  onToggleTheme: () => void
}

export function ThreadSidebar({
  collapsed,
  threads,
  selectedThreadId,
  selectedMode,
  theme,
  loading = false,
  error = null,
  onRefresh,
  onSelectThread,
  onCreateThread,
  onRenameThread,
  onArchiveThread,
  onToggleTheme,
}: Props) {
  const [settingsOpen, setSettingsOpen] = useState(false)
  const settingsItems = createSettingsMenuItems(theme)
  const modeMenuItems = createSidebarModeMenuItems(selectedMode)

  const handleSettingsAction = (itemId: SettingsMenuItemId) => {
    if (itemId === 'theme') onToggleTheme()
    if (itemId === 'update') onRefresh()
  }

  if (collapsed) return null

  return (
    <aside className="sidebar">
      <div className="sidebar-section nav-stack compact-nav">
        {modeMenuItems.map((item) => (
          <button className="nav-item" key={item.id} onClick={item.action === 'create-thread' ? onCreateThread : undefined}>
            {item.id === 'new-chat' && <MessageSquarePlus size={15} />}
            {item.id === 'projects' && <FolderKanban size={15} />}
            {item.id === 'scheduled' && <Clock3 size={15} />}
            {item.label}
          </button>
        ))}
      </div>

      <div className="sidebar-divider" />

      <div className="sidebar-section">
        <div className="section-label">
          <span>Threads</span>
        </div>
        <div className="thread-list">
          {loading && <div className="thread-state">Loading threads</div>}
          {error && <div className="thread-state error"><span>{error}</span><button type="button" onClick={onRefresh}>Retry</button></div>}
          {threads.length === 0 && <div className="thread-state empty">No {selectedMode} threads</div>}
          {threads.map((thread) => (
            <div className="thread-row" key={thread.id}>
              <button
                className={thread.id === selectedThreadId ? 'thread-card selected' : 'thread-card'}
                onClick={() => onSelectThread(thread.id)}
              >
                <span className={`run-dot ${thread.runStatus}`} />
                <span className="thread-title" onDoubleClick={(event) => {
                  event.stopPropagation()
                  const title = window.prompt('Rename thread', thread.title)
                  if (title) onRenameThread(thread.id, title)
                }}>{thread.title}</span>
              </button>
              <button className="thread-action" aria-label="Archive thread" onClick={() => onArchiveThread(thread.id)}><Archive size={12} /></button>
            </div>
          ))}
        </div>
      </div>

      <div className="sidebar-footer">
        <div className="settings-wrap">
          {settingsOpen && (
            <div className="settings-popover">
              {settingsItems.map((item) => (
                <button
                  className={item.id === 'theme' ? 'settings-menu-row theme-row' : 'settings-menu-row'}
                  key={item.id}
                  onClick={() => handleSettingsAction(item.id)}
                >
                  <span className="settings-menu-label">
                    {item.id === 'settings' && <Settings size={15} />}
                    {item.id === 'theme' && <MonitorCog size={15} />}
                    {item.id === 'update' && <RefreshCw size={15} />}
                    {item.label}
                  </span>
                  {item.id === 'settings' && <span className="settings-menu-value">›</span>}
                  {item.id === 'theme' && (
                    <span className="theme-segment" aria-label="Theme mode">
                      <span className={theme === 'light' ? 'selected' : ''}>Light</span>
                      <span className={theme === 'dark' ? 'selected' : ''}>Dark</span>
                    </span>
                  )}
                  {item.id === 'update' && (
                    <span className="settings-menu-value">
                      {item.value}
                      <Check size={13} />
                    </span>
                  )}
                </button>
              ))}
            </div>
          )}
          <button
            className={settingsOpen ? 'settings-entry open' : 'settings-entry'}
            aria-expanded={settingsOpen}
            aria-label="Settings"
            onClick={() => setSettingsOpen((value) => !value)}
          >
            <Settings size={15} /> Settings
          </button>
        </div>
      </div>
    </aside>
  )
}
