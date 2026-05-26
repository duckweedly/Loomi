import { describe, expect, test } from 'bun:test'
import type { RuntimeEvent } from '../domain'
import { applyRealRunEvent } from './realExecutionAdapter'

describe('M6 execution adapter worker events', () => {
  const baseRun = { id: 'run-a', threadId: 'thread-a', status: 'queued', model: 'Model gateway', context: 'model_gateway', source: 'model_gateway', events: [] } as const

  test('maps queued claimed and completed events into run state', () => {
    const queued: RuntimeEvent = { id: 'evt-queued', runId: 'run-a', threadId: 'thread-a', sequence: 1, type: 'run.queued', label: 'lifecycle', detail: 'Run queued', time: 'Now', status: 'queued' }
    const claimed: RuntimeEvent = { id: 'evt-claimed', runId: 'run-a', threadId: 'thread-a', sequence: 2, type: 'job.claimed', label: 'progress', detail: 'Job claimed', time: 'Now', status: 'running' }
    const completed: RuntimeEvent = { id: 'evt-completed', runId: 'run-a', threadId: 'thread-a', sequence: 3, type: 'run.completed', label: 'final', detail: 'Run completed', time: 'Now', status: 'completed' }

    const queuedRun = applyRealRunEvent(baseRun, queued)
    const runningRun = applyRealRunEvent(queuedRun, claimed)
    const completedRun = applyRealRunEvent(runningRun, completed)

    expect(queuedRun.status).toBe('queued')
    expect(queuedRun.assistantDraft).toMatchObject({ status: 'queued' })
    expect(runningRun.status).toBe('running')
    expect(completedRun.status).toBe('completed')
    expect(completedRun.completedAt).toBe('Now')
    expect(completedRun.events.map((event) => event.type)).toEqual(['run.queued', 'job.claimed', 'run.completed'])
  })

  test('maps recovery and retry exhaustion into visible run states', () => {
    const recovering: RuntimeEvent = { id: 'evt-recovering', runId: 'run-a', threadId: 'thread-a', sequence: 1, type: 'job.recovering', label: 'progress', detail: 'Job recovering', time: 'Now', status: 'recovering' }
    const retry: RuntimeEvent = { id: 'evt-retry', runId: 'run-a', threadId: 'thread-a', sequence: 2, type: 'job.retry_scheduled', label: 'progress', detail: 'Job retry scheduled', time: 'Now', status: 'recovering' }
    const exhausted: RuntimeEvent = { id: 'evt-exhausted', runId: 'run-a', threadId: 'thread-a', sequence: 3, type: 'job.retry_exhausted', label: 'error', detail: 'Retries exhausted', time: 'Later', status: 'failed' }

    const recoveringRun = applyRealRunEvent(baseRun, recovering)
    const retryRun = applyRealRunEvent(recoveringRun, retry)
    const failedRun = applyRealRunEvent(retryRun, exhausted)

    expect(recoveringRun.status).toBe('recovering')
    expect(recoveringRun.assistantDraft).toMatchObject({ status: 'recovering' })
    expect(retryRun.status).toBe('recovering')
    expect(failedRun.status).toBe('failed')
    expect(failedRun.completedAt).toBe('Later')
    expect(failedRun.assistantDraft).toMatchObject({ status: 'failed' })
    expect(failedRun.events.map((event) => event.type)).toEqual(['job.recovering', 'job.retry_scheduled', 'job.retry_exhausted'])
  })

  test('keeps pipeline events observable while the worker run is active', () => {
    const started: RuntimeEvent = { id: 'evt-step-started', runId: 'run-a', threadId: 'thread-a', sequence: 1, type: 'pipeline.step.started', label: 'progress', detail: 'Invoke runtime started', time: 'Now', status: 'running', group: 'worker-job' }
    const completed: RuntimeEvent = { id: 'evt-step-completed', runId: 'run-a', threadId: 'thread-a', sequence: 2, type: 'pipeline.step.completed', label: 'progress', detail: 'Invoke runtime completed', time: 'Now', status: 'running', group: 'worker-job' }

    const next = applyRealRunEvent(applyRealRunEvent(baseRun, started), completed)

    expect(next.status).toBe('running')
    expect(next.events.map((event) => event.type)).toEqual(['pipeline.step.started', 'pipeline.step.completed'])
  })

  test('replays M7 tool events into a stable tool-call view model', () => {
    const requested: RuntimeEvent = { id: 'evt-tool-1', runId: 'run-a', threadId: 'thread-a', sequence: 1, type: 'tool.call.requested', label: 'tool', detail: 'Tool call requested', time: 'Now', status: 'blocked_on_tool_approval', group: 'tool-call', metadata: { tool_call_id: 'tc_1', tool_name: 'runtime.get_current_time', arguments_summary: { timezone: 'UTC' }, approval_status: 'required', execution_status: 'blocked' } }
    const required: RuntimeEvent = { id: 'evt-tool-2', runId: 'run-a', threadId: 'thread-a', sequence: 2, type: 'tool.call.approval_required', label: 'tool', detail: 'Tool approval required', time: 'Now', status: 'blocked_on_tool_approval', group: 'tool-call', metadata: { tool_call_id: 'tc_1', tool_name: 'runtime.get_current_time', arguments_summary: { timezone: 'UTC' }, approval_status: 'required', execution_status: 'blocked' } }
    const approved: RuntimeEvent = { id: 'evt-tool-3', runId: 'run-a', threadId: 'thread-a', sequence: 3, type: 'tool.call.approved', label: 'tool', detail: 'Tool call approved', time: 'Now', status: 'running', group: 'tool-call', metadata: { tool_call_id: 'tc_1', tool_name: 'runtime.get_current_time', arguments_summary: { timezone: 'UTC' }, approval_status: 'approved', execution_status: 'blocked' } }

    const next = applyRealRunEvent(applyRealRunEvent(applyRealRunEvent(baseRun, requested), required), approved)

    expect(next.status).toBe('running')
    expect(next.toolCalls).toHaveLength(1)
    expect(next.toolCalls?.[0]).toMatchObject({ toolCallId: 'tc_1', name: 'runtime.get_current_time', status: 'approved', approvalStatus: 'approved', executionStatus: 'blocked', summary: 'Tool call approved', argumentsSummary: { timezone: 'UTC' } })
    expect(next.events.map((event) => event.type)).toEqual(['tool.call.requested', 'tool.call.approval_required', 'tool.call.approved'])
  })

  test('replays M7 execution terminal events into one stable tool-call view model', () => {
    const required: RuntimeEvent = { id: 'evt-tool-1', runId: 'run-a', threadId: 'thread-a', sequence: 1, type: 'tool.call.approval_required', label: 'tool', detail: 'Tool approval required', time: 'Now', status: 'blocked_on_tool_approval', group: 'tool-call', metadata: { tool_call_id: 'tc_1', tool_name: 'runtime.get_current_time', arguments_summary: { timezone: 'UTC' }, approval_status: 'required', execution_status: 'blocked' } }
    const approved: RuntimeEvent = { id: 'evt-tool-2', runId: 'run-a', threadId: 'thread-a', sequence: 2, type: 'tool.call.approved', label: 'tool', detail: 'Tool call approved', time: 'Now', status: 'running', group: 'tool-call', metadata: { tool_call_id: 'tc_1', tool_name: 'runtime.get_current_time', approval_status: 'approved', execution_status: 'not_started' } }
    const executing: RuntimeEvent = { id: 'evt-tool-3', runId: 'run-a', threadId: 'thread-a', sequence: 3, type: 'tool.call.executing', label: 'tool', detail: 'Tool call executing', time: 'Now', status: 'running', group: 'tool-call', metadata: { tool_call_id: 'tc_1', tool_name: 'runtime.get_current_time', approval_status: 'approved', execution_status: 'executing' } }
    const succeeded: RuntimeEvent = { id: 'evt-tool-4', runId: 'run-a', threadId: 'thread-a', sequence: 4, type: 'tool.call.succeeded', label: 'tool', detail: 'Tool call succeeded', time: 'Now', status: 'running', group: 'tool-call', metadata: { tool_call_id: 'tc_1', tool_name: 'runtime.get_current_time', approval_status: 'approved', execution_status: 'succeeded', result_summary: { timezone: 'UTC', local_time: '2026-05-26T10:00:00Z' } } }
    const failed: RuntimeEvent = { id: 'evt-tool-5', runId: 'run-a', threadId: 'thread-a', sequence: 5, type: 'tool.call.failed', label: 'tool', detail: 'Tool call failed', time: 'Now', status: 'running', group: 'tool-call', metadata: { tool_call_id: 'tc_2', tool_name: 'runtime.get_current_time', arguments_summary: { timezone: 'Mars/Olympus' }, approval_status: 'approved', execution_status: 'failed', error_code: 'validation_failed', error_message: 'Invalid timezone' } }
    const cancelled: RuntimeEvent = { id: 'evt-tool-6', runId: 'run-a', threadId: 'thread-a', sequence: 6, type: 'tool.call.cancelled', label: 'tool', detail: 'Tool call cancelled', time: 'Now', status: 'cancelled', group: 'tool-call', metadata: { tool_call_id: 'tc_3', tool_name: 'runtime.get_current_time', arguments_summary: { timezone: 'UTC' }, approval_status: 'denied', execution_status: 'cancelled' } }

    const next = [required, approved, executing, succeeded, failed, cancelled].reduce(applyRealRunEvent, baseRun)

    expect(next.toolCalls?.find((call) => call.toolCallId === 'tc_1')).toMatchObject({ status: 'succeeded', approvalStatus: 'approved', executionStatus: 'succeeded', resultSummary: { timezone: 'UTC', local_time: '2026-05-26T10:00:00Z' } })
    expect(next.toolCalls?.find((call) => call.toolCallId === 'tc_2')).toMatchObject({ status: 'failed', approvalStatus: 'approved', executionStatus: 'failed', errorCode: 'validation_failed', errorMessage: 'Invalid timezone' })
    expect(next.toolCalls?.find((call) => call.toolCallId === 'tc_3')).toMatchObject({ status: 'cancelled', approvalStatus: 'denied', executionStatus: 'cancelled' })
  })

  test('replays M8 workspace read tool summaries without changing the tool lifecycle model', () => {
    const required: RuntimeEvent = { id: 'evt-tool-1', runId: 'run-a', threadId: 'thread-a', sequence: 1, type: 'tool.call.approval_required', label: 'tool', detail: 'Tool approval required', time: 'Now', status: 'blocked_on_tool_approval', group: 'tool-call', metadata: { tool_call_id: 'tc_glob', tool_name: 'workspace.glob', arguments_summary: { pattern: '**/*.go', limit: 10 }, approval_status: 'required', execution_status: 'blocked' } }
    const succeeded: RuntimeEvent = { id: 'evt-tool-2', runId: 'run-a', threadId: 'thread-a', sequence: 2, type: 'tool.call.succeeded', label: 'tool', detail: 'Tool call succeeded', time: 'Now', status: 'running', group: 'tool-call', metadata: { tool_call_id: 'tc_glob', tool_name: 'workspace.glob', approval_status: 'approved', execution_status: 'succeeded', result_summary: { matches: ['internal/runtime/tools.go'], match_count: 1, truncated: false } } }

    const next = applyRealRunEvent(applyRealRunEvent(baseRun, required), succeeded)

    expect(next.toolCalls?.[0]).toMatchObject({
      toolCallId: 'tc_glob',
      name: 'workspace.glob',
      status: 'succeeded',
      argumentsSummary: { pattern: '**/*.go', limit: 10 },
      resultSummary: { matches: ['internal/runtime/tools.go'], match_count: 1, truncated: false },
    })
  })

  test('replays M9 workspace write and edit summaries through the same tool model', () => {
    const writeSucceeded: RuntimeEvent = { id: 'evt-write', runId: 'run-a', threadId: 'thread-a', sequence: 1, type: 'tool.call.succeeded', label: 'tool', detail: 'Tool call succeeded', time: 'Now', status: 'running', group: 'tool-call', metadata: { tool_call_id: 'tc_write', tool_name: 'workspace.write_file', arguments_summary: { path: 'internal/generated.txt' }, approval_status: 'approved', execution_status: 'succeeded', result_summary: { path: 'internal/generated.txt', bytes_written: 12, created: true, truncated: false } } }
    const editSucceeded: RuntimeEvent = { id: 'evt-edit', runId: 'run-a', threadId: 'thread-a', sequence: 2, type: 'tool.call.succeeded', label: 'tool', detail: 'Tool call succeeded', time: 'Now', status: 'running', group: 'tool-call', metadata: { tool_call_id: 'tc_edit', tool_name: 'workspace.edit', arguments_summary: { path: 'internal/generated.txt' }, approval_status: 'approved', execution_status: 'succeeded', result_summary: { path: 'internal/generated.txt', replacements: 1, bytes_before: 12, bytes_after: 13 } } }

    const next = applyRealRunEvent(applyRealRunEvent(baseRun, writeSucceeded), editSucceeded)

    expect(next.toolCalls?.find((call) => call.toolCallId === 'tc_write')).toMatchObject({ name: 'workspace.write_file', status: 'succeeded', resultSummary: { path: 'internal/generated.txt', bytes_written: 12, created: true } })
    expect(next.toolCalls?.find((call) => call.toolCallId === 'tc_edit')).toMatchObject({ name: 'workspace.edit', status: 'succeeded', resultSummary: { path: 'internal/generated.txt', replacements: 1, bytes_after: 13 } })
  })

  test('replays M10 workspace exec command summaries through the same tool model', () => {
    const succeeded: RuntimeEvent = { id: 'evt-exec', runId: 'run-a', threadId: 'thread-a', sequence: 1, type: 'tool.call.succeeded', label: 'tool', detail: 'Tool call succeeded', time: 'Now', status: 'running', group: 'tool-call', metadata: { tool_call_id: 'tc_exec', tool_name: 'workspace.exec_command', arguments_summary: { command: ['printf', 'hello'], cwd: '.' }, approval_status: 'approved', execution_status: 'succeeded', result_summary: { cwd: '.', exit_code: 0, stdout: 'hello', stderr: '', timed_out: false, stdout_truncated: false, stderr_truncated: false } } }

    const next = applyRealRunEvent(baseRun, succeeded)

    expect(next.toolCalls?.[0]).toMatchObject({ name: 'workspace.exec_command', status: 'succeeded', resultSummary: { cwd: '.', exit_code: 0, stdout: 'hello', timed_out: false } })
  })
})
