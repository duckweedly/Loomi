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
      <RunRail run={selectedThreadRun} open={runDetailsOpen} onStopRun={onStopRun} selectedRuntimeScript={selectedRuntimeScript} capabilityStatus={capabilityStatus} locale={locale} onSelectRuntimeScript={onSelectRuntimeScript} />
      <RightPanelMenu open={rightPanelMenuOpen} selectedPanelId={selectedPanelId} onSelectPanel={onSelectPanel} locale={locale} />
      <RightToolDrawer open={rightToolsOpen} selectedPanelId={selectedPanelId} run={selectedThreadRun} locale={locale} />
    </>
  )
}
