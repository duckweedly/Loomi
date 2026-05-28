import type { Run, ToolCall } from '../domain'

export type PreviewArtifact = {
  id: string
  title: string
  filename: string
  mimeType: string
  kind: 'markdown' | 'text' | 'html' | 'image' | 'unknown'
  content?: string
  excerpt?: string
  sourceToolCallId?: string
}

function record(value: unknown): Record<string, unknown> | null {
  return typeof value === 'object' && value !== null && !Array.isArray(value) ? value as Record<string, unknown> : null
}

function stringValue(value: unknown): string | undefined {
  return typeof value === 'string' && value.trim() ? value.trim() : undefined
}

function mimeFromFilename(filename: string) {
  const lower = filename.toLowerCase()
  if (lower.endsWith('.md') || lower.endsWith('.markdown')) return 'text/markdown'
  if (lower.endsWith('.html') || lower.endsWith('.htm')) return 'text/html'
  if (lower.endsWith('.png')) return 'image/png'
  if (lower.endsWith('.jpg') || lower.endsWith('.jpeg')) return 'image/jpeg'
  if (lower.endsWith('.svg')) return 'image/svg+xml'
  if (lower.endsWith('.txt')) return 'text/plain'
  return 'application/octet-stream'
}

function kindFromMime(mimeType: string): PreviewArtifact['kind'] {
  if (mimeType === 'text/markdown') return 'markdown'
  if (mimeType.startsWith('text/html')) return 'html'
  if (mimeType.startsWith('image/')) return 'image'
  if (mimeType.startsWith('text/')) return 'text'
  return 'unknown'
}

function artifactFromRecord(source: Record<string, unknown>, toolCall: ToolCall): PreviewArtifact | null {
  const filename = stringValue(source.filename) ?? stringValue(source.path) ?? stringValue(source.title) ?? ''
  const title = stringValue(source.title) ?? filename
  const id = stringValue(source.key) ?? stringValue(source.artifact_id) ?? stringValue(source.id) ?? toolCall.toolCallId ?? toolCall.id
  if (!title && !filename) return null

  const mimeType = stringValue(source.mime_type) ?? stringValue(source.mimeType) ?? mimeFromFilename(filename || title)
  return {
    id,
    title: title || filename || id,
    filename: filename || title || id,
    mimeType,
    kind: kindFromMime(mimeType),
    content: stringValue(source.content) ?? stringValue(source.markdown),
    excerpt: stringValue(source.text_excerpt) ?? stringValue(source.summary) ?? stringValue(source.preview),
    sourceToolCallId: toolCall.toolCallId ?? toolCall.id,
  }
}

export function getToolCallArtifact(toolCall: ToolCall): PreviewArtifact | null {
  const result = record(toolCall.resultSummary)
  if (!result) return null

  const nested = Array.isArray(result.artifacts) ? result.artifacts.map(record).find(Boolean) : null
  if (nested) return artifactFromRecord(nested, toolCall)

  if (toolCall.name === 'artifact.create_text' || toolCall.name === 'document_write' || toolCall.name === 'create_artifact') {
    return artifactFromRecord(result, toolCall)
  }

  if (toolCall.name === 'workspace.write_file') {
    const filename = stringValue(result.filename) ?? stringValue(result.path)
    const mimeType = filename ? mimeFromFilename(filename) : ''
    if (mimeType === 'text/markdown' || mimeType === 'text/plain') return artifactFromRecord(result, toolCall)
  }

  return null
}

export function getRunPreviewArtifacts(run: Run | null | undefined): PreviewArtifact[] {
  if (!run?.toolCalls?.length && !run?.events.length) return []
  const byId = new Map<string, PreviewArtifact>()

  for (const toolCall of run.toolCalls ?? []) {
    const artifact = getToolCallArtifact(toolCall)
    if (artifact) byId.set(artifact.id, artifact)
  }

  for (const event of run.events) {
    if (!event.type.startsWith('tool.call.')) continue
    const result = record(event.metadata?.result_summary)
    if (!result) continue
    const toolCall: ToolCall = {
      id: event.id,
      toolCallId: stringValue(event.metadata?.tool_call_id),
      name: stringValue(event.metadata?.tool_name) ?? event.label,
      status: event.type.endsWith('.succeeded') ? 'succeeded' : 'running',
      summary: event.detail,
      input: '',
      output: event.content ?? '',
      resultSummary: result,
    }
    const artifact = getToolCallArtifact(toolCall)
    if (artifact) byId.set(artifact.id, artifact)
  }

  return [...byId.values()]
}
