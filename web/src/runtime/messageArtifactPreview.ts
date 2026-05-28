import type { Message } from '../domain'
import type { PreviewArtifact } from './artifactPreview'
import { normalizeMarkdownContent } from './markdownNormalize'

const fencedMarkdownPattern = /```(?:md|markdown)\s*\n?([\s\S]*?)```/i
const inlineMarkdownPattern = /`(md#{1,6}[\s\S]*?)`/
const artifactLinkPattern = /\[([^\]\n]+)\]\(artifact:([^) \n]+)\)/

function filenameFromMessage(content: string) {
  return content.match(/保存为\s+`([^`]+\.md)`/)?.[1]
    ?? content.match(/保存为\s+([^\s：:]+\.md)/)?.[1]
    ?? 'Markdown.md'
}

function titleFromMarkdown(markdown: string, fallback: string) {
  const heading = markdown.match(/^#{1,6}\s+(.+)$/m)?.[1]?.trim()
  return heading || fallback.replace(/\.md$/i, '')
}

function normalizeCandidate(raw: string) {
  return normalizeMarkdownContent(raw.replace(/^(?:md|markdown)(?=#{1,6})/i, '').trim())
}

function hasSaveIntent(content: string) {
  return /保存为\s+`?[^`\s：:]+\.md`?/i.test(content)
}

export function extractMessageArtifact(message: Pick<Message, 'id' | 'content'>): PreviewArtifact | null {
  const artifactLink = message.content.match(artifactLinkPattern)
  if (artifactLink) {
    const title = artifactLink[1].trim()
    const key = artifactLink[2].trim()
    return {
      id: key,
      title,
      filename: title,
      mimeType: 'text/markdown',
      kind: 'markdown',
    }
  }

  const fenced = message.content.match(fencedMarkdownPattern)
  const inline = fenced ? null : message.content.match(inlineMarkdownPattern)
  const raw = fenced?.[1] ?? inline?.[1]
  if (!raw) return null
  if (!hasSaveIntent(message.content)) return null

  const content = normalizeCandidate(raw)
  const filename = filenameFromMessage(message.content)
  return {
    id: `message:${message.id}:markdown`,
    title: titleFromMarkdown(content, filename),
    filename,
    mimeType: 'text/markdown',
    kind: 'markdown',
    content,
  }
}

export function stripMessageArtifactSource(content: string) {
  return content
    .replace(artifactLinkPattern, '$1')
    .replace(fencedMarkdownPattern, '')
    .replace(inlineMarkdownPattern, '')
    .replace(/[ \t]+\n/g, '\n')
    .replace(/\n{3,}/g, '\n\n')
    .trim()
}

export function getMessagePreviewArtifacts(messages: Pick<Message, 'id' | 'content'>[]) {
  return messages
    .map(extractMessageArtifact)
    .filter((artifact): artifact is PreviewArtifact => artifact !== null)
}
