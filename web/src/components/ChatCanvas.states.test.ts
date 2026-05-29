import { describe, expect, mock, test } from 'bun:test'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { createElement, type ReactNode } from 'react'
import { renderToStaticMarkup } from 'react-dom/server'

mock.module('animal-island-ui', () => ({
  Button: ({ children, htmlType = 'button', ...props }: { children?: ReactNode; htmlType?: 'button' | 'submit' | 'reset'; [key: string]: unknown }) => createElement('button', { ...props, type: htmlType }, children),
  Divider: ({ type }: { type?: string }) => createElement('div', { 'data-divider': type }),
  Select: ({ value, children, onChange, ...props }: { value?: string; children?: ReactNode; onChange?: (value: string) => void; [key: string]: unknown }) => createElement('select', { ...props, value, onChange: (event: { target: { value: string } }) => onChange?.(event.target.value) }, children),
  Typewriter: ({ children }: { children?: ReactNode }) => createElement('div', { 'data-typewriter': true }, children),
}))

const { ChatCanvas } = await import('./ChatCanvas')

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

  test('keeps the new run thinking state visible in a thread with older assistant replies', () => {
    const html = renderToStaticMarkup(createElement(ChatCanvas, {
      sidebarCollapsed: false,
      thread: { id: 'thread-a', title: 'Thread A', project: 'Loomi', mode: 'chat', updatedAt: 'Now', lifecycleStatus: 'active', runStatus: 'running' },
      messages: [
        { id: 'msg-old-user', threadId: 'thread-a', role: 'user', content: '你好', createdAt: 'Earlier' },
        { id: 'msg-old-assistant', threadId: 'thread-a', role: 'assistant', content: '旧回复', createdAt: 'Earlier', runId: 'run-old' },
        { id: 'msg-new-user', threadId: 'thread-a', role: 'user', content: '再来一个', createdAt: 'Now' },
      ],
      run: {
        id: 'run-new',
        threadId: 'thread-a',
        status: 'running',
        model: 'Local simulated',
        context: 'local_simulated',
        events: [],
        assistantDraft: { content: '', status: 'streaming' },
      },
      loading: false,
      error: null,
      dataSourceMode: 'real_api',
      streamState: 'live',
      onSendMessage: () => {},
      onStopRun: () => {},
      locale: 'zh',
    }))

    expect(html).toContain('message-draft-status')
    expect(html).toMatch(/组织回复|梳理线索|核对上下文|提炼重点|推敲答案|收束思路|准备回答|再看一眼/)
    expect(html).toContain('旧回复')
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

  test('pauses bottom following when the user scrolls away during active generation', () => {
    const source = readFileSync(resolve(import.meta.dir, 'ChatCanvas.tsx'), 'utf8')
    const css = readFileSync(resolve(import.meta.dir, '../styles/20-chat.css'), 'utf8')

    expect(source).toContain('messageListRef')
    expect(source).toContain('messageEndRef')
    expect(source).toContain('shouldFollowBottomRef')
    expect(source).toContain('isNearScrollBottom')
    expect(source).toContain("addEventListener('scroll'")
    expect(source).toContain('passive: true')
    expect(source).not.toContain('shouldStickToBottom = distanceFromBottom < 180 || Boolean')
    expect(source).toContain('window.requestAnimationFrame')
    expect(source).toContain('list.scrollTop = list.scrollHeight')
    expect(source).not.toContain('scrollIntoView')
    expect(source).toContain('className="message-end-anchor"')
    expect(css).toContain('.message-end-anchor')
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

  test('typewrites a fast completed assistant message when no live stream was seen', () => {
    const source = readFileSync(resolve(import.meta.dir, 'ChatCanvas.tsx'), 'utf8')
    const typewriteStart = source.indexOf('function shouldTypewriteHistoryMessage')
    const typewriteEnd = source.indexOf('function hasAssistantMessageAfterLatestUser', typewriteStart)
    const typewriteSource = source.slice(typewriteStart, typewriteEnd)

    expect(typewriteSource).toContain("run.status !== 'completed'")
    expect(typewriteSource).toContain('hasStreamedAssistantRun(run.id)')
    expect(typewriteSource).toContain('message.runId && message.runId === run.id')
    expect(typewriteSource).not.toContain('hasPersistedTranscriptContent([message], run)')
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

  test('renders a completed real API run without final content as a neutral run notice', () => {
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

    expect(html).toContain('run-inline-notice final-missing')
    expect(html).toContain('No reply generated')
    expect(html).toContain('This turn finished without displayable content. Try again.')
    expect(html).not.toContain('runtime-final-warning')
    expect(html).not.toContain('message-row assistant final-missing')
    expect(html).not.toContain('Loomi · Completed')
    expect(html).not.toContain('Reply generated')
    expect(html).not.toContain('未生成成功回复')
  })

  test('does not show missing-final copy while waiting for workspace authorization', () => {
    const html = renderToStaticMarkup(createElement(ChatCanvas, {
      sidebarCollapsed: false,
      thread: { id: 'thread-a', title: 'Thread A', project: 'Loomi', mode: 'chat', updatedAt: 'Now', lifecycleStatus: 'active', runStatus: 'idle' },
      messages: [{ id: 'msg-user', threadId: 'thread-a', role: 'user', content: '看下我下载目录整理一下', createdAt: 'Now' }],
      run: {
        id: 'deferred-thread-a',
        threadId: 'thread-a',
        status: 'completed',
        model: 'Deferred',
        context: 'M3 thread/message only',
        events: [],
      },
      loading: false,
      error: null,
      dataSourceMode: 'real_api',
      streamState: 'closed',
      workspaceRootConfig: { configured: true, displayName: 'Loomi' },
      onChooseWorkspaceFolder: () => {},
      onSendMessage: () => {},
      onStopRun: () => {},
      locale: 'zh',
    }))

    expect(html).toContain('授权下载目录')
    expect(html).toContain('选择下载目录')
    expect(html).not.toContain('未生成回复')
    expect(html).not.toContain('这轮结束时没有拿到可显示内容')
    expect(html).not.toContain('final-missing')
  })

  test('treats restored assistant transcript text as the visible final answer', () => {
    const html = renderToStaticMarkup(createElement(ChatCanvas, {
      sidebarCollapsed: false,
      thread: { id: 'thread-a', title: 'Thread A', project: 'Loomi', mode: 'chat', updatedAt: 'Now', lifecycleStatus: 'active', runStatus: 'completed' },
      messages: [{ id: 'msg-user', threadId: 'thread-a', role: 'user', content: '重新给下', createdAt: 'Now' }],
      run: {
        id: 'run-a',
        threadId: 'thread-a',
        status: 'completed',
        model: 'local_codex',
        context: 'model_gateway',
        events: [
          { id: 'evt-delta', runId: 'run-a', threadId: 'thread-a', sequence: 1, type: 'message.model_output_delta', label: 'Message', detail: 'delta', time: 'Now', status: 'running', content: '这是恢复出来的最终回复。' },
          { id: 'evt-final', runId: 'run-a', threadId: 'thread-a', sequence: 2, type: 'run.completed', label: 'Run', detail: 'Run completed', time: 'Now', status: 'completed' },
        ],
      },
      loading: false,
      error: null,
      dataSourceMode: 'real_api',
      streamState: 'closed',
      onSendMessage: () => {},
      onStopRun: () => {},
      locale: 'zh',
    }))

    expect(html).toContain('这是恢复出来的最终回复。')
    expect(html).toContain('Loomi · 已完成')
    expect(html).not.toContain('生成中')
    expect(html).not.toContain('未生成回复')
    expect(html).not.toContain('final-missing')
  })

  test('does not leak a stale completed run warning into an empty selected thread', () => {
    const html = renderToStaticMarkup(createElement(ChatCanvas, {
      sidebarCollapsed: false,
      thread: { id: 'thread-new', title: 'New thread', project: 'Loomi', mode: 'chat', updatedAt: 'Now', lifecycleStatus: 'active', runStatus: 'idle' },
      messages: [],
      run: {
        id: 'run-old',
        threadId: 'thread-old',
        status: 'completed',
        model: 'local_codex',
        context: 'model_gateway',
        events: [{ id: 'evt-old', runId: 'run-old', threadId: 'thread-old', type: 'run.completed', label: 'Run', detail: 'Run completed', time: 'Now', status: 'completed' }],
      },
      loading: false,
      error: null,
      dataSourceMode: 'real_api',
      streamState: 'closed',
      onSendMessage: () => {},
      onStopRun: () => {},
      locale: 'zh',
    }))

    expect(html).toContain('新对话')
    expect(html).toContain('输入第一条消息')
    expect(html).not.toContain('最终回复缺失')
    expect(html).not.toContain('Run 已完成')
    expect(html).not.toContain('这轮没有生成回复')
    expect(html).not.toContain('未生成回复')
  })

  test('keeps missing-final notices out of empty selected threads', () => {
    const source = readFileSync(resolve(import.meta.dir, 'ChatCanvas.tsx'), 'utf8')

    expect(source).toContain('hasCurrentTurnUserMessage')
    expect(source).toContain('latestUserMessage(messages, thread.id)')
    expect(source).toContain('hasCurrentTurnUserMessage && visibleRun?.status')
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

  test('uses a quiet turn gap instead of decorative wave dividers', () => {
    const source = readFileSync(resolve(import.meta.dir, 'ChatCanvas.tsx'), 'utf8')
    const css = readFileSync(resolve(import.meta.dir, '../styles/20-chat.css'), 'utf8')
    expect(source).toContain("import { Typewriter } from 'animal-island-ui'")
    expect(source).toContain('<div className="conversation-divider" aria-hidden="true" />')
    expect(source).not.toContain('wave-yellow')
    expect(source).toContain("index > 0 && message.role === 'user'")
    expect(css).toContain('.conversation-divider')
    expect(css).not.toContain("[class*='divider']")
  })

  test('aligns user turns as Craft-style right-side bubbles', () => {
    const css = readFileSync(resolve(import.meta.dir, '../styles/92-unified-workspace.css'), 'utf8')

    expect(css).toContain('.message-row.user')
    expect(css).toContain('justify-items: end !important;')
    expect(css).toContain('.message-row.user .message-avatar')
    expect(css).toContain('.message-row.user .message-meta')
    expect(css).toContain('max-width: min(78%, 720px) !important;')
    expect(css).toContain('border-radius: 16px !important;')
    expect(css).toContain('mask-image: linear-gradient(to bottom, transparent 0%, black 28px')
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

  test('does not render a completed draft after a persisted assistant final in the same turn', () => {
    const html = renderToStaticMarkup(createElement(ChatCanvas, {
      sidebarCollapsed: false,
      thread: { id: 'thread-a', title: 'Thread A', project: 'Loomi', mode: 'chat', updatedAt: 'Now', lifecycleStatus: 'active', runStatus: 'completed' },
      messages: [
        { id: 'msg-user', threadId: 'thread-a', role: 'user', content: '?', createdAt: 'Now' },
        { id: 'msg-final', threadId: 'thread-a', role: 'assistant', content: '如果你想要“一个 Markdown 文件内容”，直接用下面这个通用版： `markdown #文档标题##概述这里写文档的简要说明。` 如果你要特定类型即可。', createdAt: 'Now' },
      ],
      run: {
        id: 'run-a',
        threadId: 'thread-a',
        status: 'completed',
        model: 'Local simulated',
        context: 'local_simulated',
        events: [],
        assistantDraft: { content: '如果你想要“一个Markdown文件内容”，直接用下面这个通用版： `markdown #文档标题##概述这里写文档的简要说明。` 如果你要特定类型即可。', status: 'completed' },
      },
      loading: false,
      error: null,
      dataSourceMode: 'real_api',
      streamState: 'closed',
      onSendMessage: () => {},
      onStopRun: () => {},
      locale: 'zh',
    }))

    expect(html).toContain('文档标题')
    expect(html).not.toContain('一个Markdown文件内容')
  })

  test('does not replay assistant transcript text after a persisted assistant final in the same turn', () => {
    const html = renderToStaticMarkup(createElement(ChatCanvas, {
      sidebarCollapsed: false,
      thread: { id: 'thread-a', title: 'Thread A', project: 'Loomi', mode: 'chat', updatedAt: 'Now', lifecycleStatus: 'active', runStatus: 'completed' },
      messages: [
        { id: 'msg-user', threadId: 'thread-a', role: 'user', content: 'hello', createdAt: 'Now' },
        { id: 'msg-final', threadId: 'thread-a', role: 'assistant', content: 'Hello! How can I help?', createdAt: 'Now' },
      ],
      run: {
        id: 'run-a',
        threadId: 'thread-a',
        status: 'completed',
        model: 'Local simulated',
        context: 'local_simulated',
        events: [{ id: 'evt-delta', runId: 'run-a', threadId: 'thread-a', type: 'model.delta', label: 'Model', detail: 'delta', time: 'Now', status: 'running', assistantDelta: 'Hello! How can I help?' }],
      },
      loading: false,
      error: null,
      dataSourceMode: 'real_api',
      streamState: 'closed',
      onSendMessage: () => {},
      onStopRun: () => {},
      locale: 'en',
    }))

    expect(html.match(/Hello! How can I help\?/g)).toHaveLength(1)
    expect(html).not.toContain('Hello!HowcanIhelp')
  })

  test('folds duplicate persisted assistant finals that only differ by whitespace', () => {
    const html = renderToStaticMarkup(createElement(ChatCanvas, {
      sidebarCollapsed: false,
      thread: { id: 'thread-a', title: 'Thread A', project: 'Loomi', mode: 'chat', updatedAt: 'Now', lifecycleStatus: 'active', runStatus: 'completed' },
      messages: [
        { id: 'msg-user', threadId: 'thread-a', role: 'user', content: 'hello', createdAt: 'Now' },
        { id: 'msg-a', threadId: 'thread-a', role: 'assistant', content: 'Hello! How can I help?', createdAt: 'Now' },
        { id: 'msg-b', threadId: 'thread-a', role: 'assistant', content: 'Hello!HowcanIhelp?', createdAt: 'Now' },
      ],
      run: null,
      loading: false,
      error: null,
      dataSourceMode: 'real_api',
      streamState: 'closed',
      onSendMessage: () => {},
      onStopRun: () => {},
      locale: 'en',
    }))

    expect(html.match(/Hello! How can I help\?/g)).toHaveLength(1)
    expect(html).not.toContain('Hello!HowcanIhelp')
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

  test('keeps assistant message actions quiet until the message is hovered or focused', () => {
    const css = readFileSync(resolve(import.meta.dir, '../styles/20-chat.css'), 'utf8')

    expect(css).toContain('.message-row.assistant .message-actions')
    expect(css).toContain('opacity: 0')
    expect(css).toContain('.message-row.assistant:hover .message-actions')
    expect(css).toContain('.message-row.assistant:focus-within .message-actions')
    expect(css).toContain('opacity: 1')
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

  test('renders whole markdown code fences as assistant markdown instead of source code', () => {
    const html = renderToStaticMarkup(createElement(ChatCanvas, {
      sidebarCollapsed: false,
      thread: { id: 'thread-a', title: 'Thread A', project: 'Loomi', mode: 'chat', updatedAt: 'Now', lifecycleStatus: 'active', runStatus: 'completed' },
      messages: [{ id: 'msg-a', threadId: 'thread-a', role: 'assistant', content: '```markdown\n# 项目名称\n\n一句话介绍这个项目是做什么的。\n\n## 目录\n- [项目简介](#项目简介)\n\n```bash\necho "Hello"\n```\n```', createdAt: 'Now' }],
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

    expect(html).toContain('<h1>项目名称</h1>')
    expect(html).toContain('<h2>目录</h2>')
    expect(html).toContain('<li><a href="#" rel="noreferrer" target="_blank">项目简介</a></li>')
    expect(html).toContain('<span class="message-code-block-lang">bash</span>')
    expect(html).not.toContain('message-code-block-lang">markdown')
    expect(html).not.toContain('```markdown')
  })

  test('renders dense bare README-like markdown as a document artifact instead of one oversized chat heading', () => {
    const html = renderToStaticMarkup(createElement(ChatCanvas, {
      sidebarCollapsed: false,
      thread: { id: 'thread-a', title: 'Thread A', project: 'Loomi', mode: 'chat', updatedAt: 'Now', lifecycleStatus: 'active', runStatus: 'completed' },
      messages: [{ id: 'msg-a', threadId: 'thread-a', role: 'assistant', content: 'markdown#项目名称>简短描述：这个项目用于解决什么问题，适合什么场景。###项目介绍这里填写项目的背景、目标和主要用途。##主要功能-功能一：说明功能用途-功能二：说明功能用途-支持快速部署###技术栈-前端：React/Vue/HTML-后端：Node.js/Python/Java##快速开始#### 1. 克隆项目#### 2. 安装依赖### 3. 启动开发环境## 使用说明打开浏览器访问：## 项目结构## 环境变量## 常用命令', createdAt: 'Now' }],
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

    expect(html).toContain('message-artifact-card')
    expect(html).toContain('项目名称')
    expect(html).toContain('Markdown 文档')
    expect(html).not.toContain('message-markdown\"></div>')
    expect(html).not.toContain('<h1>项目名称&gt;简短描述')
    expect(html).not.toContain('###项目介绍')
    expect(html).not.toContain('##快速开始')
  })

  test('renders raw SVG assistant output as a visual artifact card', () => {
    const html = renderToStaticMarkup(createElement(ChatCanvas, {
      sidebarCollapsed: false,
      thread: { id: 'thread-a', title: 'Thread A', project: 'Loomi', mode: 'chat', updatedAt: 'Now', lifecycleStatus: 'active', runStatus: 'completed' },
      messages: [{ id: 'msg-svg', threadId: 'thread-a', role: 'assistant', content: '下面是图：\n```svg\n<svg viewBox="0 0 20 20"><title>流程图</title><rect width="20" height="20"/></svg>\n```', createdAt: 'Now' }],
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

    expect(html).toContain('message-artifact-card')
    expect(html).toContain('流程图')
    expect(html).toContain('可视化产物')
    expect(html).not.toContain('message-code-block-lang">svg')
    expect(html).not.toContain('&lt;svg')
  })

  test('does not render trailing empty fences as blank code cards', () => {
    const html = renderToStaticMarkup(createElement(ChatCanvas, {
      sidebarCollapsed: false,
      thread: { id: 'thread-a', title: 'Thread A', project: 'Loomi', mode: 'chat', updatedAt: 'Now', lifecycleStatus: 'active', runStatus: 'completed' },
      messages: [{ id: 'msg-a', threadId: 'thread-a', role: 'assistant', content: '## 许可证\n\nMIT\n\n```\n```', createdAt: 'Now' }],
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

    expect(html).toContain('<h2>许可证</h2>')
    expect(html).toContain('<p>MIT</p>')
    expect(html).not.toContain('message-code-block')
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

  test('asks for Downloads authorization when the user requests Downloads but another workspace is selected', () => {
    const html = renderToStaticMarkup(createElement(ChatCanvas, {
      sidebarCollapsed: false,
      thread: { id: 'thread-a', title: 'Thread A', project: 'Loomi', mode: 'chat', updatedAt: 'Now', lifecycleStatus: 'active', runStatus: 'completed' },
      messages: [
        { id: 'msg-user', threadId: 'thread-a', role: 'user', content: '看下我下载目录整理一下', createdAt: 'Now' },
        { id: 'msg-assistant', threadId: 'thread-a', role: 'assistant', content: '我现在没有可用的本地文件/目录访问工具。', createdAt: 'Now' },
      ],
      run: null,
      loading: false,
      error: null,
      dataSourceMode: 'real_api',
      streamState: 'closed',
      providerCapabilities: [{ id: 'custom', family: 'openai_compatible' as const, model: 'gpt-5.5', status: 'available' as const }],
      workspaceRootConfig: { configured: true, displayName: 'Loomi' },
      onChooseWorkspaceFolder: () => {},
      onSendMessage: () => {},
      onStopRun: () => {},
      locale: 'zh',
    }))

    expect(html).toContain('授权下载目录')
    expect(html).toContain('选择下载目录')
    expect(html).toContain('当前工作区是 Loomi')
    expect(html).not.toContain('我现在没有可用的本地文件/目录访问工具')
  })

  test('does not ask for Downloads authorization when Downloads is already selected', () => {
    const html = renderToStaticMarkup(createElement(ChatCanvas, {
      sidebarCollapsed: false,
      thread: { id: 'thread-a', title: 'Thread A', project: 'Downloads', mode: 'chat', updatedAt: 'Now', lifecycleStatus: 'active', runStatus: 'completed' },
      messages: [
        { id: 'msg-user', threadId: 'thread-a', role: 'user', content: '看下我下载目录整理一下', createdAt: 'Now' },
        { id: 'msg-assistant', threadId: 'thread-a', role: 'assistant', content: '我来整理。', createdAt: 'Now' },
      ],
      run: null,
      loading: false,
      error: null,
      dataSourceMode: 'real_api',
      streamState: 'closed',
      providerCapabilities: [{ id: 'custom', family: 'openai_compatible' as const, model: 'gpt-5.5', status: 'available' as const }],
      workspaceRootConfig: { configured: true, displayName: 'Downloads' },
      onChooseWorkspaceFolder: () => {},
      onSendMessage: () => {},
      onStopRun: () => {},
      locale: 'zh',
    }))

    expect(html).not.toContain('授权下载目录')
  })

  test('asks for Documents authorization when the user requests the macOS documents folder', () => {
    const html = renderToStaticMarkup(createElement(ChatCanvas, {
      sidebarCollapsed: false,
      thread: { id: 'thread-a', title: 'Thread A', project: 'Downloads', mode: 'chat', updatedAt: 'Now', lifecycleStatus: 'active', runStatus: 'completed' },
      messages: [
        { id: 'msg-user', threadId: 'thread-a', role: 'user', content: '帮忙列下文稿目录内容 分类下', createdAt: 'Now' },
        { id: 'msg-assistant', threadId: 'thread-a', role: 'assistant', content: '不能直接查看“文稿/Documents”目录。', createdAt: 'Now' },
      ],
      run: null,
      loading: false,
      error: null,
      dataSourceMode: 'real_api',
      streamState: 'closed',
      providerCapabilities: [{ id: 'custom', family: 'openai_compatible' as const, model: 'gpt-5.5', status: 'available' as const }],
      workspaceRootConfig: { configured: true, displayName: 'Downloads' },
      onChooseWorkspaceFolder: () => {},
      onSendMessage: () => {},
      onStopRun: () => {},
      locale: 'zh',
    }))

    expect(html).toContain('授权文稿目录')
    expect(html).toContain('选择文稿目录')
    expect(html).not.toContain('不能直接查看')
  })

  test('does not ask for Documents authorization when Documents is already selected', () => {
    const html = renderToStaticMarkup(createElement(ChatCanvas, {
      sidebarCollapsed: false,
      thread: { id: 'thread-a', title: 'Thread A', project: 'Documents', mode: 'chat', updatedAt: 'Now', lifecycleStatus: 'active', runStatus: 'completed' },
      messages: [
        { id: 'msg-user', threadId: 'thread-a', role: 'user', content: '帮忙列下文稿目录内容 分类下', createdAt: 'Now' },
      ],
      run: null,
      loading: false,
      error: null,
      dataSourceMode: 'real_api',
      streamState: 'closed',
      providerCapabilities: [{ id: 'custom', family: 'openai_compatible' as const, model: 'gpt-5.5', status: 'available' as const }],
      workspaceRootConfig: { configured: true, displayName: 'Documents' },
      onChooseWorkspaceFolder: () => {},
      onSendMessage: () => {},
      onStopRun: () => {},
      locale: 'zh',
    }))

    expect(html).not.toContain('授权文稿目录')
  })

  test('keeps workspace authorization card compact and readable in the unified theme layer', () => {
    const css = readFileSync(resolve(import.meta.dir, '../styles/92-unified-workspace.css'), 'utf8')

    expect(css).toContain('.workspace-access-card')
    expect(css).toContain('width: min(560px, calc(100% - 18px)) !important')
    expect(css).toContain('grid-template-columns: 36px minmax(0, 1fr) auto !important')
    expect(css).toContain('display: inline-flex !important')
    expect(css).toContain('.workspace-access-card button span')
    expect(css).toContain('.workspace-access-card button svg')
    expect(css).toContain("color: var(--loomi-bg) !important")
    expect(css).toContain(".app-shell[data-theme='dark'] .workspace-access-card button")
    expect(css).toContain('color: #171717 !important')
    expect(css).toContain('@media (max-width: 720px)')
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
    const toolIndex = html.indexOf('完成 1 个工具')
    const continuationIndex = html.indexOf('查完了，重点是模型更新。')

    expect(firstTextIndex).toBeGreaterThan(-1)
    expect(toolIndex).toBeGreaterThan(firstTextIndex)
    expect(continuationIndex).toBeGreaterThan(toolIndex)
    expect(html).toContain('搜索 1')
    expect(html).not.toContain('class="tool-card status-succeeded')
  })

  test('renders final event content after completed tool events', () => {
    const html = renderToStaticMarkup(createElement(ChatCanvas, {
      sidebarCollapsed: false,
      thread: { id: 'thread-a', title: 'Thread A', project: 'Loomi', mode: 'chat', updatedAt: 'Now', lifecycleStatus: 'active', runStatus: 'completed' },
      messages: [{ id: 'msg-user', threadId: 'thread-a', role: 'user', content: '查一下再回答', createdAt: 'Now' }],
      run: {
        id: 'run-a',
        threadId: 'thread-a',
        status: 'completed',
        model: 'gpt-5.5',
        context: 'openai_compatible',
        events: [
          { id: 'evt-tool-start', runId: 'run-a', threadId: 'thread-a', sequence: 1, type: 'tool.call.executing', label: 'tool', detail: 'Search running', time: 'Now', status: 'running', group: 'tool-call', metadata: { tool_call_id: 'tc-search', tool_name: 'web.search', execution_status: 'executing' } },
          { id: 'evt-tool-done', runId: 'run-a', threadId: 'thread-a', sequence: 2, type: 'tool.call.succeeded', label: 'tool', detail: 'Search completed', time: 'Now', status: 'completed', group: 'tool-call', metadata: { tool_call_id: 'tc-search', tool_name: 'web.search', result_summary: { count: 3 }, execution_status: 'succeeded' } },
          { id: 'evt-final', runId: 'run-a', threadId: 'thread-a', sequence: 3, type: 'message.model_output_completed', label: 'message', detail: 'Final answer', time: 'Later', status: 'completed', content: '## 结论\n\n工具查完了，最终回答应该显示在这里。' },
        ],
      },
      loading: false,
      error: null,
      dataSourceMode: 'real_api',
      streamState: 'closed',
      providerCapabilities: [{ id: 'custom', family: 'openai_compatible' as const, model: 'gpt-5.5', status: 'available' as const }],
      onSendMessage: () => {},
      onStopRun: () => {},
      locale: 'zh',
    }))

    const toolIndex = html.indexOf('完成 1 个工具')
    const finalIndex = html.indexOf('<h2>结论</h2>')

    expect(toolIndex).toBeGreaterThan(-1)
    expect(finalIndex).toBeGreaterThan(toolIndex)
    expect(html).toContain('工具查完了，最终回答应该显示在这里。')
    expect(html).not.toContain('未生成回复')
    expect(html).not.toContain('final-missing')
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

  test('renders inline markdown file payloads as artifact cards instead of code chips', () => {
    const html = renderToStaticMarkup(createElement(ChatCanvas, {
      sidebarCollapsed: false,
      thread: { id: 'thread-a', title: 'Thread A', project: 'Loomi', mode: 'chat', updatedAt: 'Now', lifecycleStatus: 'active', runStatus: 'completed' },
      messages: [{
        id: 'msg-file-content',
        threadId: 'thread-a',
        role: 'assistant',
        content: '如果你想要“一个 Markdown 文件内容”，直接用下面这个通用版： `markdown #文档标题##概述这里写文档的简要说明。 ##目标-目标一-目标二##内容###1. 第一部分这里写第一部分内容。` 如果你要特定类型，告诉我类型即可。',
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

    expect(html).toContain('如果你想要“一个 Markdown 文件内容”')
    expect(html).toContain('文档标题')
    expect(html).toContain('Markdown 文档')
    expect(html).not.toContain('<code>markdown')
    expect(html).not.toContain('##概述这里写文档的简要说明')
  })

  test('renders completed draft markdown file payloads as artifact cards', () => {
    const html = renderToStaticMarkup(createElement(ChatCanvas, {
      sidebarCollapsed: false,
      thread: { id: 'thread-a', title: 'Thread A', project: 'Loomi', mode: 'chat', updatedAt: 'Now', lifecycleStatus: 'active', runStatus: 'completed' },
      messages: [{ id: 'msg-user', threadId: 'thread-a', role: 'user', content: '?', createdAt: 'Now' }],
      run: {
        id: 'run-a',
        threadId: 'thread-a',
        status: 'completed',
        model: 'gpt-5.5',
        context: 'openai_compatible',
        events: [],
        assistantDraft: {
          status: 'completed',
          content: '如果你想要“一个Markdown文件内容”，直接用下面这个通用版： `markdown #文档标题##概述这里写文档的简要说明。##目标-目标一-目标二-目标三##内容###1.第一部分这里写第一部分内容。###2.第二部分这里写第二部分内容。##待办事项-[]任务一-[]任务二-[]任务三##备注这里写补充说明。` 如果你要特定类型，比如README、周报、简历、会议纪要，告诉我类型即可。',
        },
      },
      loading: false,
      error: null,
      dataSourceMode: 'real_api',
      streamState: 'closed',
      providerCapabilities: [{ id: 'custom', family: 'openai_compatible' as const, model: 'gpt-5.5', status: 'available' as const }],
      onSendMessage: () => {},
      onStopRun: () => {},
      locale: 'zh',
    }))

    expect(html).toContain('文档标题')
    expect(html).toContain('Markdown 文档')
    expect(html).not.toContain('<code>markdown')
  })

  test('renders run transcript markdown file payloads as artifact cards', () => {
    const html = renderToStaticMarkup(createElement(ChatCanvas, {
      sidebarCollapsed: false,
      thread: { id: 'thread-a', title: 'Thread A', project: 'Loomi', mode: 'chat', updatedAt: 'Now', lifecycleStatus: 'active', runStatus: 'completed' },
      messages: [{ id: 'msg-user', threadId: 'thread-a', role: 'user', content: '?', createdAt: 'Now' }],
      run: {
        id: 'run-a',
        threadId: 'thread-a',
        status: 'completed',
        model: 'gpt-5.5',
        context: 'openai_compatible',
        events: [{
          id: 'evt-delta',
          runId: 'run-a',
          threadId: 'thread-a',
          type: 'model.delta',
          label: 'Model',
          detail: 'delta',
          time: 'Now',
          status: 'running',
          assistantDelta: '如果你想要“一个Markdown文件内容”，直接用下面这个通用版：markdown #文档标题##概述这里写文档的简要说明。##目标-目标一如果你要特定类型即可。',
        }],
        assistantDraft: { status: 'completed', content: '' },
      },
      loading: false,
      error: null,
      dataSourceMode: 'real_api',
      streamState: 'closed',
      providerCapabilities: [{ id: 'custom', family: 'openai_compatible' as const, model: 'gpt-5.5', status: 'available' as const }],
      onSendMessage: () => {},
      onStopRun: () => {},
      locale: 'zh',
    }))

    expect(html).toContain('文档标题')
    expect(html).toContain('Markdown 文档')
    expect(html).not.toContain('<code>markdown')
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
    const chatCss = readFileSync(resolve(import.meta.dir, '../styles/20-chat.css'), 'utf8')
    const css = [
      chatCss,
      readFileSync(resolve(import.meta.dir, '../styles/80-island-components.css'), 'utf8'),
      readFileSync(resolve(import.meta.dir, '../styles/92-unified-workspace.css'), 'utf8'),
    ].join('\n')

    expect(css).toContain('.message-code-block')
    expect(css).toContain('border: 1px solid color-mix(in srgb, var(--border-subtle) 78%, var(--text-primary) 10%)')
    expect(css).toContain('border-radius: 10px')
    expect(css).toContain('padding: 13px 14px 15px')
    expect(css).not.toContain('padding: 48px 48px 24px 34px')
    expect(css).toContain('min-height: 32px')
    expect(chatCss).not.toContain('top: 20px')
    expect(chatCss).not.toContain('left: 22px')
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

  test('repairs collapsed pipe tables before rendering assistant messages', () => {
    const html = renderToStaticMarkup(createElement(ChatCanvas, {
      sidebarCollapsed: false,
      thread: { id: 'thread-a', title: 'Thread A', project: 'Loomi', mode: 'chat', updatedAt: 'Now', lifecycleStatus: 'active', runStatus: 'completed' },
      messages: [{
        id: 'msg-table-dense',
        threadId: 'thread-a',
        role: 'assistant',
        content: '##按文件类型分类|类型|数量||---|---:||Markdown文档|35||Excel表格|2|##按大类分类|大类|数量||---|---:||文档|37||图片|2|',
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

    expect(html).toContain('<h2>按文件类型分类</h2>')
    expect(html).toContain('<table>')
    expect(html).toContain('<td>Markdown文档</td>')
    expect(html).toContain('<h2>按大类分类</h2>')
    expect(html).not.toContain('||---')
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
    expect(html).toContain('2 tools completed')
    expect(html).toContain('inspected 1 · searched 1')
    expect(html).not.toContain('Find project files')
    expect(html).not.toContain('Search project text')
  })

  test('collapses a single completed tool event instead of rendering a full tool card', () => {
    const html = renderToStaticMarkup(createElement(ChatCanvas, {
      sidebarCollapsed: false,
      thread: { id: 'thread-a', title: 'Thread A', project: 'Loomi', mode: 'chat', updatedAt: 'Now', lifecycleStatus: 'active', runStatus: 'completed' },
      messages: [{ id: 'msg-user', threadId: 'thread-a', role: 'user', content: 'Search first', createdAt: 'Now' }],
      run: {
        id: 'run-a',
        threadId: 'thread-a',
        status: 'completed',
        model: 'local_codex',
        context: 'model_gateway',
        events: [
          { id: 'evt-tool-1', runId: 'run-a', threadId: 'thread-a', type: 'tool.call.succeeded', label: 'Tool', detail: 'done', time: 'Now', status: 'completed', sequence: 1, metadata: { tool_call_id: 'tc-1', tool_name: 'web.search', arguments_summary: { query: 'latest AI news', provider: 'brave', limit: 5 }, result_summary: { result_count: 2 }, execution_status: 'succeeded' } },
          { id: 'evt-delta', runId: 'run-a', threadId: 'thread-a', type: 'message.model_output_delta', label: 'Message', detail: 'delta', time: 'Now', status: 'running', sequence: 2, content: '搜索完了。' },
          { id: 'evt-final', runId: 'run-a', threadId: 'thread-a', type: 'run.completed', label: 'Run', detail: 'Run completed', time: 'Now', status: 'completed', sequence: 3 },
        ],
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

    expect(html).toContain('tool-stack')
    expect(html).toContain('完成 1 个工具')
    expect(html).toContain('搜索 1')
    expect(html).toContain('搜索完了。')
    expect(html).not.toContain('class="tool-card status-succeeded')
    expect(html).not.toContain('请求')
    expect(html).not.toContain('结果')
  })

  test('keeps completed artifact tool events as preview resource cards', () => {
    const html = renderToStaticMarkup(createElement(ChatCanvas, {
      sidebarCollapsed: false,
      thread: { id: 'thread-a', title: 'Thread A', project: 'Loomi', mode: 'chat', updatedAt: 'Now', lifecycleStatus: 'active', runStatus: 'completed' },
      messages: [{ id: 'msg-user', threadId: 'thread-a', role: 'user', content: '写个 md', createdAt: 'Now' }],
      run: {
        id: 'run-a',
        threadId: 'thread-a',
        status: 'completed',
        model: 'local_codex',
        context: 'model_gateway',
        events: [
          { id: 'evt-tool-1', runId: 'run-a', threadId: 'thread-a', type: 'tool.call.succeeded', label: 'Tool', detail: 'done', time: 'Now', status: 'completed', sequence: 1, metadata: { tool_call_id: 'tc-1', tool_name: 'artifact.create_text', result_summary: { operation: 'create_text', artifact_id: 'art-1', title: '我的 Markdown 文档', filename: '我的文档.md', mime_type: 'text/markdown', text_excerpt: '# 我的 Markdown 文档' }, execution_status: 'succeeded' } },
          { id: 'evt-delta', runId: 'run-a', threadId: 'thread-a', type: 'message.model_output_delta', label: 'Message', detail: 'delta', time: 'Now', status: 'running', sequence: 2, content: '已经写好了。' },
          { id: 'evt-final', runId: 'run-a', threadId: 'thread-a', type: 'run.completed', label: 'Run', detail: 'Run completed', time: 'Now', status: 'completed', sequence: 3 },
        ],
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

    expect(html).toContain('artifact-resource-card')
    expect(html).toContain('我的 Markdown 文档')
    expect(html).toContain('已经写好了。')
    expect(html).not.toContain('完成 1 个工具')
    expect(html).not.toContain('artifact.create_text')
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

    expect(html).toContain('完成 1 个工具')
    expect(html).toContain('搜索 1')
    expect(html).toContain('生成中')
    expect(html).toMatch(/组织回复|梳理线索|核对上下文|提炼重点|推敲答案|收束思路|准备回答|再看一眼/)
  })

  test('keeps failure feedback visible after tool events when no final answer arrives', () => {
    const html = renderToStaticMarkup(createElement(ChatCanvas, {
      sidebarCollapsed: false,
      thread: { id: 'thread-a', title: 'Thread A', project: 'Loomi', mode: 'chat', updatedAt: 'Now', lifecycleStatus: 'active', runStatus: 'failed' },
      messages: [{ id: 'msg-user', threadId: 'thread-a', role: 'user', content: 'Search first', createdAt: 'Now' }],
      run: {
        id: 'run-a',
        threadId: 'thread-a',
        status: 'failed',
        model: 'local_codex',
        context: 'model_gateway',
        events: [
          { id: 'evt-tool-1', runId: 'run-a', threadId: 'thread-a', type: 'tool.call.succeeded', label: 'Tool', detail: 'done', time: 'Now', status: 'completed', sequence: 1, metadata: { tool_call_id: 'tc-1', tool_name: 'web.search', result_summary: { count: 2 }, execution_status: 'succeeded' } },
          { id: 'evt-failed', runId: 'run-a', threadId: 'thread-a', type: 'run.failed', label: 'Run failed', detail: 'Provider failed', time: 'Now', status: 'failed', sequence: 2 },
        ],
      },
      loading: false,
      error: null,
      dataSourceMode: 'real_api',
      streamState: 'closed',
      providerCapabilities: [{ id: 'local_codex', family: 'openai_compatible', model: 'gpt-5.5', status: 'available', executionState: 'supported' }],
      onSendMessage: () => {},
      onStopRun: () => {},
      onRetryRun: () => {},
      locale: 'zh',
    }))

    expect(html).toContain('完成 1 个工具')
    expect(html).toContain('执行失败')
    expect(html).toContain('未生成成功回复')
    expect(html).toContain('重试')
  })

  test('keeps recovery feedback visible after terminal tool events before retry output arrives', () => {
    const html = renderToStaticMarkup(createElement(ChatCanvas, {
      sidebarCollapsed: false,
      thread: { id: 'thread-a', title: 'Thread A', project: 'Loomi', mode: 'chat', updatedAt: 'Now', lifecycleStatus: 'active', runStatus: 'recovering' },
      messages: [{ id: 'msg-user', threadId: 'thread-a', role: 'user', content: 'Search first', createdAt: 'Now' }],
      run: {
        id: 'run-a',
        threadId: 'thread-a',
        status: 'recovering',
        model: 'local_codex',
        context: 'model_gateway',
        events: [
          { id: 'evt-tool-1', runId: 'run-a', threadId: 'thread-a', type: 'tool.call.succeeded', label: 'Tool', detail: 'done', time: 'Now', status: 'completed', sequence: 1, metadata: { tool_call_id: 'tc-1', tool_name: 'web.search', result_summary: { count: 2 }, execution_status: 'succeeded' } },
        ],
      },
      loading: false,
      error: null,
      dataSourceMode: 'real_api',
      streamState: 'recoverable_error',
      providerCapabilities: [{ id: 'local_codex', family: 'openai_compatible', model: 'gpt-5.5', status: 'available', executionState: 'supported' }],
      onSendMessage: () => {},
      onStopRun: () => {},
      locale: 'zh',
    }))

    expect(html).toContain('完成 1 个工具')
    expect(html).toContain('恢复中')
    expect(html).toContain('恢复中…')
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

  test('keeps active approval requests in the chat transcript without a duplicate top notice', () => {
    const html = renderToStaticMarkup(createElement(ChatCanvas, {
      sidebarCollapsed: false,
      thread: { id: 'thread-a', title: 'Thread A', project: 'Loomi', mode: 'chat', updatedAt: 'Now', lifecycleStatus: 'active', runStatus: 'blocked_on_tool_approval' },
      messages: [{ id: 'msg-user', threadId: 'thread-a', role: 'user', content: '读取 README', createdAt: 'Now' }],
      run: {
        id: 'run-a',
        threadId: 'thread-a',
        status: 'blocked_on_tool_approval',
        model: 'local_codex',
        context: 'model_gateway',
        events: [{ id: 'evt-approval', runId: 'run-a', threadId: 'thread-a', type: 'tool.call.approval_required', label: 'Tool approval required', detail: 'waiting', time: 'Now', status: 'blocked_on_tool_approval' }],
        toolCalls: [{
          id: 'tool-approval',
          toolCallId: 'tc-approval',
          name: 'workspace.read_file',
          status: 'approval_required',
          approvalStatus: 'required',
          executionStatus: 'blocked',
          summary: 'Read README.md',
          input: '',
          output: '',
          argumentsSummary: { path: 'README.md' },
          resultSummary: null,
          errorCode: null,
          errorMessage: null,
        }],
      },
      loading: false,
      error: null,
      dataSourceMode: 'real_api',
      streamState: 'closed',
      providerCapabilities: [{ id: 'local_codex', family: 'openai_compatible', model: 'gpt-5.5', status: 'available', executionState: 'supported' }],
      onSendMessage: () => {},
      onStopRun: () => {},
      onApproveToolCall: () => {},
      onDenyToolCall: () => {},
      locale: 'zh',
    }))

    expect(html).toContain('tool-card status-approval_required')
    expect(html).toContain('等待确认')
    expect(html).toContain('允许')
    expect(html).toContain('拒绝')
    expect(html).not.toContain('approval-notice')
    expect(html).not.toContain('确认下面的工具请求后')
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
