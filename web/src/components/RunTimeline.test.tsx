import { describe, expect, test } from 'bun:test'
import { createElement } from 'react'
import { renderToStaticMarkup } from 'react-dom/server'
import { RunTimeline } from './RunTimeline'

describe('RunTimeline M7 tool grouping', () => {
  test('keeps approval-required tool events separate from model stream rows', () => {
    const html = renderToStaticMarkup(createElement(RunTimeline, {
      runDetailsOpen: true,
      rightToolsOpen: false,
      selectedPanelId: 'preview',
      run: {
        id: 'run-a',
        threadId: 'thread-a',
        status: 'blocked_on_tool_approval',
        model: 'Model gateway',
        context: 'model_gateway',
        source: 'model_gateway',
        events: [
          { id: 'evt-model', runId: 'run-a', threadId: 'thread-a', sequence: 1, type: 'model.delta', label: 'Model', detail: 'draft', time: 'Now', status: 'running', group: 'model-stream' },
          { id: 'evt-tool', runId: 'run-a', threadId: 'thread-a', sequence: 2, type: 'tool.call.approval_required', label: 'Tool', detail: 'Tool approval required', time: 'Now', status: 'blocked_on_tool_approval', group: 'tool-call' },
        ],
      },
    }))

    expect(html).toContain('Preview')
    expect(html).not.toContain('Tool call waiting for approval')
    expect(html).not.toContain('Model stream')
    expect(html).not.toContain('draft')
  })
})
