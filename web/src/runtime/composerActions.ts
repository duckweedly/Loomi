import type { Message, Run } from '../domain'
import { shouldBlockRuntimeSubmit } from '../state'

export type ComposerActionsInput = {
  threadSelected: boolean
  text: string
  run: Run | null
  messages: Message[]
  providerUnavailable?: boolean
}

export function deriveComposerActions(input: ComposerActionsInput) {
  const hasText = input.text.trim().length > 0
  const activeRun = shouldBlockRuntimeSubmit(input.run)
  const completedAssistant = input.messages.some((message) => message.role === 'assistant' && Boolean(message.content.trim()))
  const failedRun = input.run?.status === 'failed'

  if (input.providerUnavailable) {
    return {
      canSend: false,
      canContinue: false,
      canStop: activeRun,
      canRetry: false,
      canRegenerate: false,
      disabledReason: 'provider-unavailable' as const,
    }
  }

  const disabledReason = activeRun
    ? input.run?.status === 'retrying'
      ? 'retry-already-scheduled'
      : input.run?.status === 'recovering'
        ? 'recovery-in-progress'
        : 'generation-in-progress'
    : hasText && input.threadSelected
        ? null
        : 'no-valid-prompt'

  return {
    canSend: input.threadSelected && hasText && !activeRun,
    canContinue: input.threadSelected && hasText && !activeRun,
    canStop: activeRun,
    canRetry: input.threadSelected && failedRun && !activeRun,
    canRegenerate: input.threadSelected && completedAssistant && !activeRun,
    disabledReason,
  }
}
