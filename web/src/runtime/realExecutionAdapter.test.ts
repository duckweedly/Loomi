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

  test('uses continuation deltas as the final assistant draft after tool success', () => {
    const run = { id: 'run-a', threadId: 'thread-a', status: 'running', model: 'Model gateway', context: 'model_gateway', source: 'model_gateway', events: [] } as const
    const initialDelta: RuntimeEvent = { id: 'evt-1', runId: 'run-a', threadId: 'thread-a', sequence: 1, type: 'message.model_output_delta', label: 'message', detail: 'Model output delta', content: 'I will check.', assistantDelta: 'I will check.', time: 'Now', status: 'running', metadata: { model_phase: 'initial' } }
    const toolSucceeded: RuntimeEvent = { id: 'evt-2', runId: 'run-a', threadId: 'thread-a', sequence: 2, type: 'tool.call.succeeded', label: 'tool', detail: 'Tool call succeeded', time: 'Now', status: 'running', group: 'tool-call', metadata: { tool_call_id: 'tc_1', tool_name: 'runtime.get_current_time', result_summary: { iso_time: '2026-05-25T10:00:00Z' } } }
    const continuationDelta: RuntimeEvent = { id: 'evt-3', runId: 'run-a', threadId: 'thread-a', sequence: 3, type: 'message.model_output_delta', label: 'message', detail: 'Model output delta', content: 'It is 2026-05-25T10:00:00Z.', assistantDelta: 'It is 2026-05-25T10:00:00Z.', time: 'Now', status: 'running', metadata: { model_phase: 'continuation' } }
    const completed: RuntimeEvent = { id: 'evt-4', runId: 'run-a', threadId: 'thread-a', sequence: 4, type: 'message.model_output_completed', label: 'message', detail: 'Model output completed', content: 'It is 2026-05-25T10:00:00Z.', time: 'Now', status: 'running', metadata: { model_phase: 'continuation' } }

    const afterInitial = applyRealRunEvent(run, initialDelta)
    const afterTool = applyRealRunEvent(afterInitial, toolSucceeded)
    const afterContinuation = applyRealRunEvent(afterTool, continuationDelta)
    const final = applyRealRunEvent(afterContinuation, completed)

    expect(afterInitial.assistantDraft?.content).toBe('I will check.')
    expect(afterTool.assistantDraft).toMatchObject({ content: 'I will check.', status: 'paused_for_tool' })
    expect(afterContinuation.assistantDraft?.content).toBe('It is 2026-05-25T10:00:00Z.')
    expect(final.assistantDraft).toMatchObject({ content: 'It is 2026-05-25T10:00:00Z.', status: 'completed' })
    expect(final.toolCalls?.[0].resultSummary).toEqual({ iso_time: '2026-05-25T10:00:00Z' })
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

  test('projects safe thinking summary metadata without raw hidden reasoning', () => {
    const run = { id: 'run-a', threadId: 'thread-a', status: 'running', model: 'Model gateway', context: 'model_gateway', source: 'model_gateway', events: [] } as const
    const event: RuntimeEvent = {
      id: 'evt-thinking-summary',
      runId: 'run-a',
      threadId: 'thread-a',
      type: 'run.completed',
      label: 'Run',
      detail: 'Run completed',
      time: 'Now',
      status: 'completed',
      metadata: {
        thinking_summary: '检查输入并收束答案',
        thinking_duration_seconds: 12,
        raw_thinking: 'do not expose this hidden chain',
      },
    }

    const next = applyRealRunEvent(run, event)

    expect(next.thinkingSummary).toBe('检查输入并收束答案')
    expect(next.thinkingDurationSeconds).toBe(12)
    expect(JSON.stringify(next)).not.toContain('hidden chain')
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

  test('maps M7 terminal tool events into result and error states', () => {
    const run = { id: 'run-a', threadId: 'thread-a', status: 'running', model: 'Model gateway', context: 'model_gateway', source: 'model_gateway', events: [], toolCalls: [] } as const
    const succeeded: RuntimeEvent = { id: 'evt-tool-success', runId: 'run-a', threadId: 'thread-a', sequence: 1, type: 'tool.call.succeeded', label: 'tool', detail: 'Tool call succeeded', time: 'Now', status: 'running', group: 'tool-call', metadata: { tool_call_id: 'tc_1', tool_name: 'runtime.get_current_time', approval_status: 'approved', execution_status: 'succeeded', result_summary: { timezone: 'UTC' } } }
    const failed: RuntimeEvent = { id: 'evt-tool-failed', runId: 'run-a', threadId: 'thread-a', sequence: 2, type: 'tool.call.failed', label: 'tool', detail: 'Tool call failed', time: 'Later', status: 'failed', group: 'tool-call', metadata: { tool_call_id: 'tc_2', tool_name: 'runtime.get_current_time', approval_status: 'approved', execution_status: 'failed', error_code: 'tool_execution_failed', error_message: 'Tool execution failed.' } }
    const denied: RuntimeEvent = { id: 'evt-tool-denied', runId: 'run-a', threadId: 'thread-a', sequence: 3, type: 'tool.call.denied', label: 'tool', detail: 'Tool call denied', time: 'Later', status: 'stopped', group: 'tool-call', metadata: { tool_call_id: 'tc_3', tool_name: 'runtime.get_current_time', approval_status: 'denied', execution_status: 'cancelled' } }

    const successRun = applyRealRunEvent(run, succeeded)
    const failedRun = applyRealRunEvent(run, failed)
    const deniedRun = applyRealRunEvent(run, denied)

    expect(successRun.toolCalls?.[0]).toMatchObject({ status: 'succeeded', executionStatus: 'succeeded', resultSummary: { timezone: 'UTC' } })
    expect(failedRun.toolCalls?.[0]).toMatchObject({ status: 'failed', executionStatus: 'failed', errorCode: 'tool_execution_failed', errorMessage: 'Tool execution failed.' })
    expect(deniedRun.toolCalls?.[0]).toMatchObject({ status: 'denied', approvalStatus: 'denied', executionStatus: 'cancelled' })
  })

  test('preserves MCP tool metadata through replayed tool events', () => {
    const run = { id: 'run-a', threadId: 'thread-a', status: 'running', model: 'Model gateway', context: 'model_gateway', source: 'model_gateway', events: [], toolCalls: [] } as const
    const required: RuntimeEvent = { id: 'evt-mcp-required', runId: 'run-a', threadId: 'thread-a', sequence: 1, type: 'tool.call.approval_required', label: 'tool', detail: 'Tool approval required', time: 'Now', status: 'blocked_on_tool_approval', group: 'tool-call', metadata: { tool_call_id: 'tc_mcp_1', tool_name: 'mcp.local-search.search', tool_source: 'mcp', server_slug: 'local-search', arguments_summary: { query: 'status' }, approval_status: 'required', execution_status: 'blocked' } }
    const succeeded: RuntimeEvent = { id: 'evt-mcp-succeeded', runId: 'run-a', threadId: 'thread-a', sequence: 2, type: 'tool.call.succeeded', label: 'tool', detail: 'Tool call succeeded', time: 'Later', status: 'running', group: 'tool-call', metadata: { tool_call_id: 'tc_mcp_1', tool_name: 'mcp.local-search.search', tool_source: 'mcp', server_slug: 'local-search', approval_status: 'approved', execution_status: 'succeeded', result_summary: { summary: 'safe' } } }

    const pending = applyRealRunEvent(run, required)
    const final = applyRealRunEvent(pending, succeeded)

    expect(pending.toolCalls?.[0]).toMatchObject({ toolCallId: 'tc_mcp_1', name: 'mcp.local-search.search', approvalStatus: 'required', executionStatus: 'blocked', argumentsSummary: { query: 'status' } })
    expect(final.toolCalls?.[0]).toMatchObject({ status: 'succeeded', resultSummary: { summary: 'safe' } })
  })

  test('does not build continuation draft after denied or failed tool terminals', () => {
    const run = { id: 'run-a', threadId: 'thread-a', status: 'running', model: 'Model gateway', context: 'model_gateway', source: 'model_gateway', events: [], assistantDraft: { content: 'I will check.', status: 'streaming' } } as const
    const denied: RuntimeEvent = { id: 'evt-denied', runId: 'run-a', threadId: 'thread-a', sequence: 1, type: 'tool.call.denied', label: 'tool', detail: 'Tool call denied', time: 'Now', status: 'stopped', group: 'tool-call', metadata: { tool_call_id: 'tc_1', tool_name: 'runtime.get_current_time', approval_status: 'denied', execution_status: 'cancelled' } }
    const failed: RuntimeEvent = { id: 'evt-failed', runId: 'run-a', threadId: 'thread-a', sequence: 1, type: 'tool.call.failed', label: 'tool', detail: 'Tool call failed', time: 'Now', status: 'failed', group: 'tool-call', metadata: { tool_call_id: 'tc_1', tool_name: 'runtime.get_current_time', approval_status: 'approved', execution_status: 'failed', error_code: 'tool_execution_failed' } }
    const continuation: RuntimeEvent = { id: 'evt-continuation', runId: 'run-a', threadId: 'thread-a', sequence: 2, type: 'message.model_output_delta', label: 'message', detail: 'Model output delta', content: 'It is now.', assistantDelta: 'It is now.', time: 'Later', status: 'running', metadata: { model_phase: 'continuation' } }

    const afterDenied = applyRealRunEvent(applyRealRunEvent(run, denied), continuation)
    const afterFailed = applyRealRunEvent(applyRealRunEvent(run, failed), continuation)

    expect(afterDenied.assistantDraft).toMatchObject({ content: 'I will check.', status: 'stopped' })
    expect(afterFailed.assistantDraft).toMatchObject({ content: 'I will check.', status: 'failed' })
    expect(afterDenied.events.map((event) => event.id)).not.toContain('evt-continuation')
    expect(afterFailed.events.map((event) => event.id)).not.toContain('evt-continuation')
  })

  test('maps backend setup provider and stream failures to capability signals', () => {
    expect(mapRealRuntimeCapabilitySignal(new Error('Failed to fetch'))).toEqual({ backendUnavailable: true })
    expect(mapRealRuntimeCapabilitySignal(Object.assign(new Error('model setup missing'), { code: 'model_setup_missing' }))).toEqual({ modelSetupMissing: true })
    expect(mapRealRuntimeCapabilitySignal(Object.assign(new Error('provider unavailable'), { code: 'provider_unavailable' }))).toEqual({ providerUnavailable: true })
    expect(mapRealRuntimeCapabilitySignal(Object.assign(new Error('stream disconnected'), { code: 'stream_disconnected' }))).toEqual({ streamDisconnected: true })
  })
})
