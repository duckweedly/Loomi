import { describe, expect, test } from 'bun:test'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { createElement } from 'react'
import { renderToStaticMarkup } from 'react-dom/server'
import { RunRail } from './RunRail'

describe('RunRail restrained runtime polish', () => {
  test('uses compact Scenario controls instead of prominent script buttons', () => {
    const source = readFileSync(resolve(import.meta.dir, 'RunRail.tsx'), 'utf8')

    expect(source).toContain('Scenario')
    expect(source).toContain('Success')
    expect(source).toContain('Fail')
    expect(source).not.toContain('成功剧本')
    expect(source).not.toContain('失败剧本')
  })

  test('uses a ghost stop action label', () => {
    const source = readFileSync(resolve(import.meta.dir, 'RunRail.tsx'), 'utf8')

    expect(source).toContain('Stop run')
    expect(source).toContain('runtime-stop-button ghost')
  })

  test('styles timeline with quiet dots and compact agent card', () => {
    const css = readFileSync(resolve(import.meta.dir, '../styles.css'), 'utf8')

    expect(css).toContain('.progress-row::before')
    expect(css).toContain('width: 7px')
    expect(css).toContain('.agent-motion-card.compact')
    expect(css).toContain('.runtime-script-switch.compact')
  })

  test('renders stable grouped timeline sections with error and usage details', () => {
    const html = renderToStaticMarkup(createElement(RunRail, {
      open: true,
      onOpenArtifact: () => {},
      run: {
        id: 'run-a',
        threadId: 'thread-a',
        status: 'failed',
        model: 'Local simulated',
        context: 'local_simulated',
        events: [
          { id: 'evt-run', runId: 'run-a', threadId: 'thread-a', sequence: 1, type: 'run.created', label: 'Run', detail: 'created', time: 'Now', status: 'running' },
          { id: 'evt-model', runId: 'run-a', threadId: 'thread-a', sequence: 2, type: 'model.usage', label: 'Usage', detail: 'usage', time: 'Now', status: 'running', usage: { inputTokens: 7, outputTokens: 11 } },
          { id: 'evt-worker', runId: 'run-a', threadId: 'thread-a', sequence: 3, type: 'worker.claimed', label: 'Worker', detail: 'claimed', time: 'Now', status: 'running' },
          { id: 'evt-error', runId: 'run-a', threadId: 'thread-a', sequence: 4, type: 'provider.error', label: 'Provider', detail: 'provider failed', time: 'Now', status: 'failed' },
        ],
      },
    }))

    expect(html).toContain('Run lifecycle')
    expect(html).toContain('Model stream')
    expect(html).toContain('Worker/job')
    expect(html).toContain('Error')
    expect(html).toContain('7 in / 11 out')
    expect(html).toContain('provider failed')
    expect(html).toContain('runtime-event-group error')
  })

  test('renders productized M6 worker event labels and unknown fallback', () => {
    const html = renderToStaticMarkup(createElement(RunRail, {
      open: true,
      onOpenArtifact: () => {},
      run: {
        id: 'run-a',
        threadId: 'thread-a',
        status: 'recovering',
        model: 'Local simulated',
        context: 'local_simulated',
        events: [
          { id: 'evt-claim', runId: 'run-a', threadId: 'thread-a', sequence: 1, type: 'job_claimed', label: 'Job', detail: 'raw claim', time: 'Now', status: 'running' },
          { id: 'evt-lease', runId: 'run-a', threadId: 'thread-a', sequence: 2, type: 'lease_renewed', label: 'Lease', detail: 'raw lease', time: 'Now', status: 'running' },
          { id: 'evt-recovering', runId: 'run-a', threadId: 'thread-a', sequence: 3, type: 'job_recovering', label: 'Job', detail: 'raw recovery', time: 'Now', status: 'recovering' },
          { id: 'evt-unknown', runId: 'run-a', threadId: 'thread-a', sequence: 4, type: 'future_worker_event', label: 'Future', detail: 'raw future', time: 'Now', status: 'running' },
        ],
      },
    }))

    expect(html).toContain('Job claimed by worker')
    expect(html).toContain('Lease renewed')
    expect(html).toContain('Job recovering')
    expect(html).toContain('Unknown worker event')
    expect(html).toContain('future_worker_event')
  })

  test('renders productized worker event labels from Chinese i18n copy', () => {
    const html = renderToStaticMarkup(createElement(RunRail, {
      open: true,
      locale: 'zh',
      onOpenArtifact: () => {},
      run: {
        id: 'run-a',
        threadId: 'thread-a',
        status: 'recovering',
        model: 'Local simulated',
        context: 'local_simulated',
        events: [
          { id: 'evt-claim', runId: 'run-a', threadId: 'thread-a', sequence: 1, type: 'job_claimed', label: 'Job', detail: 'raw claim', time: 'Now', status: 'running' },
          { id: 'evt-unknown', runId: 'run-a', threadId: 'thread-a', sequence: 2, type: 'future_worker_event', label: 'Future', detail: 'raw future', time: 'Now', status: 'running' },
        ],
      },
    }))

    expect(html).toContain('Worker 已领取任务')
    expect(html).toContain('未知 Worker 事件')
  })

  test('renders provider errors and cancelled events with distinct row classes', () => {
    const html = renderToStaticMarkup(createElement(RunRail, {
      open: true,
      onOpenArtifact: () => {},
      run: {
        id: 'run-a',
        threadId: 'thread-a',
        status: 'cancelled',
        model: 'Local simulated',
        context: 'local_simulated',
        events: [
          { id: 'evt-provider', runId: 'run-a', threadId: 'thread-a', sequence: 1, type: 'provider.error', label: 'Provider', detail: 'provider failed', time: 'Now', status: 'running', group: 'error', severity: 'error' },
          { id: 'evt-cancelled', runId: 'run-a', threadId: 'thread-a', sequence: 2, type: 'run.cancelled', label: 'Cancelled', detail: 'cancelled', time: 'Now', status: 'cancelled', group: 'run-lifecycle', severity: 'warning' },
        ],
      },
    }))

    expect(html).toContain('progress-row failed')
    expect(html).toContain('progress-row warning')
    expect(html).toContain('provider failed')
    expect(html).toContain('cancelled')
  })

  test('shows capability status detail in the rail', () => {
    const html = renderToStaticMarkup(createElement(RunRail, {
      open: true,
      onOpenArtifact: () => {},
      capabilityStatus: 'provider-unavailable',
      run: null,
    }))

    expect(html).toContain('Provider unavailable')
    expect(html).toContain('provider rejected')
  })
})

