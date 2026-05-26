import { type CSSProperties, type PointerEvent, useCallback } from 'react'
import { ConfigProvider, ThemeProvider } from '@lobehub/ui'
import { AlertCircle, PanelLeft, PanelRight, SquarePen } from 'lucide-react'
import { motion } from 'motion/react'
import { ChatCanvas } from './components/ChatCanvas'
import { RunTimeline } from './components/RunTimeline'
import { SettingsView } from './components/SettingsView'
import { ThreadSidebar } from './components/ThreadSidebar'
import { getDictionary } from './i18n'
import { deriveBackendCapabilityStatus, shouldShowProviderUnavailableWarning } from './runtime/backendCapabilityStatus'
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
    personas,
    installedSkills,
    skillsLoading,
    skillsError,
    selectedPersonaId,
    selectRuntimeScript,
    setSelectedPersonaId,
    refresh,
    selectThread,
    createThread,
    renameThread,
    archiveThread,
    sendMessage,
    stopRun,
    approveToolCall,
    denyToolCall,
    retryRun,
    regenerateRun,
    providerCapabilities,
    toolCatalog,
    webSearchConfig,
    webSearchSaveResult,
    workspaceRootConfig,
    workspaceRootSaveResult,
    mcpServers,
    mcpActionResult,
    localProviderDetections,
    localProviderDetectionError,
    providerCheckResults,
    providerSaveResult,
    memoryEntries,
    memoryQuery,
    memoryFilters,
    memoryLoading,
    memoryError,
    memoryDetail,
    memoryDetailLoading,
    memoryDetailError,
    memoryAuditItems,
    memoryAuditLoading,
    memoryAuditError,
    pendingDeleteMemoryEntry,
    checkProvider,
    detectLocalProviders,
    enableLocalProvider,
    disableLocalProvider,
    saveProvider,
    saveWebSearchKeys,
    chooseWorkspaceFolder,
    saveMCPServer,
    deleteMCPServer,
    discoverMCPServer,
    setMemorySearchQuery,
    updateMemoryFilters,
    openMemoryDetail,
    closeMemoryDetail,
    requestDeleteMemoryEntry,
    cancelDeleteMemoryEntry,
    deleteMemoryEntry,
  } = useWorkspaceState(shell.defaultWorkspaceMode)

  const dictionary = getDictionary(shell.locale)
  const selectedMode = selectedThread?.mode ?? 'chat'
  const visibleThreads = filterThreadsByMode(threads, selectedMode)
  const providerUnavailableBeforeSend = shouldShowProviderUnavailableWarning(dataSourceMode, providerCapabilities)
  const capabilityStatus = deriveBackendCapabilityStatus({
    dataSourceMode,
    runtimeSource: run?.context === 'model_gateway' ? 'model_gateway' : 'local_simulated',
    backendUnavailable: backendCapability === 'unavailable' || backendUnavailableAttempted || capabilitySignals.backendUnavailable,
    modelSetupMissing: capabilitySignals.modelSetupMissing,
    providerUnavailable: providerUnavailableBeforeSend,
    activeRun: Boolean(run && (run.status === 'pending' || run.status === 'queued' || run.status === 'running' || run.status === 'retrying' || run.status === 'recovering' || run.status === 'blocked_on_tool_approval' || run.status === 'stopping')),
    streamDisconnected: Boolean(run && (run.status === 'pending' || run.status === 'queued' || run.status === 'running' || run.status === 'retrying' || run.status === 'recovering' || run.status === 'blocked_on_tool_approval' || run.status === 'stopping') && (capabilitySignals.streamDisconnected || streamState === 'recoverable_error')),
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

  const selectMode = useCallback((mode: 'chat' | 'work') => {
    const threadId = threads.find((thread) => thread.mode === mode)?.id
    if (threadId) {
      selectThread(threadId)
      return
    }
    void createThread(mode)
  }, [createThread, selectThread, threads])

  return (
    <ConfigProvider motion={motion}>
      <ThemeProvider
        appearance={shell.theme}
        customTheme={{
          primaryColor: 'purple',
          neutralColor: 'slate',
        }}
      >
        <div className="app-shell" data-runtime={document.documentElement.dataset.runtime} data-theme={shell.theme}>
          <main className={workspaceClass} style={workspaceStyle}>
            {!shell.sidebarCollapsed && (
              <aside className="sidebar-shell glass-panel">
                <div className="sidebar-titlebar">
                  <button className="titlebar-button" aria-label={dictionary.app.collapseSidebar} onClick={() => shell.setSidebarCollapsed(true)}>
                    <PanelLeft size={15} strokeWidth={1.7} />
                  </button>
                </div>
                <ThreadSidebar
                  collapsed={shell.sidebarCollapsed}
                  threads={visibleThreads}
                  selectedThreadId={selectedThreadId}
                  selectedMode={selectedMode}
                  modeCopy={{ chat: dictionary.app.chat, work: dictionary.app.work }}
                  theme={shell.theme}
                  loading={loading}
                  error={error}
                  copy={dictionary.sidebar}
                  onRefresh={() => void refresh()}
                  onSelectThread={selectThread}
                  onSelectMode={selectMode}
                  onCreateThread={() => void createThread(selectedMode)}
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
                    <>
                      <button className="titlebar-button" aria-label={dictionary.app.openSidebar} onClick={() => shell.setSidebarCollapsed(false)}>
                        <PanelRight size={15} strokeWidth={1.7} />
                      </button>
                      <button className="titlebar-button titlebar-create-thread" aria-label={selectedMode === 'work' ? dictionary.sidebar.newWork : dictionary.sidebar.newChat} onClick={() => void createThread(selectedMode)}>
                        <SquarePen size={15} strokeWidth={1.7} />
                      </button>
                    </>
                  )}
                </div>
                <div className="titlebar-center main-thread-title">
                  <span className="titlebar-brand-icon" aria-hidden="true" />
                  <span>{selectedThread?.title ?? 'Loomi'}</span>
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
                  theme={shell.theme}
                  backendCapability={backendCapability}
                  providerCapabilities={providerCapabilities}
                  workspaceRootConfig={workspaceRootConfig}
                  workspaceRootSaveResult={workspaceRootSaveResult}
                  personas={personas}
                  installedSkills={installedSkills}
                  skillsLoading={skillsLoading}
                  skillsError={skillsError}
                  toolCatalog={toolCatalog}
                  webSearchConfig={webSearchConfig}
                  webSearchSaveResult={webSearchSaveResult}
                  mcpServers={mcpServers}
                  mcpActionResult={mcpActionResult}
                  localProviderDetections={localProviderDetections}
                  localProviderDetectionError={localProviderDetectionError}
                  memoryEntries={memoryEntries}
                  memoryQuery={memoryQuery}
                  memoryFilters={memoryFilters}
                  memoryLoading={memoryLoading}
                  memoryError={memoryError}
                  memoryDetail={memoryDetail}
                  memoryDetailLoading={memoryDetailLoading}
                  memoryDetailError={memoryDetailError}
                  memoryAuditItems={memoryAuditItems}
                  memoryAuditLoading={memoryAuditLoading}
                  memoryAuditError={memoryAuditError}
                  pendingDeleteMemoryEntry={pendingDeleteMemoryEntry}
                  providerCheckResults={providerCheckResults}
                  providerSaveResult={providerSaveResult}
                  providerDraftSettings={shell.providerDraftSettings}
                  onSelectLocale={shell.setLocale}
                  onSelectCategory={shell.setSettingsCategory}
                  onSelectDefaultWorkspaceMode={shell.setDefaultWorkspaceMode}
                  onSelectTheme={(theme) => {
                    if (theme !== shell.theme) shell.toggleTheme()
                  }}
                  onProviderDraftSettingsChange={shell.setProviderDraftSettings}
                  onSaveProvider={(settings) => {
                    void saveProvider({ baseUrl: settings.baseUrl, model: settings.model, apiKey: settings.apiKey })
                    shell.setProviderDraftSettings({ ...settings, apiKey: '', apiKeySet: true })
                  }}
                  onSaveWebSearchKeys={(input) => void saveWebSearchKeys(input)}
                  onSaveMCPServer={(input) => void saveMCPServer(input)}
                  onDeleteMCPServer={(slug) => void deleteMCPServer(slug)}
                  onDiscoverMCPServer={(slug) => void discoverMCPServer(slug)}
                  onCheckProvider={(providerId) => void checkProvider(providerId)}
                  onDetectLocalProviders={() => void detectLocalProviders()}
                  onEnableLocalProvider={(providerId) => void enableLocalProvider(providerId)}
                  onDisableLocalProvider={(providerId) => void disableLocalProvider(providerId)}
                  onMemoryQueryChange={setMemorySearchQuery}
                  onMemoryFiltersChange={updateMemoryFilters}
                  onOpenMemoryDetail={(entry) => void openMemoryDetail(entry)}
                  onCloseMemoryDetail={closeMemoryDetail}
                  onRequestDeleteMemoryEntry={requestDeleteMemoryEntry}
                  onCancelDeleteMemoryEntry={cancelDeleteMemoryEntry}
                  onConfirmDeleteMemoryEntry={(entry) => void deleteMemoryEntry(entry)}
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
                  providerCapabilities={providerCapabilities}
                  personas={personas}
                  selectedPersonaId={selectedPersonaId}
                  onSelectPersona={setSelectedPersonaId}
                  onOpenProviderSettings={() => shell.openSettings('providers')}
                  onChooseWorkspaceFolder={() => void chooseWorkspaceFolder()}
                  onSendMessage={(content, options) => void sendMessage(content, options)}
                  onStopRun={() => void stopRun()}
                  onApproveToolCall={(toolCall) => approveToolCall(toolCall)}
                  onDenyToolCall={(toolCall) => denyToolCall(toolCall)}
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
              onStopRun={() => void stopRun()}
              selectedRuntimeScript={selectedRuntimeScript}
              capabilityStatus={capabilityStatus}
              locale={shell.locale}
              onSelectRuntimeScript={dataSourceMode === 'mock' ? selectRuntimeScript : undefined}
              selectedThreadId={selectedThreadId}
            />
          </main>
        </div>
      </ThemeProvider>
    </ConfigProvider>
  )
}
