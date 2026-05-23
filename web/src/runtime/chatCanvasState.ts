import type { ChatCanvasState, Run } from '../domain'

export type ChatCanvasStateInput = {
  loading: boolean
  error?: string | null
  backendCapability?: 'available' | 'unavailable'
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
    if (input.run.status === 'pending') return 'waiting-run'
    if (input.run.status === 'running') return input.run.events.length === 0 ? 'waiting-run' : 'running'
    if (input.run.status === 'completed' && input.run.events.length > 0) return 'completed'
    if (input.run.status === 'failed' || input.run.status === 'stopped') return 'failed'
  }
  if (input.messageCount === 0) return 'empty-thread'
  return 'history'
}
