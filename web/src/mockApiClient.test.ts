import { describe, expect, test } from 'bun:test'
import { mockApiClient, setMockRuntimeScript } from './mockApiClient'

describe('mockApiClient thread runs', () => {
  test('creates a retrievable idle run for a new mock thread', async () => {
    const thread = await mockApiClient.createThread?.('New thread', 'chat')

    expect(thread).toBeDefined()
    const run = await mockApiClient.getThreadRun(thread!.id)

    expect(run).toMatchObject({
      threadId: thread!.id,
      status: 'completed',
      model: 'Claude Sonnet',
      context: 'Ready',
      events: [],
    })
  })

  test('stops a seeded mock run without adapter store coupling', async () => {
    const run = await mockApiClient.stopRun('run-1')

    expect(run.status).toBe('stopped')
    expect(run.events.at(-1)?.type).toBe('run.stopped')
  })

  test('starts a mock run before terminal script events complete', async () => {
    const thread = await mockApiClient.createThread?.('New thread', 'chat')

    const result = await mockApiClient.sendMessage(thread!.id, 'hello')

    expect(result.run.status).toBe('running')
    expect(result.run.assistantDraft).toMatchObject({ content: '', status: 'pending' })
    expect(result.messages.filter((message) => message.role === 'assistant' && message.threadId === thread!.id)).toHaveLength(0)
  })

  test('starts a subscribable mock run for retry and regenerate flows', async () => {
    const thread = await mockApiClient.createThread?.('Regenerate thread', 'chat')
    const run = await mockApiClient.startRun?.(thread!.id)
    const eventTypes: string[] = []

    mockApiClient.subscribeRunEvents?.(run!.id, 0, (event) => eventTypes.push(event.type), () => {})
    await new Promise((resolve) => setTimeout(resolve, 850))
    const storedRun = await mockApiClient.getThreadRun(thread!.id)

    expect(run?.status).toBe('running')
    expect(run?.id).not.toBe('run-1')
    expect(eventTypes).toContain('run.completed')
    expect(storedRun.id).toBe(run?.id)
    expect(storedRun.status).toBe('completed')
  })

  test('runs the failure script without duplicate failed terminal events or assistant success replies', async () => {
    const thread = await mockApiClient.createThread?.('Failure script thread', 'chat')

    setMockRuntimeScript('failure')
    const result = await mockApiClient.sendMessage(thread!.id, 'fail please')
    mockApiClient.subscribeRunEvents?.(result.run.id, 0, () => {}, () => {})
    await new Promise((resolve) => setTimeout(resolve, 80))
    setMockRuntimeScript('success')
    const run = await mockApiClient.getThreadRun(thread!.id)
    const messages = await mockApiClient.getThreadMessages(thread!.id)

    expect(result.run.status).toBe('running')
    expect(run.status).toBe('failed')
    expect(run.events.filter((event) => event.type === 'run.failed')).toHaveLength(1)
    expect(messages.filter((message) => message.role === 'assistant' && message.threadId === thread!.id)).toHaveLength(0)
  })

  test('applies mock runtime scenario changes only to future sends', async () => {
    const first = await mockApiClient.createThread?.('Scenario failure', 'chat')
    const second = await mockApiClient.createThread?.('Scenario success', 'chat')

    setMockRuntimeScript('failure')
    const failed = await mockApiClient.sendMessage(first!.id, 'fail later')
    mockApiClient.subscribeRunEvents?.(failed.run.id, 0, () => {}, () => {})
    setMockRuntimeScript('success')
    const completed = await mockApiClient.sendMessage(second!.id, 'succeed now')
    mockApiClient.subscribeRunEvents?.(completed.run.id, 0, () => {}, () => {})
    await new Promise((resolve) => setTimeout(resolve, 850))

    expect(failed.run.status).toBe('running')
    expect((await mockApiClient.getThreadRun(first!.id)).status).toBe('failed')
    expect((await mockApiClient.getThreadRun(second!.id)).status).toBe('completed')
  })

  test('lists only active mock threads after archiving', async () => {
    const thread = await mockApiClient.createThread?.('Archive candidate', 'chat')

    await mockApiClient.archiveThread?.(thread!.id)

    expect((await mockApiClient.listThreads()).some((item) => item.id === thread!.id)).toBe(false)
  })

  test('uses stable unique ids for rapid mock thread creation', async () => {
    const first = await mockApiClient.createThread?.('Rapid thread', 'chat')
    const second = await mockApiClient.createThread?.('Rapid thread', 'chat')

    expect(first!.id).not.toBe(second!.id)
  })
})
