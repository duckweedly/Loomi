import { describe, expect, test } from 'bun:test'
import { shouldApplyRuntimeEvent } from './state'

describe('shouldApplyRuntimeEvent', () => {
  test('rejects events from old selected threads and superseded runs', () => {
    expect(shouldApplyRuntimeEvent({ requestedThreadId: 'thread-a', currentSelectedThreadId: 'thread-b', runId: 'run-a', activeRunId: 'run-a' })).toBe(false)
    expect(shouldApplyRuntimeEvent({ requestedThreadId: 'thread-a', currentSelectedThreadId: 'thread-a', runId: 'run-a', activeRunId: 'run-b' })).toBe(false)
  })

  test('accepts events for the current selected thread and active run', () => {
    expect(shouldApplyRuntimeEvent({ requestedThreadId: 'thread-a', currentSelectedThreadId: 'thread-a', runId: 'run-a', activeRunId: 'run-a' })).toBe(true)
  })
})
