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

  test('defines foundational model, stopped, and replay scripts', () => {
    expect(getRuntimeScriptSteps('model-stream').map((step) => step.type)).toEqual([
      'run.created',
      'job.queued',
      'worker.claimed',
      'job.retrying',
      'model.delta',
      'model.delta',
      'model.final',
      'run.completed',
    ])
    expect(runtimeScripts['model-error'].terminalStatus).toBe('failed')
    expect(getRuntimeScriptSteps('stopped').map((step) => step.type)).toContain('run.stopped')
    expect(getRuntimeScriptSteps('replayed').map((step) => step.type)).toEqual(['run.created', 'model.delta', 'model.delta', 'model.final', 'run.completed'])
  })

  test('creates grouped events with severity and usage metadata', () => {
    const modelFinalStep = getRuntimeScriptSteps('model-stream').find((step) => step.type === 'model.final')!
    const event = createRuntimeEvent({ threadId: 'thread-a', runId: 'run-a', sequence: 6, step: modelFinalStep })

    expect(event).toMatchObject({
      type: 'model.final',
      group: 'model-stream',
      severity: 'info',
      usage: { inputTokens: 11, outputTokens: 22 },
    })
  })

  test('defines grouped worker retry cancel and provider error mock scripts', () => {
    const modelStreamTypes = getRuntimeScriptSteps('model-stream').map((step) => step.type)
    const modelErrorTypes = getRuntimeScriptSteps('model-error').map((step) => step.type)
    const stoppedTypes = getRuntimeScriptSteps('stopped').map((step) => step.type)

    expect(modelStreamTypes).toContain('job.queued')
    expect(modelStreamTypes).toContain('worker.claimed')
    expect(modelStreamTypes).toContain('job.retrying')
    expect(modelErrorTypes).toContain('provider.error')
    expect(stoppedTypes).toContain('run.cancelled')
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
