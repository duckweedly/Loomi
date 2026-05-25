import { describe, expect, test } from 'bun:test'
import type { RunEvent } from '../domain'
import { groupRuntimeEvents, mapRuntimeEventGroup } from './runtimeEventGroups'

function event(overrides: Partial<RunEvent>): RunEvent {
  return {
    id: overrides.id ?? `evt-${overrides.type ?? 'run.created'}`,
    runId: 'run-a',
    threadId: 'thread-a',
    type: overrides.type ?? 'run.created',
    label: overrides.label ?? 'Run',
    detail: overrides.detail ?? 'detail',
    time: overrides.time ?? 'Now',
    status: overrides.status ?? 'running',
    severity: overrides.severity,
    group: overrides.group,
    usage: overrides.usage,
    sequence: overrides.sequence,
  }
}

describe('runtime event groups', () => {
  test('maps lifecycle, model stream, worker job, error, and unknown events', () => {
    expect(mapRuntimeEventGroup(event({ type: 'run.created' }))).toBe('run-lifecycle')
    expect(mapRuntimeEventGroup(event({ type: 'model.delta' }))).toBe('model-stream')
    expect(mapRuntimeEventGroup(event({ type: 'worker.claimed' }))).toBe('worker-job')
    expect(mapRuntimeEventGroup(event({ type: 'pipeline.step.started' }))).toBe('worker-job')
    expect(mapRuntimeEventGroup(event({ type: 'pipeline.step.completed' }))).toBe('worker-job')
    expect(mapRuntimeEventGroup(event({ type: 'pipeline.step.failed' }))).toBe('error')
    expect(mapRuntimeEventGroup(event({ type: 'mcp.discovery.succeeded' }))).toBe('worker-job')
    expect(mapRuntimeEventGroup(event({ type: 'mcp.tools.available' }))).toBe('worker-job')
    expect(mapRuntimeEventGroup(event({ type: 'mcp.discovery.failed' }))).toBe('error')
    expect(mapRuntimeEventGroup(event({ type: 'tool.call.executing', metadata: { tool_source: 'mcp' } }))).toBe('tool-call')
    expect(mapRuntimeEventGroup(event({ type: 'provider.error', status: 'failed' }))).toBe('error')
    expect(mapRuntimeEventGroup(event({ type: 'backend.unavailable' }))).toBe('error')
    expect(mapRuntimeEventGroup(event({ type: 'provider.timeout' }))).toBe('error')
    expect(mapRuntimeEventGroup(event({ type: 'custom.telemetry' }))).toBe('run-lifecycle')
  })

  test('gives error semantics precedence over explicit groups', () => {
    expect(mapRuntimeEventGroup(event({ type: 'run.failed', status: 'failed', group: 'run-lifecycle' }))).toBe('error')
    expect(mapRuntimeEventGroup(event({ type: 'provider.timeout', severity: 'error', group: 'model-stream' }))).toBe('error')
  })

  test('returns stable groups preserving incoming stream order and usage detail', () => {
    const grouped = groupRuntimeEvents([
      event({ id: 'evt-model', sequence: 2, type: 'model.usage', usage: { inputTokens: 7, outputTokens: 11 } }),
      event({ id: 'evt-run', sequence: 1, type: 'run.created' }),
      event({ id: 'evt-worker', sequence: 3, type: 'job.queued' }),
      event({ id: 'evt-tool', sequence: 4, type: 'tool.call.requested', group: 'tool-call' }),
      event({ id: 'evt-error', sequence: 5, type: 'stream.error', status: 'failed' }),
      event({ id: 'evt-worker-recovering', sequence: 0, type: 'job.recovering', detail: 'recovering' }),
    ])

    expect(grouped.map((group) => group.id)).toEqual(['run-lifecycle', 'model-stream', 'worker-job', 'tool-call', 'error'])
    expect(grouped[0].events.map((item) => item.id)).toEqual(['evt-run'])
    expect(grouped[1].events[0].usage).toEqual({ inputTokens: 7, outputTokens: 11 })
    expect(grouped[2].events.map((item) => item.id)).toEqual(['evt-worker', 'evt-worker-recovering'])
    expect(grouped[3].events.map((item) => item.id)).toEqual(['evt-tool'])
    expect(grouped[4].events.map((item) => item.id)).toEqual(['evt-error'])
  })

  test('maps productized M6 worker job event names and unknown worker events', () => {
    expect(mapRuntimeEventGroup(event({ type: 'job_claimed' }))).toBe('worker-job')
    expect(mapRuntimeEventGroup(event({ type: 'lease_renewed' }))).toBe('worker-job')
    expect(mapRuntimeEventGroup(event({ type: 'job_recovering' }))).toBe('worker-job')
    expect(mapRuntimeEventGroup(event({ type: 'job_retry_scheduled' }))).toBe('worker-job')
    expect(mapRuntimeEventGroup(event({ type: 'job_attempt_failed' }))).toBe('error')
    expect(mapRuntimeEventGroup(event({ type: 'job_retry_exhausted' }))).toBe('error')
    expect(mapRuntimeEventGroup(event({ type: 'future_worker_event' }))).toBe('worker-job')
  })
})

describe('localized runtime event groups', () => {
  test('returns Chinese group titles when requested', () => {
    const grouped = groupRuntimeEvents([
      event({ id: 'evt-run', sequence: 1, type: 'run.created' }),
      event({ id: 'evt-model', sequence: 2, type: 'model.delta' }),
      event({ id: 'evt-worker', sequence: 3, type: 'worker.claimed' }),
      event({ id: 'evt-tool', sequence: 4, type: 'tool.call.requested', group: 'tool-call' }),
      event({ id: 'evt-error', sequence: 5, type: 'provider.error', status: 'failed' }),
    ], 'zh')

    expect(grouped.map((group) => group.title)).toEqual(['运行生命周期', '模型流', 'Worker/Job', '工具调用', '错误'])
  })
})
