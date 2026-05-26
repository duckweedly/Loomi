import { describe, expect, test } from 'bun:test'
import { createElement } from 'react'
import { renderToStaticMarkup } from 'react-dom/server'
import type { Message, Run, Thread, WorkPlanProjection } from '../domain'
import { deriveWorkPlanProjection } from '../workModeProjection'
import { ChatCanvas } from './ChatCanvas'
import { WorkPlanView } from './WorkPlanView'

const projection: WorkPlanProjection = {
  goal: 'Ship M16 work mode foundation',
  status: 'running',
  statusDetail: 'Rendering work progress',
  steps: [{ id: 'step-1', title: 'Build projection', status: 'completed' }, { id: 'step-2', title: 'Render view', status: 'running' }],
  todoSnapshot: { items: [{ id: 'todo-1', title: 'Find candidate files', status: 'completed' }, { id: 'todo-2', title: 'Read selected file', status: 'running', summary: 'Safe metadata only' }], updatedBy: 'runtime', updatedAtEventId: 'evt-todo', redactionApplied: true },
  artifacts: [{ id: 'artifact-1', title: 'Work mode plan', type: 'markdown', sourceThreadId: 'thread-work', sourceRunId: 'run-work', summary: 'Safe metadata preview', createdAt: '2026-05-25', redactionApplied: true }],
  recentEvents: [{ id: 'evt-1', type: 'work.plan.updated', detail: 'Rendering work progress', time: 'Now', status: 'running' }],
}

describe('WorkPlanView', () => {
  test('renders goal, steps, status, artifacts, and recent progress', () => {
    const html = renderToStaticMarkup(createElement(WorkPlanView, { projection, loading: false, error: null }))

    expect(html).toContain('Ship M16 work mode foundation')
    expect(html).toContain('Build projection')
    expect(html).toContain('Render view')
    expect(html).toContain('Find candidate files')
    expect(html).toContain('Read selected file')
    expect(html).toContain('Safe metadata only')
    expect(html).toContain('Updated by runtime')
    expect(html).toContain('running')
    expect(html).toContain('Work mode plan')
    expect(html).toContain('artifact-1')
    expect(html).toContain('Safe metadata preview')
    expect(html).toContain('Redacted unsafe metadata')
    expect(html).toContain('work.plan.updated')
    expect(html).not.toContain('<button')
    expect(html).not.toContain('Open artifact')
    expect(html).not.toContain('Execute')
    expect(html).not.toContain('Download')
  })

  test('renders loading and error states clearly', () => {
    const loadingHtml = renderToStaticMarkup(createElement(WorkPlanView, { projection, loading: true, error: null }))
    const errorHtml = renderToStaticMarkup(createElement(WorkPlanView, { projection, loading: false, error: 'Run events failed' }))

    expect(loadingHtml).toContain('Loading work plan')
    expect(errorHtml).toContain('Work plan unavailable')
    expect(errorHtml).toContain('Run events failed')
  })
})

