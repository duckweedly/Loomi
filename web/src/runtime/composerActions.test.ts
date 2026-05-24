import { describe, expect, test } from 'bun:test'
import type { Message, Run } from '../domain'
import { deriveComposerActions } from './composerActions'

const run: Run = {
  id: 'run-a',
  threadId: 'thread-a',
  status: 'completed',
  model: 'Local simulated',
  context: 'local_simulated',
  events: [],
}

const assistantMessage: Message = {
  id: 'msg-a',
  threadId: 'thread-a',
  role: 'assistant',
  content: 'done',
  createdAt: 'Now',
  runId: 'run-a',
}

describe('composer actions', () => {
  test('enables send and continue only for non-empty text without active run', () => {
    expect(deriveComposerActions({ threadSelected: true, text: ' hello ', run: null, messages: [] })).toMatchObject({ canSend: true, canContinue: true })
    expect(deriveComposerActions({ threadSelected: true, text: '   ', run: null, messages: [] })).toMatchObject({ canSend: false, canContinue: false })
    expect(deriveComposerActions({ threadSelected: false, text: 'hello', run: null, messages: [] })).toMatchObject({ canSend: false, canContinue: false })
  })

  test('blocks send continue retry and regenerate while a selected run is active', () => {
    for (const status of ['pending', 'running', 'retrying', 'recovering'] as const) {
      expect(deriveComposerActions({ threadSelected: true, text: 'hello', run: { ...run, status }, messages: [assistantMessage] })).toMatchObject({
        canSend: false,
        canContinue: false,
        canRetry: false,
        canRegenerate: false,
      })
    }
  })

  test('enables stop retry and regenerate from the correct terminal context', () => {
    expect(deriveComposerActions({ threadSelected: true, text: '', run: { ...run, status: 'running' }, messages: [] }).canStop).toBe(true)
    expect(deriveComposerActions({ threadSelected: true, text: '', run: { ...run, status: 'failed' }, messages: [] }).canRetry).toBe(true)
    expect(deriveComposerActions({ threadSelected: true, text: '', run, messages: [assistantMessage] }).canRegenerate).toBe(true)
  })
})
