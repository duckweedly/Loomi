import { type CSSProperties, type PointerEvent, useCallback } from 'react'
import { Button } from 'animal-island-ui'
import { ConfigProvider, ThemeProvider } from '@lobehub/ui'
import { PanelLeft, PanelRight, SquarePen } from 'lucide-react'
import { motion } from 'motion/react'
import { ChatCanvas } from './components/ChatCanvas'
import { RunTimeline } from './components/RunTimeline'
import { SettingsView } from './components/SettingsView'
import { ThreadSidebar } from './components/ThreadSidebar'
import { getDictionary } from './i18n'
import { useWorkspaceState } from './state'
import { sidebarMaxWidth, sidebarMinWidth, useWorkspaceShellState } from './useWorkspaceShellState'

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
    desktopReadiness,
    refreshDesktopReadiness,
    personas,
    installedSkills,
    skillsLoading,
    skillsError,
    selectedPersonaId,
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
    memoryWriteProposals,
    memoryProposalsLoading,
    memoryProposalsError,
    memoryProviderStatus,
    memoryErrors,
    memoryProviderSaveResult,
    memoryOverviewSnapshot,
    memoryImpressionSnapshot,
    memorySnapshotLoading,
    pendingDeleteMemoryEntry,
    approveMemoryWriteProposal,
    updateMemoryWriteProposal,
    denyMemoryWriteProposal,
    checkProvider,
    detectLocalProviders,
    enableLocalProvider,
    disableLocalProvider,
    saveProvider,
    saveWebSearchKeys,
    refreshMemoryProviderStatus,
    updateMemoryProvider,
    detectNowledgeMemoryProvider,
    detectOpenVikingMemoryProvider,
    rebuildMemoryOverviewSnapshot,
    rebuildMemoryImpressionSnapshot,
    getMemoryContent,
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
    createMemoryEntry,
    deleteMemoryEntry,
  } = useWorkspaceState(shell.defaultWorkspaceMode)

  const dictionary = getDictionary(shell.locale)
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
      shell.setSidebarWidth(Math.min(sidebarMaxWidth, Math.max(sidebarMinWidth, startWidth + moveEvent.clientX - startX)))
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
        <div className="app-shell" data-runtime={document.documentElement.dataset.runtime} data-theme={shell.theme}>
          <main className={workspaceClass} style={workspaceStyle}>
            {!shell.sidebarCollapsed && (
              <aside className="sidebar-shell glass-panel">
                <div className="sidebar-titlebar">
                  <Button className="titlebar-button" aria-label={dictionary.app.collapseSidebar} onClick={() => shell.setSidebarCollapsed(true)}>
                    <PanelLeft size={15} strokeWidth={1.7} />
                  </Button>
                </div>
                <ThreadSidebar
                  collapsed={shell.sidebarCollapsed}
                  threads={threads}
                  selectedThreadId={selectedThreadId}
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
                    <>
                      <Button className="titlebar-button" aria-label={dictionary.app.openSidebar} onClick={() => shell.setSidebarCollapsed(false)}>
                        <PanelRight size={15} strokeWidth={1.7} />
                      </Button>
                      <Button className="titlebar-button titlebar-create-thread" aria-label={dictionary.sidebar.newChat} onClick={() => void createThread()}>
                        <SquarePen size={15} strokeWidth={1.7} />
                      </Button>
                    </>
                  )}
                </div>
                <div className="titlebar-center main-thread-title">
                  <span className="titlebar-brand-icon" aria-hidden="true" />
                  <span>{selectedThread?.title ?? 'Loomi'}</span>
                </div>
                <div className="titlebar-right">
                  <Button
                    className="titlebar-button"
                    aria-label={dictionary.app.openRightTools}
                    onClick={shell.togglePreviewPanel}
                  >
                    <PanelRight size={15} strokeWidth={1.7} />
                  </Button>
                </div>
              </header>
              {shell.settingsOpen ? (
                <SettingsView
                  locale={shell.locale}
                  selectedCategoryId={shell.settingsCategoryId}
                  defaultWorkspaceMode={shell.defaultWorkspaceMode}
                  theme={shell.theme}
                  themePreference={shell.themePreference}
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
                  memoryWriteProposals={memoryWriteProposals}
                  memoryProposalsLoading={memoryProposalsLoading}
                  memoryProposalsError={memoryProposalsError}
                  memoryProviderStatus={memoryProviderStatus}
                  memoryErrors={memoryErrors}
                  memoryProviderSaveResult={memoryProviderSaveResult}
                  memoryOverviewSnapshot={memoryOverviewSnapshot}
                  memoryImpressionSnapshot={memoryImpressionSnapshot}
                  memorySnapshotLoading={memorySnapshotLoading}
                  pendingDeleteMemoryEntry={pendingDeleteMemoryEntry}
                  providerCheckResults={providerCheckResults}
                  providerSaveResult={providerSaveResult}
                  providerDraftSettings={shell.providerDraftSettings}
                  onSelectLocale={shell.setLocale}
                  onSelectCategory={shell.setSettingsCategory}
                  onSelectDefaultWorkspaceMode={shell.setDefaultWorkspaceMode}
                  onSelectTheme={shell.setThemePreference}
                  onProviderDraftSettingsChange={shell.setProviderDraftSettings}
                  onSaveProvider={(settings) => {
                    void saveProvider({ baseUrl: settings.baseUrl, model: settings.model, apiKey: settings.apiKey })
                    shell.setProviderDraftSettings({ ...settings, apiKey: '', apiKeySet: true })
                  }}
                  onSaveWebSearchKeys={(input) => void saveWebSearchKeys(input)}
                  onRefreshMemoryProviderStatus={() => void refreshMemoryProviderStatus()}
                  onUpdateMemoryProvider={(input) => void updateMemoryProvider(input)}
                  onDetectNowledgeMemoryProvider={detectNowledgeMemoryProvider}
                  onDetectOpenVikingMemoryProvider={detectOpenVikingMemoryProvider}
                  onRebuildMemoryOverviewSnapshot={() => void rebuildMemoryOverviewSnapshot()}
                  onRebuildMemoryImpressionSnapshot={() => void rebuildMemoryImpressionSnapshot()}
                  onGetMemoryContent={getMemoryContent}
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
                  onCreateMemoryEntry={(input) => void createMemoryEntry(input)}
                  onConfirmDeleteMemoryEntry={(entry) => void deleteMemoryEntry(entry)}
                  onApproveMemoryProposal={(proposal) => void approveMemoryWriteProposal(proposal)}
                  onUpdateMemoryProposal={(proposal, input) => void updateMemoryWriteProposal(proposal, input)}
                  onDenyMemoryProposal={(proposal) => void denyMemoryWriteProposal(proposal)}
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
                  workspaceRootConfig={workspaceRootConfig}
                  workspaceRootSaveResult={workspaceRootSaveResult}
                  desktopReadiness={desktopReadiness}
                  personas={personas}
                  selectedPersonaId={selectedPersonaId}
                  onSelectPersona={setSelectedPersonaId}
                  onRetryReadiness={() => void refreshDesktopReadiness()}
                  onDetectLocalProviders={() => void detectLocalProviders()}
                  onEnableLocalProvider={(providerId) => void enableLocalProvider(providerId)}
                  onOpenProviderSettings={() => shell.openSettings('providers')}
                  onOpenSkillsSettings={() => shell.openSettings('skill')}
                  onOpenConnectorsSettings={() => shell.openSettings('mcp')}
                  onOpenPluginsSettings={() => shell.openSettings('mcp')}
                  onChooseWorkspaceFolder={() => void chooseWorkspaceFolder()}
                  onSendMessage={(content, options) => void sendMessage(content, options)}
                  onStopRun={() => void stopRun()}
                  onApproveToolCall={(toolCall) => approveToolCall(toolCall)}
                  onDenyToolCall={(toolCall) => denyToolCall(toolCall)}
                  onOpenArtifact={(artifact) => shell.openArtifact(artifact.id)}
                  onRetryRun={retryRun}
                  onRegenerateRun={regenerateRun}
                  locale={shell.locale}
                />
              )}
            </section>
            <RunTimeline
              run={run}
              messages={messages}
              rightToolsOpen={!shell.settingsOpen && shell.rightPanelOpen}
              selectedPanelId={shell.selectedRightPanelId}
              selectedArtifactId={shell.previewArtifactId}
              locale={shell.locale}
              selectedThreadId={selectedThreadId}
            />
          </main>
        </div>
      </ThemeProvider>
    </ConfigProvider>
  )
}
