import { describe, expect, test } from 'bun:test'
import { createRuntimeEvent, getRuntimeScript, getRuntimeScriptSteps, runtimeScripts } from './runtimeScripts'

describe('runtimeScripts', () => {
  test('defines success script with six ordered runtime milestones and final assistant content', () => {
    expect(getRuntimeScriptSteps('success').map((step) => step.type)).toEqual([
      'run.created',
      'context.loading',
      'assistant.thinking',
      'assistant.drafting',
      'assistant.message.completed',
      'run.completed',
    ])
    expect(getRuntimeScript('success').finalAssistantMessage).toBeTruthy()
    expect(runtimeScripts.success.terminalStatus).toBe('completed')
  })

  test('defines failure script without final assistant content', () => {
    expect(getRuntimeScriptSteps('failure').map((step) => step.type)).toEqual([
      'run.created',
      'context.loading',
      'assistant.thinking',
      'run.failed',
    ])
    expect(getRuntimeScript('failure').finalAssistantMessage).toBeUndefined()
    expect(runtimeScripts.failure.terminalStatus).toBe('failed')
  })

  test('creates per-run events with thread and run identifiers', () => {
    expect(createRuntimeEvent({ threadId: 'thread-a', runId: 'run-a', sequence: 2, step: getRuntimeScriptSteps('success')[2] })).toMatchObject({
      id: 'run-a-evt-2',
      threadId: 'thread-a',
      runId: 'run-a',
      type: 'assistant.thinking',
      status: 'running',
    })
  })
})
