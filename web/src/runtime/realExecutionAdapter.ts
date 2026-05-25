import type { Run, RuntimeEvent, ToolCall, ToolCallApprovalStatus, ToolCallExecutionStatus, ToolCallLifecycle } from '../domain'
import type { ExecutionAdapter } from './executionAdapter'
import { isRuntimeTerminal } from './executionAdapter'

function delegated(method: string): never {
  throw new Error(`Use realApiClient.${method} for M4 run/event execution`)
}

export type RealRuntimeCapabilitySignal = {
  backendUnavailable?: boolean
  modelSetupMissing?: boolean
  providerUnavailable?: boolean
  streamDisconnected?: boolean
}

type ToolEventMapping = {
  status: ToolCallLifecycle
  approvalStatus?: ToolCallApprovalStatus
  executionStatus?: ToolCallExecutionStatus
}

function toolEventMapping(type: string): ToolEventMapping | null {
  switch (type) {
    case 'tool.call.requested':
      return { status: 'requested' }
    case 'tool.call.approval_required':
      return { status: 'approval_required', approvalStatus: 'required', executionStatus: 'blocked' }
    case 'tool.call.approved':
      return { status: 'approved', approvalStatus: 'approved' }
    case 'tool.call.denied':
      return { status: 'denied', approvalStatus: 'denied' }
    case 'tool.call.executing':
      return { status: 'executing', executionStatus: 'executing' }
    case 'tool.call.succeeded':
      return { status: 'succeeded', executionStatus: 'succeeded' }
    case 'tool.call.failed':
      return { status: 'failed', executionStatus: 'failed' }
    case 'tool.call.cancelled':
      return { status: 'cancelled', approvalStatus: 'cancelled', executionStatus: 'cancelled' }
    default:
      return null
  }
}

function metadataRecord(value: unknown): Record<string, unknown> | undefined {
  return typeof value === 'object' && value !== null && !Array.isArray(value) ? value as Record<string, unknown> : undefined
}

function metadataString(value: unknown): string | undefined {
  return typeof value === 'string' && value.trim() ? value : undefined
}

function metadataApprovalStatus(value: unknown): ToolCallApprovalStatus | undefined {
  return value === 'not_required' || value === 'required' || value === 'approved' || value === 'denied' || value === 'cancelled' ? value : undefined
}

function metadataExecutionStatus(value: unknown): ToolCallExecutionStatus | undefined {
  return value === 'not_started' || value === 'blocked' || value === 'executing' || value === 'succeeded' || value === 'failed' || value === 'cancelled' ? value : undefined
}

function applyToolEvent(toolCalls: ToolCall[] | undefined, event: RuntimeEvent, mapping: ToolEventMapping): ToolCall[] {
  const toolCallId = metadataString(event.metadata?.tool_call_id)
  const index = toolCalls?.findIndex((call) => call.toolCallId === toolCallId) ?? -1
  const current = index >= 0 ? toolCalls?.[index] : toolCalls?.[0]
  const next: ToolCall = {
    id: current?.id ?? event.id,
    toolCallId: toolCallId ?? current?.toolCallId,
    name: metadataString(event.metadata?.tool_name) ?? current?.name ?? event.label,
    status: mapping.status,
    approvalStatus: metadataApprovalStatus(event.metadata?.approval_status) ?? mapping.approvalStatus ?? current?.approvalStatus,
    executionStatus: metadataExecutionStatus(event.metadata?.execution_status) ?? mapping.executionStatus ?? current?.executionStatus,
    summary: event.detail,
    input: current?.input ?? '',
    output: event.content ?? current?.output ?? '',
    argumentsSummary: metadataRecord(event.metadata?.arguments_summary) ?? current?.argumentsSummary,
    resultSummary: metadataRecord(event.metadata?.result_summary) ?? current?.resultSummary,
    errorCode: metadataString(event.metadata?.error_code) ?? current?.errorCode,
    errorMessage: metadataString(event.metadata?.error_message) ?? current?.errorMessage,
  }
  if (!toolCalls?.length) return [next]
  if (index >= 0) return toolCalls.map((call, itemIndex) => itemIndex === index ? next : call)
  return [next, ...toolCalls]
}

