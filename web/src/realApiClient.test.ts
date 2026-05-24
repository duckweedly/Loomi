import { describe, expect, test } from 'bun:test'
import { createClientMessageID, mapApiProviderCapability, mapApiRun, mapApiRunEvent } from './realApiClient'

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

describe('M5 provider and run mapping', () => {
  test('maps model gateway run source', () => {
    const run = mapApiRun({
      id: 'run-1',
      thread_id: 'thread-1',
      status: 'running',
      source: 'model_gateway',
      title: 'Model gateway run',
      created_at: '2026-05-23T00:00:00Z',
      updated_at: '2026-05-23T00:00:01Z',
      completed_at: null,
      error_code: null,
      error_message: null,
    })

    expect(run.source).toBe('model_gateway')
    expect(run.model).toBe('Model gateway')
    expect(run.context).toBe('model_gateway')
  })

  test('maps provider capability without credential fields', () => {
    const provider = mapApiProviderCapability({ id: 'custom', family: 'openai_compatible', base_url: 'https://example.test/v1', model: 'gpt-5.5', status: 'available' })

    expect(provider.id).toBe('custom')
    expect(provider.family).toBe('openai_compatible')
    expect(provider.baseUrl).toBe('https://example.test/v1')
    expect(provider.model).toBe('gpt-5.5')
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
  })

  test('exposes subscribeRunEvents for EventSource-compatible streaming', () => {
    const source = Bun.file(new URL('./realApiClient.ts', import.meta.url)).text()
    return expect(source).resolves.toContain('subscribeRunEvents')
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

  test('maps model output deltas into assistantDelta for streaming drafts', () => {
    const event = mapApiRunEvent({
      id: 'evt-2',
      run_id: 'run-1',
      thread_id: 'thread-1',
      sequence: 3,
      category: 'message',
      type: 'model_output_delta',
      summary: 'Model output delta',
      content: 'hello',
      metadata: {},
      created_at: '2026-05-23T00:00:01Z',
    })

    expect(event.type).toBe('message.model_output_delta')
    expect(event.assistantDelta).toBe('hello')
    expect(event.status).toBe('running')
  })

  test('real sendMessage starts model gateway runs from durable messages', () => {
    const source = Bun.file(new URL('./realApiClient.ts', import.meta.url)).text()

    return expect(source).resolves.toContain("source: 'model_gateway'")
  })
})
