import { describe, expect, test } from 'bun:test'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { createElement } from 'react'
import { renderToStaticMarkup } from 'react-dom/server'
import { RunTimeline } from './RunTimeline'

describe('RunTimeline runtime linkage', () => {
  test('feeds selected run events through RunRail', () => {
    const source = readFileSync(resolve(import.meta.dir, 'RunTimeline.tsx'), 'utf8')

    expect(source).toContain('<RunRail run={run}')
  })

  test('RunRail renders failed and stopped statuses without marking them done', () => {
    const source = readFileSync(resolve(import.meta.dir, 'RunRail.tsx'), 'utf8')

    expect(source).toContain("event.status === 'failed'")
    expect(source).toContain("event.status === 'stopped'")
    expect(source).toContain('onStopRun')
  })

  test('RunRail labels model gateway provider and tool-boundary rows', () => {
    const source = readFileSync(resolve(import.meta.dir, 'RunRail.tsx'), 'utf8')

    expect(source).toContain('getEventDetail')
    expect(source).toContain('Provider failure')
    expect(source).toContain('Tool request blocked')
  })

  test('RunRail exposes a compact mock script selector for failure smoke', () => {
    const source = readFileSync(resolve(import.meta.dir, 'RunRail.tsx'), 'utf8')

    expect(source).toContain('onSelectRuntimeScript')
    expect(source).toContain('Scenario')
    expect(source).toContain('Fail')
  })

  test('renders mixed lifecycle model worker and error groups through RunTimeline', () => {
    const html = renderToStaticMarkup(createElement(RunTimeline, {
      runDetailsOpen: true,
      rightPanelMenuOpen: false,
      rightToolsOpen: false,
      selectedPanelId: 'activity',
      onSelectPanel: () => {},
      onOpenArtifact: () => {},
      run: {
        id: 'run-a',
        threadId: 'thread-a',
        status: 'failed',
        model: 'Local simulated',
        context: 'local_simulated',
        events: [
          { id: 'evt-run', runId: 'run-a', threadId: 'thread-a', sequence: 1, type: 'run.created', label: 'Run', detail: 'created', time: 'Now', status: 'running' },
          { id: 'evt-model', runId: 'run-a', threadId: 'thread-a', sequence: 2, type: 'model.delta', label: 'Model', detail: 'delta', time: 'Now', status: 'running' },
          { id: 'evt-worker', runId: 'run-a', threadId: 'thread-a', sequence: 3, type: 'job.retrying', label: 'Job', detail: 'retrying', time: 'Now', status: 'retrying' },
          { id: 'evt-error', runId: 'run-a', threadId: 'thread-a', sequence: 4, type: 'stream.error', label: 'Stream', detail: 'stream failed', time: 'Now', status: 'failed' },
        ],
      },
    }))

    expect(html).toContain('Run lifecycle')
    expect(html).toContain('Model stream')
    expect(html).toContain('Worker/job')
    expect(html).toContain('Error')
    expect(html).toContain('retrying')
    expect(html).toContain('stream failed')
  })
})
