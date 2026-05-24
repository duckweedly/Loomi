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
    expect(events).toEqual(['run.created', 'run.queued', 'job.claimed', 'pipeline.step.started', 'pipeline.step.completed', 'context.loading', 'assistant.thinking', 'assistant.drafting', 'assistant.message.completed', 'run.completed'])
  })

  test('accumulates assistant draft and completes exactly one assistant message', async () => {
    const adapter = createMockExecutionAdapter()
    const message = await adapter.sendMessage('thread-a', 'hello')
    const run = await adapter.createRun('thread-a', message.id, 'success')

    const drafting = await adapter.appendAssistantDelta('thread-a', run.id, 'A')
    const completed = await adapter.completeRun('thread-a', run.id, 'Answer')

    expect(drafting.assistantDraft).toMatchObject({ content: 'A', status: 'streaming' })
    expect(completed.run.assistantDraft).toMatchObject({ content: 'Answer', status: 'completed', messageId: completed.message.id })
    expect(completed.run.status).toBe('completed')
    expect(completed.message).toMatchObject({ threadId: 'thread-a', role: 'assistant', content: 'Answer', runId: run.id })
  })

  test('fails and stops without successful assistant completion', async () => {
    const adapter = createMockExecutionAdapter()
    const message = await adapter.sendMessage('thread-a', 'hello')
    const failedRun = await adapter.createRun('thread-a', message.id, 'failure')
    const stoppedRun = await adapter.createRun('thread-a', message.id, 'success')
    await adapter.appendAssistantDelta('thread-a', stoppedRun.id, 'partial')

    expect(await adapter.failRun('thread-a', failedRun.id, 'boom')).toMatchObject({ status: 'failed', assistantDraft: { status: 'failed' } })
    expect(await adapter.stopRun('thread-a', stoppedRun.id)).toMatchObject({ status: 'stopped', assistantDraft: { content: 'partial', status: 'stopped' } })
  })

  test('supports regenerated assistant attempt metadata', async () => {
    const adapter = createMockExecutionAdapter()
    const message = await adapter.sendMessage('thread-a', 'hello')
    const run = await adapter.createRun('thread-a', message.id, 'success', { attemptOfMessageId: 'assistant-a' })
    const completed = await adapter.completeRun('thread-a', run.id, 'Answer')

    expect(completed.message).toMatchObject({ runId: run.id, attemptOfMessageId: 'assistant-a' })
  })

  test('scripted model stream updates assistant draft through terminal states', async () => {
    const adapter = createMockExecutionAdapter()
    const message = await adapter.sendMessage('thread-a', 'hello')
    const run = await adapter.createRun('thread-a', message.id, 'model-error')
    const statuses: string[] = []

    await adapter.subscribeRunEvents('thread-a', run.id, (event) => statuses.push(event.status))

    expect(statuses).toContain('failed')
    await expect(adapter.completeRun('thread-a', run.id, 'stale final')).rejects.toThrow('Run is terminal')
  })

  test('does not finalize a stopped run from a stale completion', async () => {
    const adapter = createMockExecutionAdapter()
    const message = await adapter.sendMessage('thread-a', 'hello')
    const run = await adapter.createRun('thread-a', message.id, 'success')
    await adapter.appendAssistantDelta('thread-a', run.id, 'partial')
    await adapter.stopRun('thread-a', run.id)

    await expect(adapter.completeRun('thread-a', run.id, 'stale final')).rejects.toThrow('Run is terminal')
  })
})
