import type { Run, RuntimeScriptId } from '../domain'
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
  onSelectRuntimeScript?: (scriptId: RuntimeScriptId) => void
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
  onSelectRuntimeScript,
}: Props) {
  return (
    <>
      <RunRail run={run} open={runDetailsOpen} onOpenArtifact={onOpenArtifact} onStopRun={onStopRun} selectedRuntimeScript={selectedRuntimeScript} capabilityStatus={capabilityStatus} onSelectRuntimeScript={onSelectRuntimeScript} />
      <RightPanelMenu open={rightPanelMenuOpen} selectedPanelId={selectedPanelId} onSelectPanel={onSelectPanel} />
      <RightToolDrawer open={rightToolsOpen} selectedPanelId={selectedPanelId} />
    </>
  )
}
