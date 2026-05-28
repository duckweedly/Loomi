import { describe, expect, test } from 'bun:test'
import { createElement } from 'react'
import { renderToStaticMarkup } from 'react-dom/server'
import { RightToolDrawer } from './RightToolDrawer'

describe('RightToolDrawer preview panel', () => {
  test('renders only the preview drawer surface', () => {
    const html = renderToStaticMarkup(createElement(RightToolDrawer, {
      open: true,
      selectedPanelId: 'preview',
      run: {
        id: 'run-a',
        threadId: 'thread-a',
        status: 'recovering',
        model: 'Local simulated',
        context: 'local_simulated',
        events: [{ id: 'evt-worker', runId: 'run-a', threadId: 'thread-a', type: 'job_recovering', label: 'Job', detail: 'attempt 2 of 3', time: '10:01', status: 'recovering' }],
      },
    }))

    expect(html).toContain('Preview')
    expect(html).toContain('No preview yet')
    expect(html).not.toContain('Background tasks')
    expect(html).not.toContain('No background task is running')
    expect(html).not.toContain('attempt 2 of 3')
  })

  test('renders Chinese preview chrome', () => {
    const html = renderToStaticMarkup(createElement(RightToolDrawer, {
      open: true,
      selectedPanelId: 'preview',
      locale: 'zh',
    }))

    expect(html).toContain('预览')
    expect(html).toContain('暂无预览')
  })

  test('renders the latest Markdown artifact in preview', () => {
    const html = renderToStaticMarkup(createElement(RightToolDrawer, {
      open: true,
      selectedPanelId: 'preview',
      selectedArtifactId: 'art_mock',
      locale: 'zh',
      run: {
        id: 'run-a',
        threadId: 'thread-a',
        status: 'running',
        model: 'Local simulated',
        context: 'local_simulated',
        events: [{
          id: 'evt-artifact',
          runId: 'run-a',
          threadId: 'thread-a',
          type: 'tool.call.succeeded',
          label: 'Tool',
          detail: 'artifact.create_text completed',
          time: 'Now',
          status: 'running',
          metadata: {
            tool_call_id: 'tc_artifact',
            tool_name: 'artifact.create_text',
            result_summary: { artifact_id: 'art_mock', title: '三句话的 Markdown', filename: '三句话.md', mime_type: 'text/markdown', text_excerpt: '# 三句话的 Markdown' },
          },
        }],
      },
    }))

    expect(html).toContain('三句话的 Markdown')
    expect(html).toContain('<h1>')
    expect(html).not.toContain('暂无预览')
  })

  test('previews Markdown artifacts extracted from assistant messages', () => {
    const html = renderToStaticMarkup(createElement(RightToolDrawer, {
      open: true,
      selectedPanelId: 'preview',
      selectedArtifactId: 'message:msg-a:markdown',
      locale: 'zh',
      messages: [{
        id: 'msg-a',
        threadId: 'thread-a',
        role: 'assistant',
        content: '把下面内容保存为 `三句话.md`：\n\n```md\n# 三句话的 Markdown\n\n今天我开始写一个简单的 Markdown 文档。\n```',
        createdAt: 'Now',
      }],
    }))

    expect(html).toContain('三句话的 Markdown')
    expect(html).toContain('今天我开始写一个简单的 Markdown 文档。')
    expect(html).toContain('<h1>')
    expect(html).not.toContain('暂无预览')
  })
})
