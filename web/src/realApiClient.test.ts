import { describe, expect, test } from 'bun:test'
import { createClientMessageID, mapApiRun, mapApiRunEvent } from './realApiClient'

describe('createClientMessageID', () => {
  test('does not rely on Date.now alone', () => {
    const originalNow = Date.now
    Date.now = () => 123
    try {
      const first = createClientMessageID()
      const second = createClientMessageID()

      expect(first).toStartWith('web-123-')
      expect(second).toStartWith('web-123-')
      expect(second).not.toBe(first)
    } finally {
      Date.now = originalNow
    }
  })
})

describe('M4 run mapping', () => {
  test('maps local simulated run status without LLM/tool claims', () => {
    const run = mapApiRun({
      id: 'run-1',
      thread_id: 'thread-1',
      status: 'running',
      source: 'local_simulated',
      title: 'Local simulated run',
      created_at: '2026-05-23T00:00:00Z',
      updated_at: '2026-05-23T00:00:01Z',
      completed_at: null,
      error_code: null,
      error_message: null,
    })

    expect(run.status).toBe('running')
    expect(run.model).toBe('Local simulated')
    expect(run.context).toBe('local_simulated')
    expect(run.assistantDraft).toMatchObject({ content: '', status: 'pending' })
  })

  test('maps real model source and recovering events', () => {
    const run = mapApiRun({
      id: 'run-1',
      thread_id: 'thread-1',
      status: 'running',
      source: 'real_model',
      title: 'Real model run',
      created_at: '2026-05-23T00:00:00Z',
      updated_at: '2026-05-23T00:00:01Z',
      completed_at: null,
      error_code: null,
      error_message: null,
    }, [mapApiRunEvent({ id: 'evt-recovering', run_id: 'run-1', thread_id: 'thread-1', sequence: 1, category: 'lifecycle', type: 'run.recovering', summary: 'Recovering', content: null, metadata: {}, created_at: '2026-05-23T00:00:00Z' })])

    expect(run.model).toBe('Real model')
    expect(run.context).toBe('real_model')
    expect(run.status).toBe('recovering')
    expect(run.assistantDraft).toMatchObject({ status: 'recovering' })
  })

  test('restores assistant draft from loaded model event history', () => {
    const events = [
      mapApiRunEvent({ id: 'evt-delta', run_id: 'run-1', thread_id: 'thread-1', sequence: 1, category: 'message', type: 'model.delta', summary: 'Delta', content: 'Hel', metadata: {}, created_at: '2026-05-23T00:00:00Z' }),
      mapApiRunEvent({ id: 'evt-delta-2', run_id: 'run-1', thread_id: 'thread-1', sequence: 2, category: 'message', type: 'model.delta', summary: 'Delta', content: 'lo', metadata: {}, created_at: '2026-05-23T00:00:01Z' }),
    ]

    const run = mapApiRun({
      id: 'run-1',
      thread_id: 'thread-1',
      status: 'running',
      source: 'local_simulated',
      title: 'Local simulated run',
      created_at: '2026-05-23T00:00:00Z',
      updated_at: '2026-05-23T00:00:01Z',
      completed_at: null,
      error_code: null,
      error_message: null,
    }, events)

    expect(run.assistantDraft).toMatchObject({ content: 'Hello', status: 'streaming', lastEventId: 'evt-delta-2' })
  })

  test('does not restore late final events over terminal stopped history', () => {
    const events = [
      mapApiRunEvent({ id: 'evt-delta', run_id: 'run-1', thread_id: 'thread-1', sequence: 1, category: 'message', type: 'model.delta', summary: 'Delta', content: 'Partial', metadata: {}, created_at: '2026-05-23T00:00:00Z' }),
      mapApiRunEvent({ id: 'evt-stopped', run_id: 'run-1', thread_id: 'thread-1', sequence: 2, category: 'lifecycle', type: 'run.stopped', summary: 'Stopped', content: null, metadata: {}, created_at: '2026-05-23T00:00:01Z' }),
      mapApiRunEvent({ id: 'evt-final', run_id: 'run-1', thread_id: 'thread-1', sequence: 3, category: 'final', type: 'model.final', summary: 'Final', content: 'Final', metadata: {}, created_at: '2026-05-23T00:00:02Z' }),
    ]

    const run = mapApiRun({
      id: 'run-1',
      thread_id: 'thread-1',
      status: 'running',
      source: 'local_simulated',
      title: 'Local simulated run',
      created_at: '2026-05-23T00:00:00Z',
      updated_at: '2026-05-23T00:00:02Z',
      completed_at: null,
      error_code: null,
      error_message: null,
    }, events)

    expect(run.status).toBe('stopped')
    expect(run.assistantDraft).toMatchObject({ content: 'Partial', status: 'stopped', lastEventId: 'evt-stopped' })
  })

  test('exposes subscribeRunEvents for EventSource-compatible streaming', () => {
    const source = Bun.file(new URL('./realApiClient.ts', import.meta.url)).text()
    return expect(source).resolves.toContain('subscribeRunEvents')
  })

  test('maps model delta final and error events into assistant draft signals', () => {
    const delta = mapApiRunEvent({ id: 'evt-delta', run_id: 'run-1', thread_id: 'thread-1', sequence: 3, category: 'message', type: 'model.delta', summary: 'Delta', content: 'Hel', metadata: {}, created_at: '2026-05-23T00:00:00Z' })
    const final = mapApiRunEvent({ id: 'evt-final', run_id: 'run-1', thread_id: 'thread-1', sequence: 4, category: 'final', type: 'model.final', summary: 'Final', content: 'Hello', metadata: {}, created_at: '2026-05-23T00:00:01Z' })
    const error = mapApiRunEvent({ id: 'evt-error', run_id: 'run-1', thread_id: 'thread-1', sequence: 5, category: 'error', type: 'model.error', summary: 'Provider failed', content: null, metadata: {}, created_at: '2026-05-23T00:00:02Z' })

    expect(delta.type).toBe('model.delta')
    expect(delta.assistantDelta).toBe('Hel')
    expect(delta.status).toBe('running')
    expect(final.type).toBe('model.final')
    expect(final.content).toBe('Hello')
    expect(final.status).toBe('completed')
    expect(error.type).toBe('model.error')
    expect(error.status).toBe('failed')
  })

  test('restores assistant draft from current M4 local simulated event vocabulary', () => {
    const events = [
      mapApiRunEvent({ id: 'evt-drafting', run_id: 'run-1', thread_id: 'thread-1', sequence: 1, category: 'progress', type: 'drafting', summary: 'Drafting response', content: null, metadata: {}, created_at: '2026-05-23T00:00:00Z' }),
      mapApiRunEvent({ id: 'evt-message', run_id: 'run-1', thread_id: 'thread-1', sequence: 2, category: 'message', type: 'assistant_message', summary: 'Simulated response', content: 'Local simulated response', metadata: {}, created_at: '2026-05-23T00:00:01Z' }),
      mapApiRunEvent({ id: 'evt-final', run_id: 'run-1', thread_id: 'thread-1', sequence: 3, category: 'final', type: 'run_completed', summary: 'Run completed', content: null, metadata: {}, created_at: '2026-05-23T00:00:02Z' }),
    ]

    const run = mapApiRun({
      id: 'run-1',
      thread_id: 'thread-1',
      status: 'running',
      source: 'local_simulated',
      title: 'Local simulated run',
      created_at: '2026-05-23T00:00:00Z',
      updated_at: '2026-05-23T00:00:02Z',
      completed_at: null,
      error_code: null,
      error_message: null,
    }, events)

    expect(events.map((event) => event.type)).toEqual(['assistant.drafting', 'assistant.message.completed', 'run.completed'])
    expect(run.status).toBe('completed')
    expect(run.assistantDraft).toMatchObject({ content: 'Local simulated response', status: 'completed', lastEventId: 'evt-final' })
  })

  test('maps lifecycle progress message error and final event categories', () => {
    const event = mapApiRunEvent({
      id: 'evt-1',
      run_id: 'run-1',
      thread_id: 'thread-1',
      sequence: 2,
      category: 'progress',
      type: 'context_loaded',
      summary: 'Context loaded',
      content: null,
      metadata: {},
      created_at: '2026-05-23T00:00:00Z',
    })

    expect(event.type).toBe('progress.context_loaded')
    expect(event.label).toBe('progress')
    expect(event.detail).toBe('Context loaded')
    expect(event.status).toBe('running')
  })

  test('preserves token usage and provider metadata as event details', () => {
    const usage = mapApiRunEvent({ id: 'evt-usage', run_id: 'run-1', thread_id: 'thread-1', sequence: 6, category: 'message', type: 'model.usage', summary: 'Usage', content: null, metadata: { input_tokens: 5, output_tokens: 8, total_tokens: 13 }, created_at: '2026-05-23T00:00:03Z' })
    const providerError = mapApiRunEvent({ id: 'evt-provider', run_id: 'run-1', thread_id: 'thread-1', sequence: 7, category: 'error', type: 'provider.error', summary: 'Provider unavailable', content: null, metadata: { provider: 'anthropic', code: 'overloaded' }, created_at: '2026-05-23T00:00:04Z' })

    expect(usage.group).toBe('model-stream')
    expect(usage.usage).toEqual({ inputTokens: 5, outputTokens: 8, totalTokens: 13 })
    expect(usage.detail).not.toContain('input_tokens')
    expect(providerError.group).toBe('error')
    expect(providerError.severity).toBe('error')
    expect(providerError.detail).toContain('anthropic')
    expect(providerError.detail).toContain('overloaded')
  })

  test('preserves canonical dotted worker and backend event types from real API events', () => {
    const worker = mapApiRunEvent({ id: 'evt-worker', run_id: 'run-1', thread_id: 'thread-1', sequence: 8, category: 'progress', type: 'worker.claimed', summary: 'Worker claimed', content: null, metadata: {}, created_at: '2026-05-23T00:00:05Z' })
    const backend = mapApiRunEvent({ id: 'evt-backend', run_id: 'run-1', thread_id: 'thread-1', sequence: 9, category: 'progress', type: 'backend.unavailable', summary: 'Backend unavailable', content: null, metadata: {}, created_at: '2026-05-23T00:00:06Z' })

    expect(worker.type).toBe('worker.claimed')
    expect(worker.group).toBe('worker-job')
    expect(backend.type).toBe('backend.unavailable')
    expect(backend.group).toBe('error')
    expect(backend.severity).toBe('error')
  })
})
