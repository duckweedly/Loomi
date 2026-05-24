import { type CSSProperties, type PointerEvent } from 'react'
import { ConfigProvider, ThemeProvider } from '@lobehub/ui'
import { AlertCircle, PanelLeft, PanelRight, Search } from 'lucide-react'
import { motion } from 'motion/react'
import { ChatCanvas } from './components/ChatCanvas'
import { RunTimeline } from './components/RunTimeline'
import { SettingsView } from './components/SettingsView'
import { ThreadSidebar } from './components/ThreadSidebar'
import { getDictionary } from './i18n'
import { deriveBackendCapabilityStatus } from './runtime/backendCapabilityStatus'
import { useWorkspaceState } from './state'
import { filterThreadsByMode } from './threadFilters'
import { useWorkspaceShellState } from './useWorkspaceShellState'

export default function App() {
  const shell = useWorkspaceShellState()
  const {
    threads,
    selectedThread,
    selectedThreadId,
    messages,
    run,
    streamState,
    loading,
    error,
    dataSourceMode,
    backendCapability,
    backendUnavailableAttempted,
    capabilitySignals,
    selectedRuntimeScript,
    selectRuntimeScript,
    refresh,
    selectThread,
    createThread,
    renameThread,
    archiveThread,
    sendMessage,
    stopRun,
    retryRun,
    regenerateRun,
    providerCapabilities,
  } = useWorkspaceState(shell.defaultWorkspaceMode)

  const dictionary = getDictionary(shell.locale)
  const selectedMode = selectedThread?.mode ?? 'chat'
  const visibleThreads = filterThreadsByMode(threads, selectedMode)
  const capabilityStatus = deriveBackendCapabilityStatus({
    dataSourceMode,
    runtimeSource: run?.context === 'model_gateway' ? 'model_gateway' : 'local_simulated',
    backendUnavailable: backendCapability === 'unavailable' || backendUnavailableAttempted || capabilitySignals.backendUnavailable,
    modelSetupMissing: capabilitySignals.modelSetupMissing,
    providerUnavailable: capabilitySignals.providerUnavailable,
    activeRun: Boolean(run && (run.status === 'pending' || run.status === 'queued' || run.status === 'running' || run.status === 'retrying' || run.status === 'recovering' || run.status === 'stopping')),
    streamDisconnected: Boolean(run && (run.status === 'pending' || run.status === 'queued' || run.status === 'running' || run.status === 'retrying' || run.status === 'recovering' || run.status === 'stopping') && (capabilitySignals.streamDisconnected || streamState === 'recoverable_error')),
    runRecovering: run?.status === 'recovering' || run?.assistantDraft?.status === 'recovering',
  })
  const workspaceStyle = { '--sidebar-width': `${shell.sidebarWidth}px` } as CSSProperties
  const workspaceClass = [
    'workspace-grid',
    shell.sidebarCollapsed ? 'sidebar-collapsed' : '',
    shell.rightPanelOpen && !shell.settingsOpen ? 'right-tools-open' : '',
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
                  <button className="titlebar-button" aria-label={dictionary.app.collapseSidebar} onClick={() => shell.setSidebarCollapsed(true)}>
                    <PanelLeft size={15} strokeWidth={1.7} />
                  </button>
                  <button className="titlebar-button" aria-label={dictionary.app.search}>
                    <Search size={14} strokeWidth={1.65} />
                  </button>
                </div>
                <ThreadSidebar
                  collapsed={shell.sidebarCollapsed}
                  threads={visibleThreads}
                  selectedThreadId={selectedThreadId}
                  selectedMode={selectedMode}
                  theme={shell.theme}
                  loading={loading}
                  error={error}
                  copy={dictionary.sidebar}
                  onRefresh={() => void refresh()}
                  onSelectThread={selectThread}
                  onCreateThread={() => void createThread()}
                  onRenameThread={(threadId, title) => void renameThread(threadId, title)}
                  onArchiveThread={(threadId) => void archiveThread(threadId)}
                  onToggleTheme={shell.toggleTheme}
                  onOpenSettings={shell.openSettings}
                />
              </aside>
            )}
            {!shell.sidebarCollapsed && <div className="sidebar-resizer" role="separator" aria-orientation="vertical" onPointerDown={handleSidebarResize} />}
            <section className="main-region">
              <header className="main-titlebar">
                <div className="titlebar-left">
                  {shell.sidebarCollapsed && (
                    <button className="titlebar-button" aria-label={dictionary.app.openSidebar} onClick={() => shell.setSidebarCollapsed(false)}>
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
                    {dictionary.app.chat}
                  </button>
                  <button
                    className={selectedThread?.mode === 'work' ? 'selected' : undefined}
                    onClick={() => {
                      const threadId = threads.find((thread) => thread.mode === 'work')?.id
                      if (threadId) selectThread(threadId)
                    }}
                  >
                    {dictionary.app.work}
                  </button>
                </div>
                <div className="titlebar-right">
                  <button
                    className="titlebar-button"
                    aria-label={dictionary.app.openRunDetails}
                    onClick={shell.toggleRunDetails}
                  >
                    <AlertCircle size={15} strokeWidth={1.7} />
                  </button>
                  <button
                    className="titlebar-button"
                    aria-label={dictionary.app.openRightTools}
                    onClick={shell.toggleRightPanelMenu}
                  >
                    <PanelRight size={15} strokeWidth={1.7} />
                  </button>
                </div>
              </header>
              {shell.settingsOpen ? (
                <SettingsView
                  locale={shell.locale}
                  selectedCategoryId={shell.settingsCategoryId}
                  defaultWorkspaceMode={shell.defaultWorkspaceMode}
                  selectedRuntimeScript={selectedRuntimeScript}
                  dataSourceMode={dataSourceMode}
                  backendCapability={backendCapability}
                  streamState={streamState}
                  selectedThreadTitle={selectedThread?.title}
                  selectedRunStatus={run?.status}
                  providerCapabilities={providerCapabilities}
                  providerDraftSettings={shell.providerDraftSettings}
                  onSelectLocale={shell.setLocale}
                  onSelectCategory={shell.setSettingsCategory}
                  onSelectDefaultWorkspaceMode={shell.setDefaultWorkspaceMode}
                  onSelectRuntimeScript={selectRuntimeScript}
                  onProviderDraftSettingsChange={shell.setProviderDraftSettings}
                  onBack={shell.closeSettings}
                />
              ) : (
                <ChatCanvas
                  sidebarCollapsed={shell.sidebarCollapsed}
                  thread={selectedThread}
                  messages={messages}
                  run={run}
                  loading={loading}
                  error={error}
                  dataSourceMode={dataSourceMode}
                  streamState={streamState}
                  backendCapability={backendCapability}
                  backendUnavailableAttempted={backendUnavailableAttempted}
                  capabilitySignals={capabilitySignals}
                  onSendMessage={(content) => void sendMessage(content)}
                  onStopRun={() => void stopRun()}
                  onRetryRun={retryRun}
                  onRegenerateRun={regenerateRun}
                  locale={shell.locale}
                />
              )}
            </section>
            <RunTimeline
              run={run}
              runDetailsOpen={!shell.settingsOpen && shell.runDetailsOpen}
              rightPanelMenuOpen={!shell.settingsOpen && shell.rightPanelMenuOpen}
              rightToolsOpen={!shell.settingsOpen && shell.rightPanelOpen}
              selectedPanelId={shell.selectedRightPanelId}
              onSelectPanel={shell.openRightPanel}
              onOpenArtifact={shell.openArtifact}
              onStopRun={() => void stopRun()}
              selectedRuntimeScript={selectedRuntimeScript}
              capabilityStatus={capabilityStatus}
              onSelectRuntimeScript={dataSourceMode === 'mock' ? selectRuntimeScript : undefined}
            />
          </main>
        </div>
      </ThemeProvider>
    </ConfigProvider>
  )
}
