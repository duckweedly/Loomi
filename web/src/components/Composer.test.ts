import { describe, expect, test } from 'bun:test'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { createElement } from 'react'
import { renderToStaticMarkup } from 'react-dom/server'
import { Composer, getComposerContextMenuPosition } from './Composer'

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

  test('turns the primary send button into stop for an active run', () => {
    const html = renderToStaticMarkup(createElement(Composer, {
      threadSelected: true,
      run: { id: 'run-a', threadId: 'thread-a', status: 'running', model: 'Mock', context: 'M3.5 mock', events: [] },
      messages: [],
      modelOptions: [{ key: 'mock:model', providerId: 'mock', model: 'model', label: 'model' }],
      onSubmit: () => {},
      onStop: () => {},
      onRetry: () => {},
      onRegenerate: () => {},
    }))

    expect(html).toContain('aria-label="Stop"')
    expect(html).toContain('is-stopping')
    expect(html).not.toContain('composer-actions')
  })

  test('keeps the primary button enabled as stop while an active run blocks send', () => {
    const html = renderToStaticMarkup(createElement(Composer, {
      threadSelected: true,
      run: { id: 'run-a', threadId: 'thread-a', status: 'running', model: 'Mock', context: 'M3.5 mock', events: [] },
      messages: [],
      modelOptions: [{ key: 'mock:model', providerId: 'mock', model: 'model', label: 'model' }],
      onSubmit: () => {},
      onStop: () => {},
      onRetry: () => {},
      onRegenerate: () => {},
    }))

    expect(html).toContain('type="button"')
    expect(html).not.toContain('disabled')
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
    expect(source).toContain('const showStopButton = Boolean(actions.canStop && onStop)')
  })

  test('clears submitted input and restores focus to the textarea', () => {
    const source = readFileSync(resolve(import.meta.dir, 'Composer.tsx'), 'utf8')

    expect(source).toContain('const inputRef = useRef<HTMLTextAreaElement>(null)')
    expect(source).toContain('inputRef.current?.focus()')
    expect(source).toContain('setValue(\'\')')
    expect(source).toContain('setAttachments([])')
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
    expect(html).toContain('aria-label="Attach"')
    expect(html).toContain('accept="image/*,.pdf')
    expect(html).not.toContain('lucide-mic')
  })

  test('renders one placeholder and surfaces workspace state without a mode switch', () => {
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
    expect(chatHtml).toContain('No folder selected')
    expect(workHtml).toContain('placeholder="Message Loomi"')
    expect(workHtml).not.toContain('Work tools limited')
    expect(workHtml).not.toContain('Mock demo mode')
    expect(workHtml).toContain('No folder selected')

    const defaultHtml = renderToStaticMarkup(createElement(Composer, {
      threadSelected: true,
      run: null,
      messages: [],
      onSubmit: () => {},
      onChooseWorkspaceFolder: () => {},
    }))

    expect(defaultHtml).toContain('选择工作区')
  })

  test('renders workspace context as business context instead of a file picker', () => {
    const emptyHtml = renderToStaticMarkup(createElement(Composer, {
      threadSelected: true,
      run: null,
      messages: [],
      onSubmit: () => {},
      onChooseWorkspaceFolder: () => {},
    }))
    const selectedHtml = renderToStaticMarkup(createElement(Composer, {
      threadSelected: true,
      run: null,
      messages: [],
      workspaceFolderStatus: '工作区 · Loomi',
      onSubmit: () => {},
      onChooseWorkspaceFolder: () => {},
    }))

    expect(emptyHtml).toContain('composer-context-status is-empty')
    expect(emptyHtml).toContain('aria-label="业务上下文：选择工作区"')
    expect(selectedHtml).toContain('composer-context-status is-selected')
    expect(selectedHtml).toContain('aria-label="业务上下文：工作区 · Loomi"')
  })

  test('shows current workspace name as a safe title without absolute paths', () => {
    const html = renderToStaticMarkup(createElement(Composer, {
      threadSelected: true,
      run: null,
      messages: [],
      workspaceFolderStatus: '工作区 · Downloads',
      onSubmit: () => {},
      onChooseWorkspaceFolder: () => {},
    }))

    expect(html).toContain('工作区 · Downloads')
    expect(html).toContain('title="当前工作区：Downloads"')
    expect(html).not.toContain('/Users/')
  })

  test('keeps context actions behind the plus menu contract', () => {
    const source = readFileSync(resolve(import.meta.dir, 'Composer.tsx'), 'utf8')

    expect(source).toContain('composer-context-trigger')
    expect(source).toContain('composer-context-menu')
    expect(source).toContain("import { LoomiFloatingMenu, LoomiMenuItem, LoomiMenuSeparator } from './LoomiMenu'")
    expect(source).toContain('data-loomi-menu-trigger="composer-context"')
    expect(source).toContain('ignoreSelector="[data-loomi-menu-trigger=\'composer-context\']"')
    expect(source).toContain('<LoomiMenuSeparator />')
    expect(source).toContain('addFilesAndPhotosLabel')
    expect(source).toContain('onOpenSkills?.()')
    expect(source).toContain('onOpenConnectors?.()')
    expect(source).toContain('onOpenPlugins?.()')
    expect(source).not.toContain('composer-folder-status')
  })

  test('clamps context menu placement inside the viewport', () => {
    const leftEdge = getComposerContextMenuPosition({ left: 4, top: 700, right: 38, bottom: 734 }, 390, 844)
    const rightEdge = getComposerContextMenuPosition({ left: 340, top: 700, right: 374, bottom: 734 }, 390, 844)
    const topEdge = getComposerContextMenuPosition({ left: 160, top: 40, right: 194, bottom: 74 }, 390, 844)

    expect(leftEdge).toEqual({ bottom: 152, left: 12, width: 228 })
    expect(rightEdge).toEqual({ bottom: 152, left: 150, width: 228 })
    expect(topEdge).toEqual({ top: 82, left: 150, width: 228 })
  })
})
