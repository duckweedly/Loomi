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
    expect(source).toContain('模型 Provider 未配置或不可用')
    expect(source).toContain('Local Codex 已启用，但暂不支持执行')
    expect(source).toContain('Local Codex 登录态不可用，请重新检测或配置 OpenAI-compatible provider')
    expect(source).toContain('打开设置')
    expect(source).toContain('Retrying')
    expect(source).toContain('重试中')
    expect(source).toContain('已耗尽')
  })

  test('renders assistant draft bubble states from run.assistantDraft without duplicating final content', () => {
    const source = readFileSync(resolve(import.meta.dir, 'ChatCanvas.tsx'), 'utf8')

    expect(source).toContain('assistantDraft')
    expect(source).toContain('draft.content')
    expect(source).toContain('streaming')
    expect(source).toContain('failed')
    expect(source).toContain('stopped')
    expect(source).toContain('recovering')
    expect(source).not.toContain("draft.status === 'completed') return null")
    expect(source).toContain('visibleRunForTranscript')
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

  test('does not render the old runtime capability header', () => {
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

    expect(mockHtml).not.toContain('Mock demo mode')
    expect(mockHtml).not.toContain('not real model output')
    expect(disconnectedHtml).not.toContain('Stream disconnected')
    expect(disconnectedHtml).not.toContain('event stream disconnected')
    expect(disconnectedHtml).not.toContain('model is thinking')
  })

  test('renders basic markdown in assistant messages', () => {
    const html = renderToStaticMarkup(createElement(ChatCanvas, {
      sidebarCollapsed: false,
      thread: { id: 'thread-a', title: 'Thread A', project: 'Loomi', mode: 'chat', updatedAt: 'Now', lifecycleStatus: 'active', runStatus: 'completed' },
      messages: [{ id: 'msg-a', threadId: 'thread-a', role: 'assistant', content: '# Title\n- item\n\n**bold** and `code`', createdAt: 'Now' }],
      run: null,
      loading: false,
      error: null,
      dataSourceMode: 'real_api',
      streamState: 'closed',
      providerCapabilities: [{ id: 'custom', family: 'openai_compatible' as const, model: 'gpt-5.5', status: 'available' as const }],
      onSendMessage: () => {},
      onStopRun: () => {},
      locale: 'en',
    }))

    expect(html).toContain('<h1>Title</h1>')
    expect(html).toContain('<li>item</li>')
    expect(html).toContain('<strong>bold</strong>')
    expect(html).toContain('<code>code</code>')
  })

  test('renders markdown tables in assistant messages', () => {
    const html = renderToStaticMarkup(createElement(ChatCanvas, {
      sidebarCollapsed: false,
      thread: { id: 'thread-a', title: 'Thread A', project: 'Loomi', mode: 'chat', updatedAt: 'Now', lifecycleStatus: 'active', runStatus: 'completed' },
      messages: [{
        id: 'msg-table',
        threadId: 'thread-a',
        role: 'assistant',
        content: '| 序号 | 新闻 | 链接 |\n|---|---|---|\n| 1 | Reuters AI News | https://example.com |\n| 2 | LLM News Today | https://example.com/llm |',
        createdAt: '2026-05-26T08:30:00Z',
      }],
      run: null,
      loading: false,
      error: null,
      dataSourceMode: 'real_api',
      streamState: 'closed',
      providerCapabilities: [{ id: 'custom', family: 'openai_compatible' as const, model: 'gpt-5.5', status: 'available' as const }],
      onSendMessage: () => {},
      onStopRun: () => {},
      locale: 'zh',
    }))

    expect(html).toContain('<table>')
    expect(html).toContain('<th>序号</th>')
    expect(html).toContain('<td>Reuters AI News</td>')
    expect(html).toContain('<td>https://example.com</td>')
    expect(html).not.toContain('| 序号 | 新闻 | 链接 |')
  })

  test('does not leak stale active tool approval into a newer completed chat turn', () => {
    const html = renderToStaticMarkup(createElement(ChatCanvas, {
      sidebarCollapsed: false,
      thread: { id: 'thread-a', title: 'Thread A', project: 'Loomi', mode: 'chat', updatedAt: 'Now', lifecycleStatus: 'active', runStatus: 'blocked_on_tool_approval' },
      messages: [
        { id: 'msg-old-user', threadId: 'thread-a', role: 'user', content: 'Sketch a small agent profile surface.', createdAt: '2026-05-25T08:00:00Z' },
        { id: 'msg-old-assistant', threadId: 'thread-a', role: 'assistant', content: 'Use a compact profile card later.', createdAt: '2026-05-25T08:00:01Z', runId: 'run-old' },
        { id: 'msg-new-user', threadId: 'thread-a', role: 'user', content: '你好呀', createdAt: '2026-05-26T09:19:00Z' },
        { id: 'msg-new-assistant', threadId: 'thread-a', role: 'assistant', content: '正在整理答案。', createdAt: '2026-05-26T09:19:01Z', runId: 'run-new' },
      ],
      run: {
        id: 'run-old',
        threadId: 'thread-a',
        status: 'blocked_on_tool_approval',
        model: 'Model gateway',
        context: 'model_gateway',
        events: [{ id: 'evt-old', runId: 'run-old', threadId: 'thread-a', type: 'tool.call.approval_required', label: 'Tool approval required', detail: 'waiting', time: '2026-05-25T08:00:00Z', status: 'blocked_on_tool_approval' }],
        toolCalls: [{
          id: 'tc-old',
          toolCallId: 'tc-old',
          name: 'agent.spawn',
          status: 'approval_required',
          approvalStatus: 'required',
          executionStatus: 'blocked',
          summary: 'Coordinate agent task',
          input: '',
          output: '',
        }],
      },
      loading: false,
      error: null,
      dataSourceMode: 'real_api',
      streamState: 'closed',
      providerCapabilities: [{ id: 'local_codex', family: 'openai_compatible', model: 'gpt-5.5', status: 'available', executionState: 'supported' }],
      onSendMessage: () => {},
      onStopRun: () => {},
      locale: 'zh',
    }))

    expect(html).toContain('你好呀')
    expect(html).toContain('正在整理答案。')
    expect(html).not.toContain('等待你确认')
    expect(html).not.toContain('协调子任务')
    expect(html).not.toContain('允许')
    expect(html).not.toContain('拒绝')
    expect(html).not.toContain('<textarea class="composer-input" disabled=""')
  })
})

