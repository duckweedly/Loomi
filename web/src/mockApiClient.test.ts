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

  test('persists the completed M3.5 mock run created by the first message in a new mock thread', async () => {
    const thread = await mockApiClient.createThread?.('New thread', 'chat')

    await mockApiClient.sendMessage(thread!.id, 'hello')
    const run = await mockApiClient.getThreadRun(thread!.id)

    expect(run.status).toBe('completed')
    expect(run.events.map((event) => event.type)).toContain('run.completed')
  })

  test('runs the failure script without duplicate failed terminal events or assistant success replies', async () => {
    const thread = await mockApiClient.createThread?.('Failure script thread', 'chat')

    setMockRuntimeScript('failure')
    const result = await mockApiClient.sendMessage(thread!.id, 'fail please')
    setMockRuntimeScript('success')

    expect(result.run.status).toBe('failed')
    expect(result.run.events.filter((event) => event.type === 'run.failed')).toHaveLength(1)
    expect(result.messages.filter((message) => message.role === 'assistant' && message.threadId === thread!.id)).toHaveLength(0)
  })

  test('applies mock runtime scenario changes only to future sends', async () => {
    const first = await mockApiClient.createThread?.('Scenario failure', 'chat')
    const second = await mockApiClient.createThread?.('Scenario success', 'chat')

    setMockRuntimeScript('failure')
    const failed = await mockApiClient.sendMessage(first!.id, 'fail later')
    setMockRuntimeScript('success')
    const completed = await mockApiClient.sendMessage(second!.id, 'succeed now')

    expect(failed.run.status).toBe('failed')
    expect(completed.run.status).toBe('completed')
    expect((await mockApiClient.getThreadRun(first!.id)).status).toBe('failed')
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
