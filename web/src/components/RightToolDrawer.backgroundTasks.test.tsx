import { describe, expect, test } from 'bun:test'
import { createElement } from 'react'
import { renderToStaticMarkup } from 'react-dom/server'
import { RightToolDrawer } from './RightToolDrawer'

describe('RightToolDrawer background tasks panel', () => {
  test('renders a read-only empty state instead of a placeholder for Background tasks', () => {
    const html = renderToStaticMarkup(createElement(RightToolDrawer, {
      open: true,
      selectedPanelId: 'background-tasks',
    }))

    expect(html).toContain('Background tasks')
    expect(html).toContain('No background task is running')
    expect(html).toContain('Run a real model message')
    expect(html).not.toContain('Coming soon')
    expect(html).not.toContain('Retry job')
    expect(html).not.toContain('Recover job')
    expect(html).not.toContain('Cancel job')
  })

  test('renders Chinese background task copy from i18n', () => {
    const html = renderToStaticMarkup(createElement(RightToolDrawer, {
      open: true,
      selectedPanelId: 'background-tasks',
      locale: 'zh',
    }))

    expect(html).toContain('只读观察面板')
    expect(html).toContain('当前没有后台任务')
    expect(html).toContain('发送一条真实模型消息')
  })

  test('renders empty state when the selected run has no job evidence', () => {
    const html = renderToStaticMarkup(createElement(RightToolDrawer, {
      open: true,
      selectedPanelId: 'background-tasks',
      run: {
        id: 'run-a',
        threadId: 'thread-a',
        status: 'recovering',
        model: 'Local simulated',
        context: 'local_simulated',
        events: [],
      },
    }))

    expect(html).toContain('No background task is running')
    expect(html).not.toContain('Current run job')
    expect(html).not.toContain('run-a')
  })

  test('renders Chinese background task run status from i18n', () => {
    const html = renderToStaticMarkup(createElement(RightToolDrawer, {
      open: true,
      selectedPanelId: 'background-tasks',
      locale: 'zh',
      run: {
        id: 'run-a',
        threadId: 'thread-a',
        status: 'recovering',
        model: 'Local simulated',
        context: 'local_simulated',
        events: [
          { id: 'evt-recovering', runId: 'run-a', threadId: 'thread-a', type: 'job_recovering', label: 'Job', detail: 'attempt 2 of 3', time: '10:01', status: 'recovering' },
        ],
      },
    }))

    expect(html).toContain('当前 Run Job')
    expect(html).toContain('恢复中')
    expect(html).not.toContain('Recovering')
  })

  test('renders current run job status diagnostics and latest worker events', () => {
    const html = renderToStaticMarkup(createElement(RightToolDrawer, {
      open: true,
      selectedPanelId: 'background-tasks',
      run: {
        id: 'run-a',
        threadId: 'thread-a',
        status: 'recovering',
        model: 'Local simulated',
        context: 'local_simulated',
        events: [
          { id: 'evt-claim', runId: 'run-a', threadId: 'thread-a', type: 'job_claimed', label: 'Job', detail: 'worker-a claimed job-a', time: '10:00', status: 'running' },
          { id: 'evt-recovering', runId: 'run-a', threadId: 'thread-a', type: 'job_recovering', label: 'Job', detail: 'attempt 2 of 3', time: '10:01', status: 'recovering' },
          { id: 'evt-diagnostics', runId: 'run-a', threadId: 'thread-a', type: 'worker_diagnostics', label: 'Worker', detail: 'worker-a · lease active · attempt 2/3', time: '10:02', status: 'running' },
        ],
      },
    }))

    expect(html).toContain('Current run job')
    expect(html).toContain('run-a')
    expect(html).toContain('Recovering')
    expect(html).toContain('Worker diagnostics')
    expect(html).toContain('attempt 2/3')
    expect(html).toContain('Latest worker/job events')
    expect(html).toContain('Job claimed by worker')
    expect(html).toContain('Job recovering')
  })
})
