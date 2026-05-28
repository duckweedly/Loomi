import type { Message, Run } from '../domain'
import type { Locale } from '../i18n'
import type { RightPanelItemId } from '../rightPanelItems'
import { RightToolDrawer } from './RightToolDrawer'

type Props = {
  run: Run | null
  messages?: Message[]
  runDetailsOpen?: boolean
  rightToolsOpen: boolean
  selectedPanelId: RightPanelItemId
  selectedArtifactId?: string
  onStopRun?: () => void
  locale: Locale
  selectedThreadId?: string
}

export function RunTimeline({
  run,
  messages = [],
  rightToolsOpen,
  selectedPanelId,
  selectedArtifactId,
  locale,
  selectedThreadId,
}: Props) {
  const selectedThreadRun = !selectedThreadId || run?.threadId === selectedThreadId ? run : null

  return (
    <RightToolDrawer open={rightToolsOpen} selectedPanelId={selectedPanelId} selectedArtifactId={selectedArtifactId} run={selectedThreadRun} messages={messages} locale={locale} />
  )
}
