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
    expect(source).toContain('thinkingHintForRun')
    expect(source).toContain("from '../runtime/markdownNormalize'")
  })

  test('renders a short per-run thinking hint while assistant content is empty', () => {
    const html = renderToStaticMarkup(createElement(ChatCanvas, {
      sidebarCollapsed: false,
      thread: { id: 'thread-a', title: 'Thread A', project: 'Loomi', mode: 'chat', updatedAt: 'Now', lifecycleStatus: 'active', runStatus: 'running' },
      messages: [{ id: 'msg-user', threadId: 'thread-a', role: 'user', content: '你好', createdAt: 'Now' }],
      run: {
        id: 'run-a',
        threadId: 'thread-a',
        status: 'running',
        model: 'Local simulated',
        context: 'local_simulated',
        events: [],
        assistantDraft: { content: '', status: 'streaming' },
      },
      loading: false,
      error: null,
      dataSourceMode: 'mock',
      streamState: 'open',
      onSendMessage: () => {},
      onStopRun: () => {},
      locale: 'zh',
    }))

    expect(html).toContain('message-draft-status')
    expect(html).toMatch(/组织回复|梳理线索|核对上下文|提炼重点|推敲答案|收束思路|准备回答|再看一眼/)
    expect(html).not.toContain('模型正在生成回复')
  })

  test('uses text shimmer rather than a pulse dot for pending thinking copy', () => {
    const source = readFileSync(resolve(import.meta.dir, 'ChatCanvas.tsx'), 'utf8')
    const css = readFileSync(resolve(import.meta.dir, '../styles/20-chat.css'), 'utf8')

    expect(source).toContain('thinking-shimmer')
    expect(css).toContain('.thinking-shimmer')
    expect(css).not.toContain('draft-pulse')
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

  test('turns assistant md code payloads into a compact document card', () => {
    const html = renderToStaticMarkup(createElement(ChatCanvas, {
      sidebarCollapsed: false,
      thread: { id: 'thread-a', title: 'Thread A', project: 'Loomi', mode: 'chat', updatedAt: 'Now', lifecycleStatus: 'active', runStatus: 'completed' },
      messages: [{
        id: 'msg-a',
        threadId: 'thread-a',
        role: 'assistant',
        content: '把下面内容保存为 `三句话.md`：\n\n```md\n# 三句话的 Markdown\n\n今天我开始写一个简单的 Markdown 文档。\n```',
        createdAt: 'Now',
      }],
      run: null,
      loading: false,
      error: null,
      dataSourceMode: 'mock',
      streamState: 'closed',
      onSendMessage: () => {},
      onStopRun: () => {},
      locale: 'zh',
    }))

    expect(html).toContain('message-artifact-card')
    expect(html).toContain('三句话的 Markdown')
    expect(html).toContain('Markdown 文档')
    expect(html).not.toContain('```md')
    expect(html).not.toContain('今天我开始写一个简单的 Markdown 文档。')
  })

  test('flags a completed real API run that has no final assistant content', () => {
    const html = renderToStaticMarkup(createElement(ChatCanvas, {
      sidebarCollapsed: false,
      thread: { id: 'thread-a', title: 'Thread A', project: 'Loomi', mode: 'chat', updatedAt: 'Now', lifecycleStatus: 'active', runStatus: 'completed' },
      messages: [{ id: 'msg-user', threadId: 'thread-a', role: 'user', content: 'Use a tool', createdAt: 'Now' }],
      run: {
        id: 'run-a',
        threadId: 'thread-a',
        status: 'completed',
        model: 'local_codex',
        context: 'model_gateway',
        events: [{ id: 'evt-final', runId: 'run-a', threadId: 'thread-a', type: 'run.completed', label: 'Run', detail: 'Run completed', time: 'Now', status: 'completed' }],
      },
      loading: false,
      error: null,
      dataSourceMode: 'real_api',
      streamState: 'closed',
      onSendMessage: () => {},
      onStopRun: () => {},
      locale: 'en',
    }))

    expect(html).toContain('Final assistant message missing')
    expect(html).not.toContain('Reply generated')
    expect(html).not.toContain('未生成成功回复')
  })

  test('uses the Animal Island typewriter for the latest completed assistant answer', () => {
    const source = readFileSync(resolve(import.meta.dir, 'ChatCanvas.tsx'), 'utf8')
    expect(source).toContain("Typewriter } from 'animal-island-ui'")
    expect(source).toContain('message-markdown-typewriter')
    expect(source).toContain('shouldTypewriteHistoryMessage')
    expect(source).toContain("run.status !== 'completed'")
    expect(source).toContain('loomi.completedTypewriterMessages')
    expect(source).toContain('loomi.streamedAssistantRuns')
    expect(source).toContain('markCompletedTypewriter(typewriterTrigger)')
    expect(source).toContain('markStreamedAssistantRun(run.id)')
  })

  test('uses the Animal Island wave divider between conversation turns', () => {
    const source = readFileSync(resolve(import.meta.dir, 'ChatCanvas.tsx'), 'utf8')
    const css = readFileSync(resolve(import.meta.dir, '../styles/20-chat.css'), 'utf8')
    expect(source).toContain("import { Divider, Typewriter } from 'animal-island-ui'")
    expect(source).toContain('<Divider type="wave-yellow" />')
    expect(source).toContain("index > 0 && message.role === 'user'")
    expect(css).toContain('.conversation-divider')
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

  test('renders message-level retry regenerate and copy actions from selected run context', () => {
    const failedHtml = renderToStaticMarkup(createElement(ChatCanvas, {
      sidebarCollapsed: false,
      thread: { id: 'thread-a', title: 'Thread A', project: 'Loomi', mode: 'chat', updatedAt: 'Now', lifecycleStatus: 'active', runStatus: 'failed' },
      messages: [],
      run: { id: 'run-a', threadId: 'thread-a', status: 'failed', model: 'Local simulated', context: 'local_simulated', events: [], assistantDraft: { content: 'Failed response', status: 'failed' } },
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
    expect(failedHtml).not.toContain('composer-action" type="button">Retry')
    expect(completedHtml).toContain('Copy')
    expect(completedHtml).toContain('Regenerate')
    expect(completedHtml).not.toContain('composer-action" type="button">Regenerate')
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

  test('renders desktop readiness reason and next-step actions instead of raw fetch copy', () => {
    const html = renderToStaticMarkup(createElement(ChatCanvas, {
      sidebarCollapsed: false,
      thread: { id: 'thread-a', title: 'Thread A', project: 'Loomi', mode: 'chat', updatedAt: 'Now', lifecycleStatus: 'active', runStatus: 'idle' },
      messages: [],
      run: null,
      loading: false,
      error: 'Failed to fetch',
      dataSourceMode: 'real_api',
      streamState: 'closed',
      desktopReadiness: {
        primary: {
          code: 'local_codex_detected_disabled',
          title: 'Local Codex detected but disabled',
          detail: 'Enable it for this API session before sending.',
          action: 'enable_local_codex',
          providerId: 'local_codex',
        },
        issues: [],
      },
      onRetryReadiness: () => {},
      onOpenProviderSettings: () => {},
      onDetectLocalProviders: () => {},
      onEnableLocalProvider: () => {},
      onChooseWorkspaceFolder: () => {},
      onSendMessage: () => {},
      onStopRun: () => {},
      locale: 'en',
    }))

    expect(html).toContain('Local Codex detected but disabled')
    expect(html).toContain('Enable Local Codex')
    expect(html).toContain('Detect Local Provider')
    expect(html).toContain('Open Settings')
    expect(html).toContain('Retry')
    expect(html).not.toContain('Failed to fetch')
  })

  test('renders escaped and indented assistant headings without visible hash marks', () => {
    const html = renderToStaticMarkup(createElement(ChatCanvas, {
      sidebarCollapsed: false,
      thread: { id: 'thread-a', title: 'Thread A', project: 'Loomi', mode: 'chat', updatedAt: 'Now', lifecycleStatus: 'active', runStatus: 'completed' },
      messages: [{ id: 'msg-a', threadId: 'thread-a', role: 'assistant', content: '\u3000\\# 六、最可能占空间、建议重点处理\n\n优先看这些：', createdAt: 'Now' }],
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

    expect(html).toContain('<h1>六、最可能占空间、建议重点处理</h1>')
    expect(html).not.toContain('# 六、')
  })

  test('strips doubled heading markers from rendered assistant headings', () => {
    const html = renderToStaticMarkup(createElement(ChatCanvas, {
      sidebarCollapsed: false,
      thread: { id: 'thread-a', title: 'Thread A', project: 'Loomi', mode: 'chat', updatedAt: 'Now', lifecycleStatus: 'active', runStatus: 'completed' },
      messages: [{ id: 'msg-a', threadId: 'thread-a', role: 'assistant', content: '# # 六、最可能占空间、建议重点处理', createdAt: 'Now' }],
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

    expect(html).toContain('<h1>六、最可能占空间、建议重点处理</h1>')
    expect(html).not.toContain('<h1># 六、')
  })

  test('renders streaming draft content as markdown before the final answer completes', () => {
    const html = renderToStaticMarkup(createElement(ChatCanvas, {
      sidebarCollapsed: false,
      thread: { id: 'thread-a', title: 'Thread A', project: 'Loomi', mode: 'chat', updatedAt: 'Now', lifecycleStatus: 'active', runStatus: 'running' },
      messages: [{ id: 'msg-user', threadId: 'thread-a', role: 'user', content: 'Harness Agent', createdAt: 'Now' }],
      run: {
        id: 'run-a',
        threadId: 'thread-a',
        status: 'running',
        model: 'Local simulated',
        context: 'local_simulated',
        events: [],
        assistantDraft: { content: '## 正在整理\n\n- 第一条\n- 第二条', status: 'streaming' },
      },
      loading: false,
      error: null,
      dataSourceMode: 'real_api',
      streamState: 'live',
      providerCapabilities: [{ id: 'custom', family: 'openai_compatible' as const, model: 'gpt-5.5', status: 'available' as const }],
      onSendMessage: () => {},
      onStopRun: () => {},
      locale: 'zh',
    }))

    expect(html).toContain('正在整理')
    expect(html).toContain('<li>第一条</li>')
    expect(html).toContain('<li>第二条</li>')
    expect(html).not.toMatch(/组织回复|梳理线索|核对上下文|提炼重点|推敲答案|收束思路|准备回答|再看一眼/)
  })

  test('renders active run events in chronological assistant tool assistant order', () => {
    const html = renderToStaticMarkup(createElement(ChatCanvas, {
      sidebarCollapsed: false,
      thread: { id: 'thread-a', title: 'Thread A', project: 'Loomi', mode: 'chat', updatedAt: 'Now', lifecycleStatus: 'active', runStatus: 'running' },
      messages: [{ id: 'msg-user', threadId: 'thread-a', role: 'user', content: '查一下', createdAt: 'Now' }],
      run: {
        id: 'run-a',
        threadId: 'thread-a',
        status: 'running',
        model: 'Model gateway',
        context: 'model_gateway',
        events: [
          { id: 'evt-delta-1', runId: 'run-a', threadId: 'thread-a', sequence: 1, type: 'message.model_output_delta', label: 'message', detail: 'Model output delta', time: 'Now', status: 'running', assistantDelta: '我先查一下实时资料。', metadata: { model_phase: 'initial' } },
          { id: 'evt-tool-start', runId: 'run-a', threadId: 'thread-a', sequence: 2, type: 'tool.call.executing', label: 'tool', detail: 'Search web running', time: 'Now', status: 'running', group: 'tool-call', metadata: { tool_call_id: 'tc-search', tool_name: 'web.search', arguments_summary: { query: 'AI news' }, execution_status: 'executing' } },
          { id: 'evt-tool-done', runId: 'run-a', threadId: 'thread-a', sequence: 3, type: 'tool.call.succeeded', label: 'tool', detail: 'Search web completed', time: 'Now', status: 'running', group: 'tool-call', metadata: { tool_call_id: 'tc-search', tool_name: 'web.search', result_summary: { count: 3 }, execution_status: 'succeeded' } },
          { id: 'evt-delta-2', runId: 'run-a', threadId: 'thread-a', sequence: 4, type: 'message.model_output_delta', label: 'message', detail: 'Model output delta', time: 'Later', status: 'running', assistantDelta: '查完了，重点是模型更新。', metadata: { model_phase: 'continuation' } },
        ],
        assistantDraft: { content: '查完了，重点是模型更新。', status: 'streaming' },
      },
      loading: false,
      error: null,
      dataSourceMode: 'real_api',
      streamState: 'live',
      providerCapabilities: [{ id: 'custom', family: 'openai_compatible' as const, model: 'gpt-5.5', status: 'available' as const }],
      onSendMessage: () => {},
      onStopRun: () => {},
      locale: 'zh',
    }))

    const firstTextIndex = html.indexOf('我先查一下实时资料。')
    const toolIndex = html.indexOf('搜索网页')
    const continuationIndex = html.indexOf('查完了，重点是模型更新。')

    expect(firstTextIndex).toBeGreaterThan(-1)
    expect(toolIndex).toBeGreaterThan(firstTextIndex)
    expect(continuationIndex).toBeGreaterThan(toolIndex)
    expect(html.match(/搜索网页/g)).toHaveLength(1)
  })

  test('does not turn multiline fenced content into inline code chips', () => {
    const html = renderToStaticMarkup(createElement(ChatCanvas, {
      sidebarCollapsed: false,
      thread: { id: 'thread-a', title: 'Thread A', project: 'Loomi', mode: 'chat', updatedAt: 'Now', lifecycleStatus: 'active', runStatus: 'completed' },
      messages: [{ id: 'msg-a', threadId: 'thread-a', role: 'assistant', content: 'Before `inline`\n```text\nGrafanaAgent\n```\nAfter `ok`', createdAt: 'Now' }],
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

    expect(html).toContain('<p>Before <code>inline</code></p>')
    expect(html).toContain('<pre><code class="language-text">GrafanaAgent</code></pre>')
    expect(html).toContain('<p>After <code>ok</code></p>')
    expect(html).not.toContain('<span>text</span>')
    expect(html).not.toContain('<code>text')
  })

  test('renders an unfinished streaming fenced code block before the closing fence arrives', () => {
    const html = renderToStaticMarkup(createElement(ChatCanvas, {
      sidebarCollapsed: false,
      thread: { id: 'thread-a', title: 'Thread A', project: 'Loomi', mode: 'chat', updatedAt: 'Now', lifecycleStatus: 'active', runStatus: 'running' },
      messages: [{ id: 'msg-user', threadId: 'thread-a', role: 'user', content: '再来一个 SQL', createdAt: 'Now' }],
      run: {
        id: 'run-a',
        threadId: 'thread-a',
        status: 'running',
        model: 'Local simulated',
        context: 'local_simulated',
        events: [],
        assistantDraft: { content: '这里是 SQL：\n\n```sqlCREATE VIEW sales_summary AS\nSELECT * FROM orders;', status: 'streaming' },
      },
      loading: false,
      error: null,
      dataSourceMode: 'real_api',
      streamState: 'live',
      providerCapabilities: [{ id: 'custom', family: 'openai_compatible' as const, model: 'gpt-5.5', status: 'available' as const }],
      onSendMessage: () => {},
      onStopRun: () => {},
      locale: 'zh',
    }))

    expect(html).toContain('<div class="message-code-block">')
    expect(html).toContain('<span class="message-code-block-lang">sql</span>')
    expect(html).toContain('<code class="language-sql">CREATE VIEW sales_summary AS\nSELECT * FROM orders;</code>')
    expect(html).not.toContain('<code>sql')
  })

  test('themes assistant code blocks for light and dark chat surfaces', () => {
    const css = [
      readFileSync(resolve(import.meta.dir, '../styles/20-chat.css'), 'utf8'),
      readFileSync(resolve(import.meta.dir, '../styles/80-island-components.css'), 'utf8'),
      readFileSync(resolve(import.meta.dir, '../styles/92-unified-workspace.css'), 'utf8'),
    ].join('\n')

    expect(css).toContain('.message-code-block')
    expect(css).toContain('border: 1px solid color-mix(in srgb, var(--border-subtle) 86%, var(--text-primary) 14%)')
    expect(css).toContain('border-radius: 12px')
    expect(css).toContain('padding: 48px 48px 24px 34px')
    expect(css).toContain('background: color-mix(in srgb, var(--text-primary) 7%, transparent)')
    expect(css).toContain('.message-code-block-lang')
    expect(css).toContain(".app-shell[data-theme='dark'] .message-code-block")
    expect(css).toContain(".app-shell[data-theme='dark'] .message-markdown pre code {\n  border-color: transparent;\n  background: transparent;")
    expect(css).toContain(".app-shell[data-theme='dark'] .message-code-block-head button")
    expect(css).not.toContain('.message-markdown pre,\n.message-table-wrap,\n.tool-grid')
  })

  test('repairs collapsed markdown headings without touching fenced code', () => {
    const html = renderToStaticMarkup(createElement(ChatCanvas, {
      sidebarCollapsed: false,
      thread: { id: 'thread-a', title: 'Thread A', project: 'Loomi', mode: 'chat', updatedAt: 'Now', lifecycleStatus: 'active', runStatus: 'completed' },
      messages: [{
        id: 'msg-collapsed',
        threadId: 'thread-a',
        role: 'assistant',
        content: '先说结论。---##1.Hessian是什么？正文---###情况一：碗形\n```text\n---##not-heading\n```',
        createdAt: 'Now',
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

    expect(html).toContain('<h2>1. Hessian是什么？正文</h2>')
    expect(html).toContain('<h3>情况一：碗形</h3>')
    expect(html).toContain('---##not-heading')
  })

  test('repairs dense Chinese report markdown without turning the whole report into one heading', () => {
    const html = renderToStaticMarkup(createElement(ChatCanvas, {
      sidebarCollapsed: false,
      thread: { id: 'thread-a', title: 'Thread A', project: 'Loomi', mode: 'chat', updatedAt: 'Now', lifecycleStatus: 'active', runStatus: 'completed' },
      messages: [{
        id: 'msg-dense-report',
        threadId: 'thread-a',
        role: 'assistant',
        content: '2026 年 5 月值得关注的 AIAgent 开源项目##一、 2026 年 AIAgent 开源生态的几个核心变化到 2026 年，AIAgent 已经不只是 AutoGPT 式自主循环了，主流方向明显变成： 1.-状态管理、回滚、checkpoint、人类审批、错误恢复。 -代表： LangGraph、Microsoft AutoGen、OpenAI Agents SDK。 2.-Agent 不再只是自己写 tools，而是通过标准协议连接工具、数据源、服务。 -MCP 已经成为很多 Agent 工具接入的事实标准之一。##二、 2026 年重点开源项目',
        createdAt: 'Now',
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

    expect(html).toContain('<p>2026 年 5 月值得关注的 AIAgent 开源项目</p>')
    expect(html).toContain('<h2>一、 2026 年 AIAgent 开源生态的几个核心变化</h2>')
    expect(html).toContain('<li>状态管理、回滚、checkpoint、人类审批、错误恢复。</li>')
    expect(html).toContain('<li>Agent 不再只是自己写 tools，而是通过标准协议连接工具、数据源、服务。</li>')
    expect(html).toContain('<h2>二、 2026 年重点开源项目</h2>')
    expect(html).not.toContain('<h2>2026 年 5 月值得关注的 AIAgent 开源项目##一')
    expect(html).not.toContain('代表： LangGraph、Microsoft AutoGen、OpenAI Agents SDK。 2.-Agent')
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
    expect(html).toContain('<td><a href="https://example.com"')
    expect(html).not.toContain('| 序号 | 新闻 | 链接 |')
  })

  test('renders real smoke final markdown tables and code blocks', () => {
    const html = renderToStaticMarkup(createElement(ChatCanvas, {
      sidebarCollapsed: false,
      thread: { id: 'thread-a', title: 'Thread A', project: 'Loomi', mode: 'chat', updatedAt: 'Now', lifecycleStatus: 'active', runStatus: 'completed' },
      messages: [{
        id: 'msg-final',
        threadId: 'thread-a',
        role: 'assistant',
        content: '## Final\n\n| File | Kind |\n| --- | --- |\n| `cmd/loomi` | CLI |\n\n```bash\nloomi smoke agent\n```',
        createdAt: 'Now',
        runId: 'run-a',
      }],
      run: {
        id: 'run-a',
        threadId: 'thread-a',
        status: 'completed',
        model: 'local_codex',
        context: 'model_gateway',
        events: [{ id: 'evt-final', runId: 'run-a', threadId: 'thread-a', type: 'run.completed', label: 'Run', detail: 'Run completed', time: 'Now', status: 'completed' }],
      },
      loading: false,
      error: null,
      dataSourceMode: 'real_api',
      streamState: 'closed',
      onSendMessage: () => {},
      onStopRun: () => {},
      locale: 'en',
    }))

    expect(html).toContain('<h2>Final</h2>')
    expect(html).toContain('<table>')
    expect(html).toContain('<code>cmd/loomi</code>')
    expect(html).toContain('<code class="language-bash">loomi smoke agent</code>')
  })

  test('groups consecutive run tool events as one turn activity', () => {
    const html = renderToStaticMarkup(createElement(ChatCanvas, {
      sidebarCollapsed: false,
      thread: { id: 'thread-a', title: 'Thread A', project: 'Loomi', mode: 'chat', updatedAt: 'Now', lifecycleStatus: 'active', runStatus: 'running' },
      messages: [{ id: 'msg-user', threadId: 'thread-a', role: 'user', content: 'Check files', createdAt: 'Now' }],
      run: {
        id: 'run-a',
        threadId: 'thread-a',
        status: 'running',
        model: 'local_codex',
        context: 'model_gateway',
        events: [
          { id: 'evt-tool-1', runId: 'run-a', threadId: 'thread-a', type: 'tool.call.succeeded', label: 'Tool', detail: 'done', time: 'Now', status: 'completed', sequence: 1, metadata: { tool_call_id: 'tc-1', tool_name: 'workspace.glob', arguments_summary: { pattern: '*.tsx' }, result_summary: { match_count: 2 } } },
          { id: 'evt-tool-2', runId: 'run-a', threadId: 'thread-a', type: 'tool.call.succeeded', label: 'Tool', detail: 'done', time: 'Now', status: 'completed', sequence: 2, metadata: { tool_call_id: 'tc-2', tool_name: 'workspace.grep', arguments_summary: { pattern: 'ToolCallCard' }, result_summary: { match_count: 4 } } },
        ],
      },
      loading: false,
      error: null,
      dataSourceMode: 'real_api',
      streamState: 'open',
      providerCapabilities: [{ id: 'local_codex', family: 'openai_compatible', model: 'gpt-5.5', status: 'available', executionState: 'supported' }],
      onSendMessage: () => {},
      onStopRun: () => {},
      locale: 'en',
    }))

    expect(html).toContain('tool-transcript-group')
    expect(html.match(/tool-call-draft/g)).toHaveLength(1)
    expect(html).toContain('Find project files')
    expect(html).toContain('Search project text')
  })

  test('keeps a waiting-for-model state after terminal tool events before continuation text arrives', () => {
    const html = renderToStaticMarkup(createElement(ChatCanvas, {
      sidebarCollapsed: false,
      thread: { id: 'thread-a', title: 'Thread A', project: 'Loomi', mode: 'chat', updatedAt: 'Now', lifecycleStatus: 'active', runStatus: 'running' },
      messages: [{ id: 'msg-user', threadId: 'thread-a', role: 'user', content: 'Search first', createdAt: 'Now' }],
      run: {
        id: 'run-a',
        threadId: 'thread-a',
        status: 'running',
        model: 'local_codex',
        context: 'model_gateway',
        events: [
          { id: 'evt-tool-1', runId: 'run-a', threadId: 'thread-a', type: 'tool.call.succeeded', label: 'Tool', detail: 'done', time: 'Now', status: 'completed', sequence: 1, metadata: { tool_call_id: 'tc-1', tool_name: 'web.search', result_summary: { count: 2 }, execution_status: 'succeeded' } },
        ],
      },
      loading: false,
      error: null,
      dataSourceMode: 'real_api',
      streamState: 'open',
      providerCapabilities: [{ id: 'local_codex', family: 'openai_compatible', model: 'gpt-5.5', status: 'available', executionState: 'supported' }],
      onSendMessage: () => {},
      onStopRun: () => {},
      locale: 'zh',
    }))

    expect(html).toContain('搜索网页')
    expect(html).toContain('生成中')
    expect(html).toMatch(/组织回复|梳理线索|核对上下文|提炼重点|推敲答案|收束思路|准备回答|再看一眼/)
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
    expect(html).toContain('打开设置')
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
