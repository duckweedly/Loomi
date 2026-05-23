import { describe, expect, test } from 'bun:test'
import { mockApiClient } from './mockApiClient'

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

  test('persists the run event created by the first message in a new mock thread', async () => {
    const thread = await mockApiClient.createThread?.('New thread', 'chat')

    await mockApiClient.sendMessage(thread!.id, 'hello')
    const run = await mockApiClient.getThreadRun(thread!.id)

    expect(run.status).toBe('running')
    expect(run.events.at(-1)).toMatchObject({
      type: 'message.queued',
      status: 'running',
    })
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
