import { describe, expect, test } from 'bun:test'
import { createMockExecutionAdapter } from './mockExecutionAdapter'

describe('createMockExecutionAdapter', () => {
  test('sends a visible user message before creating a run', async () => {
    const adapter = createMockExecutionAdapter()

    const message = await adapter.sendMessage('thread-a', 'hello')

    expect(message).toMatchObject({ threadId: 'thread-a', role: 'user', content: 'hello' })
  })

  test('creates independent success runs and emits ordered events', async () => {
    const adapter = createMockExecutionAdapter()
    const firstMessage = await adapter.sendMessage('thread-a', 'first')
    const secondMessage = await adapter.sendMessage('thread-a', 'second')

    const firstRun = await adapter.createRun('thread-a', firstMessage.id, 'success')
    const secondRun = await adapter.createRun('thread-a', secondMessage.id, 'success')
    const events: string[] = []
    await adapter.subscribeRunEvents('thread-a', firstRun.id, (event) => events.push(event.type))

    expect(firstRun.id).not.toBe(secondRun.id)
    expect(events).toEqual(['run.created', 'context.loading', 'assistant.thinking', 'assistant.drafting', 'assistant.message.completed', 'run.completed'])
  })

  test('accumulates assistant draft and completes exactly one assistant message', async () => {
    const adapter = createMockExecutionAdapter()
    const message = await adapter.sendMessage('thread-a', 'hello')
    const run = await adapter.createRun('thread-a', message.id, 'success')

    const drafting = await adapter.appendAssistantDelta('thread-a', run.id, 'A')
    const completed = await adapter.completeRun('thread-a', run.id, 'Answer')

    expect(drafting.assistantDraft).toMatchObject({ content: 'A', status: 'drafting' })
    expect(completed.run.status).toBe('completed')
    expect(completed.message).toMatchObject({ threadId: 'thread-a', role: 'assistant', content: 'Answer' })
  })

  test('fails and stops without successful assistant completion', async () => {
    const adapter = createMockExecutionAdapter()
    const message = await adapter.sendMessage('thread-a', 'hello')
    const failedRun = await adapter.createRun('thread-a', message.id, 'failure')
    const stoppedRun = await adapter.createRun('thread-a', message.id, 'success')

    expect(await adapter.failRun('thread-a', failedRun.id, 'boom')).toMatchObject({ status: 'failed' })
    expect(await adapter.stopRun('thread-a', stoppedRun.id)).toMatchObject({ status: 'stopped' })
  })
})