describe('RunRail localized runtime copy', () => {
  test('renders Chinese runtime group and capability copy when locale is zh', () => {
    const html = renderToStaticMarkup(createElement(RunRail, {
      open: true,
      locale: 'zh',
      onOpenArtifact: () => {},
      capabilityStatus: 'provider-unavailable',
      run: {
        id: 'run-a',
        threadId: 'thread-a',
        status: 'failed',
        model: 'Local simulated',
        context: 'local_simulated',
        events: [
          { id: 'evt-run', runId: 'run-a', threadId: 'thread-a', sequence: 1, type: 'run.created', label: 'Run', detail: 'created', time: 'Now', status: 'running' },
          { id: 'evt-worker', runId: 'run-a', threadId: 'thread-a', sequence: 2, type: 'worker.claimed', label: 'Worker', detail: 'claimed', time: 'Now', status: 'running' },
          { id: 'evt-error', runId: 'run-a', threadId: 'thread-a', sequence: 3, type: 'provider.error', label: 'Provider', detail: 'provider failed', time: 'Now', status: 'failed' },
        ],
      },
    }))

    expect(html).toContain('运行生命周期')
    expect(html).toContain('模型流')
    expect(html).toContain('Worker/Job')
    expect(html).toContain('错误')
    expect(html).toContain('Provider 不可用')
    expect(html).toContain('Provider 拒绝或未能完成生成')
  })
})
