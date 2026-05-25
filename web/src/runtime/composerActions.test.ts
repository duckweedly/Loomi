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
    expect(deriveComposerActions({ threadSelected: true, text: ' hello ', run: null, messages: [] })).toMatchObject({ canSend: true, canContinue: true, disabledReason: null })
    expect(deriveComposerActions({ threadSelected: true, text: '   ', run: null, messages: [] })).toMatchObject({ canSend: false, canContinue: false, disabledReason: 'no-valid-prompt' })
    expect(deriveComposerActions({ threadSelected: false, text: 'hello', run: null, messages: [] })).toMatchObject({ canSend: false, canContinue: false, disabledReason: 'no-valid-prompt' })
  })

  test('blocks send continue retry and regenerate while a selected run is active', () => {
    for (const status of ['pending', 'queued', 'running', 'retrying', 'recovering'] as const) {
      expect(deriveComposerActions({ threadSelected: true, text: 'hello', run: { ...run, status }, messages: [assistantMessage] })).toMatchObject({
        canSend: false,
        canContinue: false,
        canRetry: false,
        canRegenerate: false,
      })
    }
  })

  test('blocks all send retry and regenerate actions when provider is unavailable', () => {
    expect(deriveComposerActions({ threadSelected: true, text: 'hello', run: { ...run, status: 'failed' }, messages: [assistantMessage], providerUnavailable: true })).toMatchObject({
      canSend: false,
      canContinue: false,
      canRetry: false,
      canRegenerate: false,
      disabledReason: 'provider-unavailable',
    })
  })

  test('enables stop retry and regenerate from the correct state matrix context', () => {
    expect(deriveComposerActions({ threadSelected: true, text: '', run: { ...run, status: 'queued' }, messages: [] }).canStop).toBe(true)
    expect(deriveComposerActions({ threadSelected: true, text: '', run: { ...run, status: 'running' }, messages: [] }).canStop).toBe(true)
    expect(deriveComposerActions({ threadSelected: true, text: '', run: { ...run, status: 'retrying' }, messages: [] }).canStop).toBe(true)
    expect(deriveComposerActions({ threadSelected: true, text: '', run: { ...run, status: 'recovering' }, messages: [] }).canStop).toBe(true)
    expect(deriveComposerActions({ threadSelected: true, text: '', run: { ...run, status: 'failed' }, messages: [] }).canRetry).toBe(true)
    expect(deriveComposerActions({ threadSelected: true, text: '', run, messages: [assistantMessage] }).canRegenerate).toBe(true)
    expect(deriveComposerActions({ threadSelected: true, text: '', run: { ...run, status: 'cancelled' }, messages: [assistantMessage] }).canRegenerate).toBe(true)
    expect(deriveComposerActions({ threadSelected: true, text: '', run: { ...run, status: 'stopped' }, messages: [assistantMessage] }).canRegenerate).toBe(true)
  })
})