export function applyRealRunEvent(run: Run, event: RuntimeEvent): Run {
  if (isRuntimeTerminal(run.status)) return run
  if (run.events.some((existing) => existing.id === event.id)) return run

  const events = [...run.events, event].sort((a, b) => (a.sequence ?? 0) - (b.sequence ?? 0))
  const completedAt = event.status === 'completed' || event.status === 'failed' || event.status === 'stopped' ? event.time : run.completedAt
  const toolMapping = toolEventMapping(event.type)
  if (toolMapping) {
    const assistantDraft = event.type === 'tool.call.succeeded' && run.assistantDraft?.status === 'streaming'
      ? { ...run.assistantDraft, status: 'paused_for_tool' as const, lastEventId: event.id }
      : event.type === 'tool.call.denied'
        ? { ...run.assistantDraft, content: run.assistantDraft?.content ?? '', status: 'stopped' as const, lastEventId: event.id }
        : event.type === 'tool.call.failed'
          ? { ...run.assistantDraft, content: run.assistantDraft?.content ?? '', status: 'failed' as const, lastEventId: event.id }
          : run.assistantDraft
    return { ...run, status: event.status, events, completedAt, assistantDraft, toolCalls: applyToolEvent(run.toolCalls, event, toolMapping) }
  }
  if (event.type === 'model.delta' || event.type === 'message.model_output_delta') {
    const delta = event.assistantDelta ?? event.content ?? ''
    const isContinuation = event.metadata?.model_phase === 'continuation'
    const baseContent = isContinuation && run.assistantDraft?.status === 'paused_for_tool' ? '' : run.assistantDraft?.content ?? ''
    return {
      ...run,
      status: event.status,
      events,
      completedAt,
      assistantDraft: {
        ...run.assistantDraft,
        content: `${baseContent}${delta}`,
        status: 'streaming',
        lastEventId: event.id,
      },
    }
  }
  if (event.type === 'assistant.message.completed' || event.type === 'message.model_output_completed' || event.type === 'model.final') {
    return {
      ...run,
      status: event.status,
      events,
      completedAt,
      assistantDraft: {
        ...run.assistantDraft,
        content: event.content ?? run.assistantDraft?.content ?? '',
        status: 'completed',
        lastEventId: event.id,
      },
    }
  }
  if (event.status === 'completed') {
    return { ...run, status: 'completed', events, completedAt }
  }
  if (event.status === 'failed' || event.status === 'stopped') {
    return {
      ...run,
      status: event.status,
      events,
      completedAt,
      assistantDraft: {
        ...run.assistantDraft,
        content: run.assistantDraft?.content ?? event.content ?? '',
        status: event.status,
        lastEventId: event.id,
      },
    }
  }
  if (event.status === 'recovering' || event.status === 'queued' || event.status === 'stopping') {
    return {
      ...run,
      status: event.status,
      events,
      assistantDraft: {
        ...run.assistantDraft,
        content: run.assistantDraft?.content ?? event.content ?? '',
        status: event.status,
        lastEventId: event.id,
      },
    }
  }
  return { ...run, status: event.status, events }
}

export function mapRealRuntimeCapabilitySignal(error: unknown): RealRuntimeCapabilitySignal {
  const code = typeof error === 'object' && error !== null && 'code' in error ? String(error.code) : ''
  const message = error instanceof Error ? error.message.toLowerCase() : ''
  if (code === 'stream_disconnected') return { streamDisconnected: true }
  if (code === 'provider_unavailable' || message.includes('provider')) return { providerUnavailable: true }
  if (code === 'model_setup_missing' || message.includes('model setup')) return { modelSetupMissing: true }
  return { backendUnavailable: true }
}

export const realExecutionAdapter: ExecutionAdapter = {
  runtimeCapability: 'available',
  async sendMessage() {
    delegated('sendMessage')
  },
  async createRun() {
    delegated('startRun')
  },
  async subscribeRunEvents() {
    return () => {}
  },
  async appendAssistantDelta() {
    delegated('subscribeRunEvents')
  },
  async completeRun() {
    delegated('subscribeRunEvents')
  },
  async failRun() {
    delegated('subscribeRunEvents')
  },
  async stopRun() {
    delegated('stopRun')
  },
}
