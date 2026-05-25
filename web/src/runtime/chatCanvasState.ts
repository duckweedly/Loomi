import type { BackendCapabilityState, ChatCanvasState, Run } from '../domain'

export type ChatCanvasStateInput = {
  loading: boolean
  error?: string | null
  backendCapability?: BackendCapabilityState
  backendUnavailableAttempted?: boolean
  selectedThreadId?: string | null
  messageCount: number
  run?: Run | null
}

export function deriveChatCanvasState(input: ChatCanvasStateInput): ChatCanvasState {
  if (input.loading) return 'loading'
  if (input.error) return 'error'
  if (input.backendCapability === 'unavailable' && input.backendUnavailableAttempted) return 'backend-unavailable'
  if (!input.selectedThreadId) return 'no-thread'
  if (input.run) {
    const draftStatus = input.run.assistantDraft?.status
    if (draftStatus === 'recovering' || input.run.status === 'recovering') return 'recovering'
    if (draftStatus === 'stopping' || input.run.status === 'stopping') return 'stopping'
    if (draftStatus === 'stopped' || input.run.status === 'stopped') return 'stopped'
    if (draftStatus === 'failed' || input.run.status === 'failed') return 'failed'
    if (input.run.status === 'pending' || input.run.status === 'queued' || input.run.status === 'blocked_on_tool_approval' || draftStatus === 'pending' || draftStatus === 'queued') return 'waiting-run'
    if (input.run.status === 'running') return input.run.events.length === 0 && !input.run.assistantDraft?.content ? 'waiting-run' : 'running'
    if (input.run.status === 'completed' && input.run.events.length > 0) return 'completed'
  }
  if (input.messageCount === 0) return 'empty-thread'
  return 'history'
}