describe('ChatCanvas Work mode integration', () => {
  const workThread: Thread = { id: 'thread-work', title: 'Work thread', project: 'Loomi', mode: 'work', updatedAt: 'Now', lifecycleStatus: 'active', runStatus: 'running' }
  const chatThread: Thread = { ...workThread, id: 'thread-chat', mode: 'chat' }
  const messages: Message[] = [{ id: 'msg-1', threadId: workThread.id, role: 'user', content: 'Build Work mode UI', createdAt: 'Now' }]
  const run: Run = {
    id: 'run-work',
    threadId: workThread.id,
    status: 'running',
    model: 'Mock',
    context: 'local_simulated',
    events: [{ id: 'evt-1', type: 'work.plan.updated', label: 'Work', detail: 'Projected from run events', time: 'Now', status: 'running', metadata: { work_goal: 'Projected goal', work_steps: [{ title: 'Projected step', status: 'running' }] } }],
  }

  test('mounts Work Plan View for Work mode threads', () => {
    const html = renderToStaticMarkup(createElement(ChatCanvas, {
      sidebarCollapsed: false,
      thread: workThread,
      messages,
      run,
      loading: false,
      error: null,
      dataSourceMode: 'mock',
      streamState: 'closed',
      onSendMessage: () => {},
      onStopRun: () => {},
      locale: 'en',
    }))

    expect(html).toContain('Work plan')
    expect(html).toContain('Projected goal')
    expect(html).toContain('Projected step')
    expect(html).toContain('<textarea class="composer-input" disabled=""')
    expect(html).toContain('Stop</button>')
  })

  test('keeps Stop visible while Work mode is blocked on tool approval', () => {
    const html = renderToStaticMarkup(createElement(ChatCanvas, {
      sidebarCollapsed: false,
      thread: workThread,
      messages,
      run: { ...run, status: 'blocked_on_tool_approval' },
      loading: false,
      error: null,
      dataSourceMode: 'mock',
      streamState: 'open',
      onSendMessage: () => {},
      onStopRun: () => {},
      locale: 'en',
    }))

    expect(html).toContain('Work plan')
    expect(html).toContain('<textarea class="composer-input" disabled=""')
    expect(html).toContain('Stop</button>')
  })

  test('allows Work mode Composer when provider is available and no run is active', () => {
    const html = renderToStaticMarkup(createElement(ChatCanvas, {
      sidebarCollapsed: false,
      thread: { ...workThread, runStatus: 'completed' },
      messages,
      run: { ...run, status: 'completed', events: run.events },
      loading: false,
      error: null,
      dataSourceMode: 'real_api',
      streamState: 'closed',
      providerCapabilities: [{ id: 'local_codex', family: 'openai_compatible', model: 'gpt-5.5', status: 'available', executionState: 'supported' }],
      onSendMessage: () => {},
      onStopRun: () => {},
      locale: 'en',
    }))

    expect(html).toContain('Work plan')
    expect(html).toContain('<textarea class="composer-input"')
    expect(html).not.toContain('<textarea class="composer-input" disabled=""')
    expect(html).not.toContain('M16 Work mode is read-only for plan and progress')
  })

  test('keeps Work mode Composer disabled for provider unavailable states', () => {
    const html = renderToStaticMarkup(createElement(ChatCanvas, {
      sidebarCollapsed: false,
      thread: { ...workThread, runStatus: 'completed' },
      messages,
      run: { ...run, status: 'completed', events: run.events },
      loading: false,
      error: null,
      dataSourceMode: 'real_api',
      streamState: 'closed',
      providerCapabilities: [],
      onSendMessage: () => {},
      onStopRun: () => {},
      locale: 'en',
    }))

    expect(html).toContain('provider-warning')
    expect(html).toContain('<textarea class="composer-input" disabled=""')
  })

  test('keeps Chat mode isolated from Work Plan View', () => {
    const html = renderToStaticMarkup(createElement(ChatCanvas, {
      sidebarCollapsed: false,
      thread: chatThread,
      messages: messages.map((message) => ({ ...message, threadId: chatThread.id })),
      run: { ...run, threadId: chatThread.id },
      loading: false,
      error: null,
      dataSourceMode: 'mock',
      streamState: 'closed',
      onSendMessage: () => {},
      onStopRun: () => {},
      locale: 'en',
    }))

    expect(deriveWorkPlanProjection(chatThread, messages, run)).toBeNull()
    expect(html).not.toContain('Work plan')
    expect(html).not.toContain('Projected goal')
    expect(html).toContain('Build Work mode UI')
    expect(html).toContain('<textarea class="composer-input" disabled=""')
  })

  test('keeps Chat mode Composer available when provider is available and no run is active', () => {
    const html = renderToStaticMarkup(createElement(ChatCanvas, {
      sidebarCollapsed: false,
      thread: { ...chatThread, runStatus: 'completed' },
      messages: messages.map((message) => ({ ...message, threadId: chatThread.id })),
      run: { ...run, threadId: chatThread.id, status: 'completed', events: run.events },
      loading: false,
      error: null,
      dataSourceMode: 'real_api',
      streamState: 'closed',
      providerCapabilities: [{ id: 'local_codex', family: 'openai_compatible', model: 'gpt-5.5', status: 'available', executionState: 'supported' }],
      onSendMessage: () => {},
      onStopRun: () => {},
      locale: 'en',
    }))

    expect(html).not.toContain('Work plan')
    expect(html).toContain('<textarea class="composer-input"')
    expect(html).not.toContain('<textarea class="composer-input" disabled=""')
  })
})
