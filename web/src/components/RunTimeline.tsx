import type { Run, RuntimeScriptId } from '../domain'
import type { Locale } from '../i18n'
import type { BackendCapabilityStatus } from '../runtime/backendCapabilityStatus'
import type { RightPanelItemId } from '../rightPanelItems'
import { RightPanelMenu } from './RightPanelMenu'
import { RightToolDrawer } from './RightToolDrawer'
import { RunRail } from './RunRail'

type Props = {
  run: Run | null
  runDetailsOpen: boolean
  rightPanelMenuOpen: boolean
  rightToolsOpen: boolean
  selectedPanelId: RightPanelItemId
  onSelectPanel: (panelId: RightPanelItemId) => void
  onOpenArtifact: () => void
  onStopRun?: () => void
  selectedRuntimeScript?: RuntimeScriptId
  capabilityStatus?: BackendCapabilityStatus
  locale: Locale
  onSelectRuntimeScript?: (scriptId: RuntimeScriptId) => void
  selectedThreadId?: string
}

export function RunTimeline({
  run,
  runDetailsOpen,
  rightPanelMenuOpen,
  rightToolsOpen,
  selectedPanelId,
  onSelectPanel,
  onOpenArtifact,
  onStopRun,
  selectedRuntimeScript,
  capabilityStatus,
  locale,
  onSelectRuntimeScript,
  selectedThreadId,
}: Props) {
  const selectedThreadRun = !selectedThreadId || run?.threadId === selectedThreadId ? run : null

  return (
    <>
      <RunRail run={run} open={runDetailsOpen} onOpenArtifact={onOpenArtifact} onStopRun={onStopRun} selectedRuntimeScript={selectedRuntimeScript} capabilityStatus={capabilityStatus} locale={locale} onSelectRuntimeScript={onSelectRuntimeScript} />
      <RightPanelMenu open={rightPanelMenuOpen} selectedPanelId={selectedPanelId} onSelectPanel={onSelectPanel} />
      <RightToolDrawer open={rightToolsOpen} selectedPanelId={selectedPanelId} run={selectedThreadRun} locale={locale} />
    </>
  )
}
