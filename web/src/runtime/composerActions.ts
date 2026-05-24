import type { Message, Run } from '../domain'
import { shouldBlockRuntimeSubmit } from '../state'

export type ComposerActionsInput = {
  threadSelected: boolean
  text: string
  run: Run | null
  messages: Message[]
}

export function deriveComposerActions(input: ComposerActionsInput) {
  const hasText = input.text.trim().length > 0
  const activeRun = shouldBlockRuntimeSubmit(input.run)
  const completedAssistant = input.messages.some((message) => message.role === 'assistant' && Boolean(message.content.trim()))
  const failedRun = input.run?.status === 'failed'

  return {
    canSend: input.threadSelected && hasText && !activeRun,
    canContinue: input.threadSelected && hasText && !activeRun,
    canStop: input.run?.status === 'queued' || input.run?.status === 'running' || input.run?.status === 'retrying' || input.run?.status === 'recovering',
    canRetry: input.threadSelected && failedRun && !activeRun,
    canRegenerate: input.threadSelected && completedAssistant && !activeRun,
  }
}
