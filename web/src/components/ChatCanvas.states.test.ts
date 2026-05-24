import { describe, expect, test } from 'bun:test'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { createElement } from 'react'
import { renderToStaticMarkup } from 'react-dom/server'
import { ChatCanvas } from './ChatCanvas'

describe('ChatCanvas state copy', () => {
  test('gets sparse Chinese labels from the i18n dictionary', () => {
    const source = readFileSync(resolve(import.meta.dir, '../i18n.ts'), 'utf8')

    expect(source).toContain('未选择会话')
    expect(source).toContain('新对话')
    expect(source).toContain('加载中')
    expect(source).toContain('加载失败')
    expect(source).toContain('等待执行')
    expect(source).toContain('执行中')
    expect(source).toContain('已完成')
    expect(source).toContain('执行失败')
    expect(source).toContain('已停止')
    expect(source).toContain('恢复中')
    expect(source).toContain('后端能力未接入')
    expect(source).toContain('模型网关')
    expect(source).toContain('工具调用未执行')
    expect(source).toContain('已停止生成')
  })

  test('renders assistant draft bubble states from run.assistantDraft without duplicating final content', () => {
    const source = readFileSync(resolve(import.meta.dir, 'ChatCanvas.tsx'), 'utf8')

    expect(source).toContain('assistantDraft')
    expect(source).toContain('draft.content')
    expect(source).toContain('pending')
    expect(source).toContain('streaming')
    expect(source).toContain('failed')
    expect(source).toContain('stopped')
    expect(source).toContain('recovering')
    expect(source).not.toContain("draft.status === 'completed') return null")
    expect(source).not.toContain('message.runId !== run.id')
    expect(source).toContain("status === 'failed'")
    expect(source).toContain('draft.content || draftFallback')
  })

  test('routes ChatCanvas rendering through deriveChatCanvasState', () => {
    const source = readFileSync(resolve(import.meta.dir, 'ChatCanvas.tsx'), 'utf8')

    expect(source).toContain("from '../runtime/chatCanvasState'")
    expect(source).toContain('deriveChatCanvasState')
  })

  test('renders completed assistant content exactly once when message and draft both exist', () => {
    const html = renderToStaticMarkup(createElement(ChatCanvas, {
      sidebarCollapsed: false,
      thread: { id: 'thread-a', title: 'Thread A', project: 'Loomi', mode: 'chat', updatedAt: 'Now', lifecycleStatus: 'active', runStatus: 'completed' },
      messages: [{ id: 'msg-a', threadId: 'thread-a', role: 'assistant', content: 'Final answer', createdAt: 'Now', runId: 'run-a' }],
      run: {
        id: 'run-a',
        threadId: 'thread-a',
        status: 'completed',
        model: 'Local simulated',
        context: 'local_simulated',
        events: [{ id: 'evt-final', runId: 'run-a', threadId: 'thread-a', type: 'model.final', label: 'Model', detail: 'final', time: 'Now', status: 'completed' }],
        assistantDraft: { content: 'Final answer', status: 'completed', messageId: 'msg-a' },
      },
      loading: false,
      error: null,
      dataSourceMode: 'mock',
      streamState: 'closed',
      onSendMessage: () => {},
      onStopRun: () => {},
      locale: 'en',
    }))

    expect(html.match(/Final answer/g)).toHaveLength(1)
  })

  test('deduplicates restored completed draft when persisted assistant message lacks run linkage', () => {
    const html = renderToStaticMarkup(createElement(ChatCanvas, {
      sidebarCollapsed: false,
      thread: { id: 'thread-a', title: 'Thread A', project: 'Loomi', mode: 'chat', updatedAt: 'Now', lifecycleStatus: 'active', runStatus: 'completed' },
      messages: [{ id: 'msg-a', threadId: 'thread-a', role: 'assistant', content: 'Real final', createdAt: 'Now' }],
      run: {
        id: 'run-a',
        threadId: 'thread-a',
        status: 'completed',
        model: 'Local simulated',
        context: 'local_simulated',
        events: [{ id: 'evt-final', runId: 'run-a', threadId: 'thread-a', type: 'model.final', label: 'Model', detail: 'final', time: 'Now', status: 'completed', content: 'Real final' }],
        assistantDraft: { content: 'Real final', status: 'completed' },
      },
      loading: false,
      error: null,
      dataSourceMode: 'real_api',
      streamState: 'closed',
      onSendMessage: () => {},
      onStopRun: () => {},
      locale: 'en',
    }))

    expect(html.match(/Real final/g)).toHaveLength(1)
  })

  test('renders composer retry and regenerate actions from selected run context', () => {
    const failedHtml = renderToStaticMarkup(createElement(ChatCanvas, {
      sidebarCollapsed: false,
      thread: { id: 'thread-a', title: 'Thread A', project: 'Loomi', mode: 'chat', updatedAt: 'Now', lifecycleStatus: 'active', runStatus: 'failed' },
      messages: [],
      run: { id: 'run-a', threadId: 'thread-a', status: 'failed', model: 'Local simulated', context: 'local_simulated', events: [] },
      loading: false,
      error: null,
      dataSourceMode: 'mock',
      streamState: 'closed',
      onSendMessage: () => {},
      onStopRun: () => {},
      onRetryRun: () => {},
      onRegenerateRun: () => {},
      locale: 'en',
    }))
    const completedHtml = renderToStaticMarkup(createElement(ChatCanvas, {
      sidebarCollapsed: false,
      thread: { id: 'thread-a', title: 'Thread A', project: 'Loomi', mode: 'chat', updatedAt: 'Now', lifecycleStatus: 'active', runStatus: 'completed' },
      messages: [{ id: 'msg-a', threadId: 'thread-a', role: 'assistant', content: 'Done', createdAt: 'Now' }],
      run: { id: 'run-a', threadId: 'thread-a', status: 'completed', model: 'Local simulated', context: 'local_simulated', events: [] },
      loading: false,
      error: null,
      dataSourceMode: 'mock',
      streamState: 'closed',
      onSendMessage: () => {},
      onStopRun: () => {},
      onRetryRun: () => {},
      onRegenerateRun: () => {},
      locale: 'en',
    }))

    expect(failedHtml).toContain('Retry')
    expect(completedHtml).toContain('Regenerate')
  })

  test('renders capability status chip and detail without implying fake model thinking', () => {
    const mockHtml = renderToStaticMarkup(createElement(ChatCanvas, {
      sidebarCollapsed: false,
      thread: { id: 'thread-a', title: 'Thread A', project: 'Loomi', mode: 'chat', updatedAt: 'Now', lifecycleStatus: 'active', runStatus: 'completed' },
      messages: [],
      run: null,
      loading: false,
      error: null,
      dataSourceMode: 'mock',
      streamState: 'closed',
      onSendMessage: () => {},
      onStopRun: () => {},
      locale: 'en',
    }))
    const disconnectedHtml = renderToStaticMarkup(createElement(ChatCanvas, {
      sidebarCollapsed: false,
      thread: { id: 'thread-a', title: 'Thread A', project: 'Loomi', mode: 'chat', updatedAt: 'Now', lifecycleStatus: 'active', runStatus: 'running' },
      messages: [],
      run: { id: 'run-a', threadId: 'thread-a', status: 'running', model: 'Local simulated', context: 'local_simulated', events: [] },
      loading: false,
      error: null,
      dataSourceMode: 'real_api',
      streamState: 'recoverable_error',
      onSendMessage: () => {},
      onStopRun: () => {},
      locale: 'en',
    }))

    expect(mockHtml).toContain('Mock')
    expect(mockHtml).toContain('not real model output')
    expect(disconnectedHtml).toContain('Stream disconnected')
    expect(disconnectedHtml).toContain('event stream disconnected')
    expect(disconnectedHtml).not.toContain('model is thinking')
  })
})