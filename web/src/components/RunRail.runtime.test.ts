import { describe, expect, test } from 'bun:test'
import { createElement } from 'react'
import { renderToStaticMarkup } from 'react-dom/server'
import { RunRail } from './RunRail'

describe('RunRail tool continuation runtime states', () => {
  test('labels initial model tool execution continuation and final phases', () => {
    const html = renderToStaticMarkup(createElement(RunRail, {
      run: {
        id: 'run-a',
        threadId: 'thread-a',
        status: 'completed',
        model: 'Model gateway',
        context: 'model_gateway',
        events: [
          { id: 'evt-initial', sequence: 1, type: 'message.model_output_delta', label: 'message', detail: 'Model output delta', time: 'Now', status: 'running', group: 'model-stream', metadata: { model_phase: 'initial' } },
          { id: 'evt-tool', sequence: 2, type: 'tool.call.succeeded', label: 'tool', detail: 'Tool call succeeded', time: 'Now', status: 'running', group: 'tool-call', metadata: { tool_call_id: 'tc_1', tool_name: 'runtime.get_current_time', result_summary: { timezone: 'UTC' } } },
          { id: 'evt-continuation', sequence: 3, type: 'message.model_output_delta', label: 'message', detail: 'Model output delta', time: 'Now', status: 'running', group: 'model-stream', metadata: { model_phase: 'continuation' } },
          { id: 'evt-final', sequence: 4, type: 'run.completed', label: 'run', detail: 'Run completed', time: 'Now', status: 'completed', group: 'run-lifecycle' },
        ],
      },
      open: true,
      onOpenArtifact: () => {},
    }))

    expect(html).toContain('Initial model phase')
    expect(html).toContain('Tool call succeeded')
    expect(html).toContain('Continuation model phase')
    expect(html).toContain('Run completed')
  })

  test('shows M9 pipeline foundation stage rows', () => {
    const html = renderToStaticMarkup(createElement(RunRail, {
      run: {
        id: 'run-a',
        threadId: 'thread-a',
        status: 'completed',
        model: 'Model gateway',
        context: 'model_gateway',
        events: [
          { id: 'evt-context', sequence: 1, type: 'pipeline.step.completed', label: 'progress', detail: 'Pipeline step completed · step: prepare_context', time: 'Now', status: 'running', group: 'worker-job', metadata: { step: 'prepare_context' } },
          { id: 'evt-tools', sequence: 2, type: 'pipeline.step.completed', label: 'progress', detail: 'Pipeline step completed · step: resolve_tools', time: 'Now', status: 'running', group: 'worker-job', metadata: { step: 'resolve_tools' } },
          { id: 'evt-runtime', sequence: 3, type: 'pipeline.step.completed', label: 'progress', detail: 'Pipeline step completed · step: invoke_runtime', time: 'Now', status: 'running', group: 'worker-job', metadata: { step: 'invoke_runtime' } },
          { id: 'evt-finalize', sequence: 4, type: 'pipeline.step.completed', label: 'progress', detail: 'Pipeline step completed · step: finalize', time: 'Now', status: 'running', group: 'worker-job', metadata: { step: 'finalize' } },
        ],
      },
      open: true,
      onOpenArtifact: () => {},
    }))

    expect(html).toContain('prepare_context')
    expect(html).toContain('resolve_tools')
    expect(html).toContain('invoke_runtime')
    expect(html).toContain('finalize')
  })
})
