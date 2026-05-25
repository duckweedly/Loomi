import { describe, expect, test } from 'bun:test'
import type { RuntimeEvent } from '../domain'
import { applyRealRunEvent, mapRealRuntimeCapabilitySignal } from './realExecutionAdapter'

describe('applyRealRunEvent', () => {
  test('applies model gateway delta and completion events', () => {
    const run = { id: 'run-a', threadId: 'thread-a', status: 'running', model: 'Model gateway', context: 'model_gateway', source: 'model_gateway', events: [] } as const
    const delta: RuntimeEvent = { id: 'evt-1', runId: 'run-a', threadId: 'thread-a', type: 'message.model_output_delta', label: 'message', detail: 'Model output delta', content: 'hel', assistantDelta: 'hel', time: 'Now', status: 'running' }
    const completed: RuntimeEvent = { id: 'evt-2', runId: 'run-a', threadId: 'thread-a', type: 'message.model_output_completed', label: 'message', detail: 'Model output completed', content: 'hello', time: 'Now', status: 'running' }

    const drafting = applyRealRunEvent(run, delta)
    const final = applyRealRunEvent(drafting, completed)

    expect(drafting.assistantDraft?.content).toBe('hel')
    expect(final.assistantDraft).toMatchObject({ content: 'hello', status: 'completed', lastEventId: 'evt-2' })
    expect(final.events.map((event) => event.id)).toEqual(['evt-1', 'evt-2'])
  })

  test('ignores late provider events after a terminal run', () => {
    const run = { id: 'run-a', threadId: 'thread-a', status: 'stopped', model: 'Model gateway', context: 'model_gateway', source: 'model_gateway', events: [], assistantDraft: { content: '', status: 'stopped' } } as const
    const event: RuntimeEvent = { id: 'evt-late', runId: 'run-a', threadId: 'thread-a', type: 'message.model_output_delta', label: 'message', detail: 'Late delta', content: 'late', assistantDelta: 'late', time: 'Now', status: 'running' }

    const next = applyRealRunEvent(run, event)

    expect(next).toBe(run)
  })

  test('surfaces provider failure states without mock fallback output', () => {
    const run = { id: 'run-a', threadId: 'thread-a', status: 'running', model: 'Model gateway', context: 'model_gateway', source: 'model_gateway', events: [], assistantDraft: { content: '', status: 'empty' } } as const
    const event: RuntimeEvent = { id: 'evt-failed', runId: 'run-a', threadId: 'thread-a', type: 'error.provider_rate_limited', label: 'error', detail: 'Provider rate limit reached.', time: 'Now', status: 'failed' }

    const next = applyRealRunEvent(run, event)

    expect(next.status).toBe('failed')
    expect(next.assistantDraft?.content).toBe('')
    expect(next.events[0].type).toBe('error.provider_rate_limited')
  })

  test('keeps tool boundary events observable without executing them', () => {
    const run = { id: 'run-a', threadId: 'thread-a', status: 'running', model: 'Model gateway', context: 'model_gateway', source: 'model_gateway', events: [] } as const
    const event: RuntimeEvent = { id: 'evt-tool', runId: 'run-a', threadId: 'thread-a', type: 'progress.tool_call_blocked', label: 'progress', detail: 'Tool execution is outside this milestone.', time: 'Now', status: 'running' }

    const next = applyRealRunEvent(run, event)

    expect(next.events).toHaveLength(1)
    expect(next.events[0].type).toBe('progress.tool_call_blocked')
  })

  test('maps M7 tool lifecycle events into stable tool call state', () => {
    const run = { id: 'run-a', threadId: 'thread-a', status: 'running', model: 'Model gateway', context: 'model_gateway', source: 'model_gateway', events: [], toolCalls: [] } as const
    const requested: RuntimeEvent = { id: 'evt-tool-1', runId: 'run-a', threadId: 'thread-a', sequence: 1, type: 'tool.call.requested', label: 'tool', detail: 'Tool call requested', time: 'Now', status: 'blocked_on_tool_approval', group: 'tool-call' }
    const approved: RuntimeEvent = { id: 'evt-tool-2', runId: 'run-a', threadId: 'thread-a', sequence: 2, type: 'tool.call.approved', label: 'tool', detail: 'Tool call approved', time: 'Now', status: 'running', group: 'tool-call' }

    const pending = applyRealRunEvent(run, requested)
    const next = applyRealRunEvent(pending, approved)

    expect(pending.status).toBe('blocked_on_tool_approval')
    expect(pending.toolCalls?.[0]).toMatchObject({ status: 'requested', summary: 'Tool call requested' })
    expect(next.toolCalls?.[0]).toMatchObject({ status: 'approved', approvalStatus: 'approved' })
  })

  test('maps backend setup provider and stream failures to capability signals', () => {
    expect(mapRealRuntimeCapabilitySignal(new Error('Failed to fetch'))).toEqual({ backendUnavailable: true })
    expect(mapRealRuntimeCapabilitySignal(Object.assign(new Error('model setup missing'), { code: 'model_setup_missing' }))).toEqual({ modelSetupMissing: true })
    expect(mapRealRuntimeCapabilitySignal(Object.assign(new Error('provider unavailable'), { code: 'provider_unavailable' }))).toEqual({ providerUnavailable: true })
    expect(mapRealRuntimeCapabilitySignal(Object.assign(new Error('stream disconnected'), { code: 'stream_disconnected' }))).toEqual({ streamDisconnected: true })
  })
})
