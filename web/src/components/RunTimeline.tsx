import type { Run } from '../domain'
import type { RightPanelItemId } from '../rightPanelItems'
import { ArtifactDrawer } from './ArtifactDrawer'
import { RightPanelMenu } from './RightPanelMenu'
import { RightToolDrawer } from './RightToolDrawer'
import { RunRail } from './RunRail'

type Props = {
  run: Run | null
  runDetailsOpen: boolean
  rightPanelMenuOpen: boolean
  rightToolsOpen: boolean
  artifactOpen: boolean
  selectedPanelId: RightPanelItemId
  onSelectPanel: (panelId: RightPanelItemId) => void
  onCloseRunDetails: () => void
  onCloseRightTools: () => void
  onOpenArtifact: () => void
  onCloseArtifact: () => void
}

export function RunTimeline({
  run,
  runDetailsOpen,
  rightPanelMenuOpen,
  rightToolsOpen,
  artifactOpen,
  selectedPanelId,
  onSelectPanel,
  onCloseRunDetails,
  onCloseRightTools,
  onOpenArtifact,
  onCloseArtifact,
}: Props) {
  return (
    <>
      <RunRail run={run} open={runDetailsOpen} onClose={onCloseRunDetails} onOpenArtifact={onOpenArtifact} />
      <RightPanelMenu open={rightPanelMenuOpen} selectedPanelId={selectedPanelId} onSelectPanel={onSelectPanel} />
      <RightToolDrawer open={rightToolsOpen} selectedPanelId={selectedPanelId} onClose={onCloseRightTools} />
      <ArtifactDrawer open={artifactOpen} onClose={onCloseArtifact} />
    </>
  )
}
