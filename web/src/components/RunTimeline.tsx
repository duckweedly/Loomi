import type { Run } from '../domain'
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
}

export function RunTimeline({
  run,
  runDetailsOpen,
  rightPanelMenuOpen,
  rightToolsOpen,
  selectedPanelId,
  onSelectPanel,
  onOpenArtifact,
}: Props) {
  return (
    <>
      <RunRail run={run} open={runDetailsOpen} onOpenArtifact={onOpenArtifact} />
      <RightPanelMenu open={rightPanelMenuOpen} selectedPanelId={selectedPanelId} onSelectPanel={onSelectPanel} />
      <RightToolDrawer open={rightToolsOpen} selectedPanelId={selectedPanelId} />
    </>
  )
}
