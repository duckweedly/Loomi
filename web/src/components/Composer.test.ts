import { describe, expect, test } from 'bun:test'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { createElement } from 'react'
import { renderToStaticMarkup } from 'react-dom/server'
import { Composer } from './Composer'

describe('Composer interactions', () => {
  test('keeps retry and regenerate out of the input composer', () => {
    const html = renderToStaticMarkup(createElement(Composer, {
      threadSelected: true,
      run: { id: 'run-a', threadId: 'thread-a', status: 'failed', model: 'Mock', context: 'M3.5 mock', events: [] },
      messages: [{ id: 'msg-a', threadId: 'thread-a', role: 'assistant', content: 'done', createdAt: 'Now' }],
      onSubmit: () => {},
      onStop: () => {},
      onRetry: () => {},
      onRegenerate: () => {},
    }))

    expect(html).not.toContain('Retry')
    expect(html).not.toContain('Regenerate')
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
    expect(source).toContain('if (composerDisabled || !canSubmit || (!content && !hasAttachments)) return')
  })

  test('uses animal-island-ui Button directly for composer primary action', () => {
    const source = readFileSync(resolve(import.meta.dir, 'Composer.tsx'), 'utf8')

    expect(source).toContain("import { Button, Select } from 'animal-island-ui'")
    expect(source).not.toContain("import { Button } from '@lobehub/ui'")
  })

  test('uses animal-island-ui Select directly for composer model choice', () => {
    const source = readFileSync(resolve(import.meta.dir, 'Composer.tsx'), 'utf8')

    expect(source).toContain('<Select disabled={modelOptions.length === 0}')
    expect(source).not.toContain('<select')
  })

  test('renders attachment and model affordances without unused persona or voice controls', () => {
    const html = renderToStaticMarkup(createElement(Composer, {
      threadSelected: true,
      run: undefined,
      messages: [],
      modelOptions: [{ key: 'custom:gpt-5.5', providerId: 'custom', model: 'gpt-5.5', label: 'gpt-5.5 · openai_compatible' }],
      onSubmit: () => {},
      onStop: () => {},
      onRetry: () => {},
      onRegenerate: () => {},
    }))

    expect(html).not.toContain('aria-label="Persona"')
    expect(html).toContain('composer-model-select')
    expect(html).toContain('gpt-5.5 · openai_compatible')
    expect(html).toContain('type="file"')
    expect(html).toContain('accept="image/*,.pdf')
    expect(html).not.toContain('lucide-mic')
  })

  test('renders one placeholder and surfaces folder state without a mode switch', () => {
    const chatHtml = renderToStaticMarkup(createElement(Composer, {
      threadSelected: true,
      run: null,
      messages: [],
      workspaceFolderStatus: 'No folder selected',
      onSubmit: () => {},
      onChooseWorkspaceFolder: () => {},
    }))
    const workHtml = renderToStaticMarkup(createElement(Composer, {
      threadSelected: true,
      run: null,
      messages: [],
      dataSourceMode: 'mock',
      workspaceFolderStatus: 'No folder selected',
      onSubmit: () => {},
      onChooseWorkspaceFolder: () => {},
    }))

    expect(chatHtml).toContain('placeholder="Message Loomi"')
    expect(chatHtml).not.toContain('Work tools limited')
    expect(chatHtml).toContain('选择目录')
    expect(chatHtml).toContain('No folder selected')
    expect(workHtml).toContain('placeholder="Message Loomi"')
    expect(workHtml).not.toContain('Work tools limited')
    expect(workHtml).not.toContain('Mock demo mode')
    expect(workHtml).toContain('选择目录')
    expect(workHtml).toContain('No folder selected')
  })
})
