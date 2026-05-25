import { describe, expect, test } from 'bun:test'
import { createElement } from 'react'
import { renderToStaticMarkup } from 'react-dom/server'
import { RunTimeline } from './RunTimeline'

describe('RunTimeline M7 tool grouping', () => {
  test('keeps approval-required tool events separate from model stream rows', () => {
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

    expect(html).toContain('Model stream')
    expect(html).toContain('Tool call')
    expect(html).toContain('draft')
    expect(html).toContain('Tool approval required')
    expect(html.indexOf('Model stream')).toBeLessThan(html.indexOf('Tool call'))
  })
})
