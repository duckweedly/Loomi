import { describe, expect, test } from 'bun:test'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { createElement } from 'react'
import { renderToStaticMarkup } from 'react-dom/server'
import { Composer } from './Composer'

describe('Composer interactions', () => {
  test('renders retry and regenerate controls when enabled', () => {
    const html = renderToStaticMarkup(createElement(Composer, {
      threadSelected: true,
      run: { id: 'run-a', threadId: 'thread-a', status: 'failed', model: 'Mock', context: 'M3.5 mock', events: [] },
      messages: [{ id: 'msg-a', threadId: 'thread-a', role: 'assistant', content: 'done', createdAt: 'Now' }],
      onSubmit: () => {},
      onStop: () => {},
      onRetry: () => {},
      onRegenerate: () => {},
    }))

    expect(html).toContain('Retry')
    expect(html).toContain('Regenerate')
  })

  test('renders stop control for an active run', () => {
    const html = renderToStaticMarkup(createElement(Composer, {
      threadSelected: true,
      run: { id: 'run-a', threadId: 'thread-a', status: 'running', model: 'Mock', context: 'M3.5 mock', events: [] },
      messages: [],
      onSubmit: () => {},
      onStop: () => {},
      onRetry: () => {},
      onRegenerate: () => {},
    }))

    expect(html).toContain('Stop')
  })

  test('disables submit while active run blocks send', () => {
    const html = renderToStaticMarkup(createElement(Composer, {
      threadSelected: true,
      run: { id: 'run-a', threadId: 'thread-a', status: 'running', model: 'Mock', context: 'M3.5 mock', events: [] },
      messages: [],
      onSubmit: () => {},
      onStop: () => {},
      onRetry: () => {},
      onRegenerate: () => {},
    }))

    expect(html).toContain('disabled')
  })

  test('uses Enter for submit and Shift+Enter for newline', () => {
    const source = readFileSync(resolve(import.meta.dir, 'Composer.tsx'), 'utf8')

    expect(source).toContain("event.key === 'Enter'")
    expect(source).toContain('!event.shiftKey')
  })

  test('derives action guards from current textarea value and blocks disabled submit paths', () => {
    const source = readFileSync(resolve(import.meta.dir, 'Composer.tsx'), 'utf8')

    expect(source).toContain('deriveComposerActions({ threadSelected, text: value, run, messages, providerUnavailable })')
    expect(source).toContain('if (composerDisabled || !canSubmit || !content) return')
  })
})
