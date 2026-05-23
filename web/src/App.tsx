import { type CSSProperties, type PointerEvent } from 'react'
import { ConfigProvider, ThemeProvider } from '@lobehub/ui'
import { AlertCircle, PanelLeft, PanelRight, Search } from 'lucide-react'
import { motion } from 'motion/react'
import { ChatCanvas } from './components/ChatCanvas'
import { RunTimeline } from './components/RunTimeline'
import { ThreadSidebar } from './components/ThreadSidebar'
import { useWorkspaceState } from './state'
import { useWorkspaceShellState } from './useWorkspaceShellState'

export default function App() {
  const shell = useWorkspaceShellState()
  const {
    threads,
    selectedThread,
    selectedThreadId,
    messages,
    run,
    loading,
    refresh,
    selectThread,
    sendMessage,
  } = useWorkspaceState()

  const workspaceStyle = { '--sidebar-width': `${shell.sidebarWidth}px` } as CSSProperties
  const workspaceClass = [
    'workspace-grid',
    shell.sidebarCollapsed ? 'sidebar-collapsed' : '',
    shell.rightPanelOpen ? 'right-tools-open' : '',
  ].filter(Boolean).join(' ')

  const handleSidebarResize = (event: PointerEvent<HTMLDivElement>) => {
    const startX = event.clientX
    const startWidth = shell.sidebarWidth
    event.currentTarget.setPointerCapture(event.pointerId)

    const handlePointerMove = (moveEvent: globalThis.PointerEvent) => {
      shell.setSidebarWidth(Math.min(380, Math.max(248, startWidth + moveEvent.clientX - startX)))
    }

    const handlePointerUp = () => {
      window.removeEventListener('pointermove', handlePointerMove)
      window.removeEventListener('pointerup', handlePointerUp)
    }

    window.addEventListener('pointermove', handlePointerMove)
    window.addEventListener('pointerup', handlePointerUp)
  }

  return (
    <ConfigProvider motion={motion}>
      <ThemeProvider
        appearance={shell.theme}
        customTheme={{
          primaryColor: 'purple',
          neutralColor: 'slate',
        }}
      >
        <div className="app-shell" data-theme={shell.theme}>
          <main className={workspaceClass} style={workspaceStyle}>
            {!shell.sidebarCollapsed && (
              <aside className="sidebar-shell glass-panel">
                <div className="sidebar-titlebar">
                  <button className="titlebar-button" aria-label="Collapse sidebar" onClick={() => shell.setSidebarCollapsed(true)}>
                    <PanelLeft size={15} strokeWidth={1.7} />
                  </button>
                  <button className="titlebar-button" aria-label="Search">
                    <Search size={14} strokeWidth={1.65} />
                  </button>
                </div>
                <ThreadSidebar
                  collapsed={shell.sidebarCollapsed}
                  threads={threads}
                  selectedThreadId={selectedThreadId}
                  run={run}
                  theme={shell.theme}
                  onRefresh={() => void refresh()}
                  onSelectThread={selectThread}
                  onToggleTheme={shell.toggleTheme}
                />
              </aside>
            )}
            {!shell.sidebarCollapsed && <div className="sidebar-resizer" role="separator" aria-orientation="vertical" onPointerDown={handleSidebarResize} />}
            <section className="main-region">
              <header className="main-titlebar">
                <div className="titlebar-left">
                  {shell.sidebarCollapsed && (
                    <button className="titlebar-button" aria-label="Open sidebar" onClick={() => shell.setSidebarCollapsed(false)}>
                      <PanelRight size={15} strokeWidth={1.7} />
                    </button>
                  )}
                </div>
                <div className="titlebar-center mode-tabs">
                  <button
                    className={selectedThread?.mode === 'chat' ? 'selected' : undefined}
                    onClick={() => {
                      const threadId = threads.find((thread) => thread.mode === 'chat')?.id
                      if (threadId) selectThread(threadId)
                    }}
                  >
                    Chat
                  </button>
                  <button
                    className={selectedThread?.mode === 'work' ? 'selected' : undefined}
                    onClick={() => {
                      const threadId = threads.find((thread) => thread.mode === 'work')?.id
                      if (threadId) selectThread(threadId)
                    }}
                  >
                    Work
                  </button>
                </div>
                <div className="titlebar-right">
                  <button
                    className="titlebar-button"
                    aria-label="Open run details"
                    onClick={shell.toggleRunDetails}
                  >
                    <AlertCircle size={15} strokeWidth={1.7} />
                  </button>
                  <button
                    className="titlebar-button"
                    aria-label="Open right tools"
                    onClick={shell.toggleRightPanelMenu}
                  >
                    <PanelRight size={15} strokeWidth={1.7} />
                  </button>
                </div>
              </header>
              <ChatCanvas
                sidebarCollapsed={shell.sidebarCollapsed}
                thread={selectedThread}
                messages={messages}
                run={run}
                loading={loading}
                onSendMessage={(content) => void sendMessage(content)}
              />
            </section>
            <RunTimeline
              run={run}
              runDetailsOpen={shell.runDetailsOpen}
              rightPanelMenuOpen={shell.rightPanelMenuOpen}
              rightToolsOpen={shell.rightPanelOpen}
              artifactOpen={shell.artifactOpen}
              selectedPanelId={shell.selectedRightPanelId}
              onSelectPanel={shell.openRightPanel}
              onCloseRunDetails={shell.closeRunDetails}
              onCloseRightTools={shell.closeRightPanel}
              onOpenArtifact={shell.openArtifact}
              onCloseArtifact={shell.closeArtifact}
            />
          </main>
        </div>
      </ThemeProvider>
    </ConfigProvider>
  )
}
