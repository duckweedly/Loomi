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
