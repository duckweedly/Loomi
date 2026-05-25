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

  test('renders safe worker diagnostic metadata without secret-looking fields', () => {
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
        status: 'recovering',
        model: 'Local simulated',
        context: 'local_simulated',
        events: [
          { id: 'evt-worker', runId: 'run-a', threadId: 'thread-a', sequence: 1, type: 'job.recovering', label: 'Worker', detail: 'Job recovering · stale_count: 1 · dead_count: 0', time: 'Now', status: 'recovering' },
        ],
      },
    }))

    expect(html).toContain('Worker/job')
    expect(html).toContain('stale_count: 1')
    expect(html).not.toContain('password')
    expect(html).not.toContain('token')
  })

  test('does not show a previous thread run in Background tasks after thread selection changes', () => {
    const html = renderToStaticMarkup(createElement(RunTimeline, {
      runDetailsOpen: false,
      rightPanelMenuOpen: false,
      rightToolsOpen: true,
      selectedPanelId: 'background-tasks',
      selectedThreadId: 'thread-b',
      onSelectPanel: () => {},
      onOpenArtifact: () => {},
      run: {
        id: 'run-a',
        threadId: 'thread-a',
        status: 'recovering',
        model: 'Local simulated',
        context: 'local_simulated',
        events: [
          { id: 'evt-worker', runId: 'run-a', threadId: 'thread-a', sequence: 1, type: 'job_recovering', label: 'Worker', detail: 'previous thread job', time: 'Now', status: 'recovering' },
        ],
      },
    }))

    expect(html).toContain('No background task is running')
    expect(html).not.toContain('Current run job')
    expect(html).not.toContain('background-task-event')
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
          { id: 'evt-pipeline', runId: 'run-a', threadId: 'thread-a', sequence: 3.5, type: 'pipeline.step.started', label: 'Pipeline', detail: 'invoke runtime', time: 'Now', status: 'running' },
          { id: 'evt-error', runId: 'run-a', threadId: 'thread-a', sequence: 4, type: 'stream.error', label: 'Stream', detail: 'stream failed', time: 'Now', status: 'failed' },
        ],
      },
    }))

    expect(html).toContain('Run lifecycle')
    expect(html).toContain('Model stream')
    expect(html).toContain('Worker/job')
    expect(html).toContain('Error')
    expect(html).toContain('retrying')
    expect(html).toContain('invoke runtime')
    expect(html).toContain('stream failed')
  })

  test('renders M9 pipeline foundation stage trace safely', () => {
    const html = renderToStaticMarkup(createElement(RunTimeline, {
      runDetailsOpen: true,
      rightPanelMenuOpen: false,
      rightToolsOpen: true,
      selectedPanelId: 'background-tasks',
      onSelectPanel: () => {},
      onOpenArtifact: () => {},
      run: {
        id: 'run-a',
        threadId: 'thread-a',
        status: 'completed',
        model: 'Model gateway',
        context: 'model_gateway',
        events: [
          { id: 'evt-context', runId: 'run-a', threadId: 'thread-a', sequence: 1, type: 'pipeline.step.completed', label: 'progress', detail: 'Pipeline step completed · step: prepare_context · message_count: 1', time: 'Now', status: 'running', group: 'worker-job', metadata: { step: 'prepare_context', message_count: 1 } },
          { id: 'evt-tools', runId: 'run-a', threadId: 'thread-a', sequence: 2, type: 'pipeline.step.completed', label: 'progress', detail: 'Pipeline step completed · step: resolve_tools · enabled_tool_count: 1', time: 'Now', status: 'running', group: 'worker-job', metadata: { step: 'resolve_tools', enabled_tool_count: 1 } },
          { id: 'evt-runtime', runId: 'run-a', threadId: 'thread-a', sequence: 3, type: 'pipeline.step.completed', label: 'progress', detail: 'Pipeline step completed · step: invoke_runtime', time: 'Now', status: 'running', group: 'worker-job', metadata: { step: 'invoke_runtime' } },
          { id: 'evt-finalize', runId: 'run-a', threadId: 'thread-a', sequence: 4, type: 'pipeline.step.completed', label: 'progress', detail: 'Pipeline step completed · step: finalize', time: 'Now', status: 'running', group: 'worker-job', metadata: { step: 'finalize' } },
        ],
      },
    }))

    expect(html).toContain('prepare_context')
    expect(html).toContain('resolve_tools')
    expect(html).toContain('invoke_runtime')
    expect(html).toContain('finalize')
    expect(html).not.toContain('api_key')
    expect(html).not.toContain('secret')
  })

  test('renders M11 MCP discovery labels without sensitive config data', () => {
    const html = renderToStaticMarkup(createElement(RunTimeline, {
      runDetailsOpen: true,
      rightPanelMenuOpen: false,
      rightToolsOpen: true,
      selectedPanelId: 'background-tasks',
      onSelectPanel: () => {},
      onOpenArtifact: () => {},
      run: {
        id: 'run-a',
        threadId: 'thread-a',
        status: 'completed',
        model: 'Model gateway',
        context: 'model_gateway',
        events: [
          { id: 'evt-mcp', runId: 'run-a', threadId: 'thread-a', sequence: 1, type: 'mcp.discovery.succeeded', label: 'progress', detail: 'MCP discovery succeeded · mcp_candidate_count: 1 · mcp_non_executable_candidate_names: mcp.local-search.search', time: 'Now', status: 'running', group: 'worker-job', metadata: { mcp_candidate_count: 1, mcp_non_executable_candidate_names: ['mcp.local-search.search'], mcp_execution_enabled: false } },
          { id: 'evt-mcp-disabled', runId: 'run-a', threadId: 'thread-a', sequence: 2, type: 'mcp.tools.non_executable', label: 'progress', detail: 'MCP execution disabled', time: 'Now', status: 'running', group: 'worker-job', metadata: { mcp_execution_enabled: false } },
        ],
      },
    }))

    expect(html).toContain('MCP discovery succeeded')
    expect(html).toContain('mcp.local-search.search')
    expect(html).toContain('MCP execution disabled')
    expect(html).not.toContain('sk-live')
    expect(html).not.toContain('command_path')
    expect(html).not.toContain('stderr')
  })

  test('renders two model phases around tool result continuation', () => {
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
        status: 'completed',
        model: 'Model gateway',
        context: 'model_gateway',
        events: [
          { id: 'evt-initial', runId: 'run-a', threadId: 'thread-a', sequence: 1, type: 'message.model_output_delta', label: 'message', detail: 'Model output delta', time: 'Now', status: 'running', group: 'model-stream', metadata: { model_phase: 'initial' } },
          { id: 'evt-tool', runId: 'run-a', threadId: 'thread-a', sequence: 2, type: 'tool.call.succeeded', label: 'tool', detail: 'Tool call succeeded', time: 'Now', status: 'running', group: 'tool-call', metadata: { tool_call_id: 'tc_1', tool_name: 'runtime.get_current_time' } },
          { id: 'evt-continuation', runId: 'run-a', threadId: 'thread-a', sequence: 3, type: 'message.model_output_delta', label: 'message', detail: 'Model output delta', time: 'Now', status: 'running', group: 'model-stream', metadata: { model_phase: 'continuation' } },
          { id: 'evt-final', runId: 'run-a', threadId: 'thread-a', sequence: 4, type: 'run.completed', label: 'run', detail: 'Run completed', time: 'Now', status: 'completed', group: 'run-lifecycle' },
        ],
      },
    }))

    expect(html).toContain('Initial model phase')
    expect(html).toContain('Tool call succeeded')
    expect(html).toContain('Continuation model phase')
    expect(html).toContain('Run completed')
  })
})