describe('ChatCanvas provider unavailable warning', () => {
  const thread = { id: 'thread-a', title: 'Thread A', project: 'Loomi', mode: 'chat' as const, updatedAt: 'Now', lifecycleStatus: 'active' as const, runStatus: 'completed' as const }
  const availableProvider = { id: 'custom', family: 'openai_compatible' as const, model: 'gpt-5.5', status: 'available' as const }
  const unavailableProvider = { ...availableProvider, status: 'unavailable' as const }

  test('does not show real provider warning in mock mode', () => {
    const html = renderToStaticMarkup(createElement(ChatCanvas, {
      sidebarCollapsed: false,
      thread,
      messages: [],
      run: null,
      loading: false,
      error: null,
      dataSourceMode: 'mock',
      streamState: 'closed',
      providerCapabilities: [],
      onSendMessage: () => {},
      onStopRun: () => {},
      locale: 'en',
    }))

    expect(html).not.toContain('Model provider is not configured or unavailable')
  })

  test('shows guidance when real API has no available provider and disables Composer submit', () => {
    const html = renderToStaticMarkup(createElement(ChatCanvas, {
      sidebarCollapsed: false,
      thread,
      messages: [],
      run: null,
      loading: false,
      error: null,
      dataSourceMode: 'real_api',
      streamState: 'closed',
      providerCapabilities: [],
      onOpenProviderSettings: () => {},
      onSendMessage: () => {},
      onStopRun: () => {},
      locale: 'en',
    }))

    expect(html).toContain('Model provider is not configured or unavailable')
    expect(html).toContain('Provider Settings')
    expect(html).not.toContain('Generating')
    expect(html).toContain('<textarea')
    expect(html).toContain('<textarea class="composer-input" disabled=""')
    expect(html).not.toContain('Retry</button>')
    expect(html).not.toContain('Regenerate</button>')
  })

  test('shows guidance when real API providers are configured but unavailable', () => {
    const html = renderToStaticMarkup(createElement(ChatCanvas, {
      sidebarCollapsed: false,
      thread,
      messages: [],
      run: null,
      loading: false,
      error: null,
      dataSourceMode: 'real_api',
      streamState: 'closed',
      providerCapabilities: [unavailableProvider],
      onOpenProviderSettings: () => {},
      onSendMessage: () => {},
      onStopRun: () => {},
      locale: 'zh',
    }))

    expect(html).toContain('模型 Provider 未配置或不可用')
    expect(html).toContain('Provider Settings')
  })

  test('keeps guidance when only enabled local provider execution is unsupported', () => {
    const html = renderToStaticMarkup(createElement(ChatCanvas, {
      sidebarCollapsed: false,
      thread,
      messages: [],
      run: null,
      loading: false,
      error: null,
      dataSourceMode: 'real_api',
      streamState: 'closed',
      providerCapabilities: [{ ...unavailableProvider, id: 'local_codex', localProvider: true, sessionLocal: true, credentialReference: 'redacted', executionState: 'unsupported', message: 'enabled but execution unsupported' }],
      onOpenProviderSettings: () => {},
      onSendMessage: () => {},
      onStopRun: () => {},
      locale: 'en',
    }))

    expect(html).toContain('Local Codex is enabled but execution is not supported yet')
    expect(html).toContain('<textarea class="composer-input" disabled=""')
  })

  test('shows local unavailable guidance when enabled local codex login is unavailable', () => {
    const html = renderToStaticMarkup(createElement(ChatCanvas, {
      sidebarCollapsed: false,
      thread,
      messages: [],
      run: null,
      loading: false,
      error: null,
      dataSourceMode: 'real_api',
      streamState: 'closed',
      providerCapabilities: [{ ...unavailableProvider, id: 'local_codex', localProvider: true, sessionLocal: true, credentialReference: 'redacted', executionState: 'supported', message: 'login unavailable' }],
      onOpenProviderSettings: () => {},
      onSendMessage: () => {},
      onStopRun: () => {},
      locale: 'zh',
    }))

    expect(html).toContain('Local Codex 登录态不可用，请重新检测或配置 OpenAI-compatible provider')
    expect(html).toContain('<textarea class="composer-input" disabled=""')
  })

  test('hides guidance for available supported local codex provider', () => {
    const html = renderToStaticMarkup(createElement(ChatCanvas, {
      sidebarCollapsed: false,
      thread,
      messages: [],
      run: null,
      loading: false,
      error: null,
      dataSourceMode: 'real_api',
      streamState: 'closed',
      providerCapabilities: [{ ...availableProvider, id: 'local_codex', localProvider: true, sessionLocal: true, credentialReference: 'redacted', executionState: 'supported' }],
      onOpenProviderSettings: () => {},
      onSendMessage: () => {},
      onStopRun: () => {},
      locale: 'en',
    }))

    expect(html).not.toContain('Model provider is not configured or unavailable')
    expect(html).not.toContain('Provider unavailable')
    expect(html).not.toContain('<textarea class="composer-input" disabled=""')
  })

  test('hides guidance when real API has an available provider', () => {
    const html = renderToStaticMarkup(createElement(ChatCanvas, {
      sidebarCollapsed: false,
      thread,
      messages: [],
      run: null,
      loading: false,
      error: null,
      dataSourceMode: 'real_api',
      streamState: 'closed',
      providerCapabilities: [availableProvider],
      onOpenProviderSettings: () => {},
      onSendMessage: () => {},
      onStopRun: () => {},
      locale: 'en',
    }))

    expect(html).not.toContain('Model provider is not configured or unavailable')
    expect(html).not.toContain('Open Settings')
  })
})
