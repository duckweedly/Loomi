import { describe, expect, test } from 'bun:test'
import { getRunPreviewArtifacts, getToolCallArtifact } from './artifactPreview'
import type { Run, ToolCall } from '../domain'

function toolCall(overrides: Partial<ToolCall>): ToolCall {
  return {
    id: 'tool-a',
    toolCallId: 'tc-a',
    name: 'artifact.create_text',
    status: 'succeeded',
    summary: 'done',
    input: '',
    output: '',
    resultSummary: null,
    ...overrides,
  }
}

describe('artifactPreview', () => {
  test('extracts first-class artifact metadata from tool results', () => {
    const artifact = getToolCallArtifact(toolCall({
      resultSummary: {
        artifacts: [{
          key: 'run/report.md',
          filename: 'report.md',
          mime_type: 'text/markdown',
          title: 'Report',
          content: '# Report',
        }],
      },
    }))

    expect(artifact).toMatchObject({
      id: 'run/report.md',
      title: 'Report',
      filename: 'report.md',
      kind: 'markdown',
      content: '# Report',
    })
  })

  test('treats artifact.create_text summaries as previewable documents', () => {
    const artifact = getToolCallArtifact(toolCall({
      resultSummary: {
        operation: 'create_text',
        artifact_id: 'art_mock',
        title: 'Notes',
        text_excerpt: 'Safe bounded text',
      },
    }))

    expect(artifact).toMatchObject({
      id: 'art_mock',
      title: 'Notes',
      excerpt: 'Safe bounded text',
    })
  })

  test('collects run artifacts from event result summaries', () => {
    const run: Run = {
      id: 'run-a',
      threadId: 'thread-a',
      status: 'running',
      model: 'Local simulated',
      context: 'local_simulated',
      events: [{
        id: 'evt-artifact',
        type: 'tool.call.succeeded',
        label: 'Tool',
        detail: 'artifact.create_text completed',
        time: 'Now',
        status: 'running',
        metadata: {
          tool_call_id: 'tc-artifact',
          tool_name: 'artifact.create_text',
          result_summary: { artifact_id: 'art-a', title: 'Plan', text_excerpt: '# Plan' },
        },
      }],
    }

    expect(getRunPreviewArtifacts(run)).toHaveLength(1)
    expect(getRunPreviewArtifacts(run)[0]).toMatchObject({ id: 'art-a', title: 'Plan' })
  })

  test('extracts visual artifact content from create_visual results', () => {
    const artifact = getToolCallArtifact(toolCall({
      name: 'artifact.create_visual',
      resultSummary: {
        artifacts: [{
          key: 'art-svg',
          title: 'Flow',
          filename: 'flow.svg',
          mime_type: 'image/svg+xml',
          content: '<svg viewBox="0 0 10 10"></svg>',
        }],
      },
    }))

    expect(artifact).toMatchObject({
      id: 'art-svg',
      title: 'Flow',
      kind: 'svg',
      content: '<svg viewBox="0 0 10 10"></svg>',
    })
  })
})
