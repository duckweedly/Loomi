import type { Message } from '../domain'
import type { PreviewArtifact } from './artifactPreview'
import { normalizeMarkdownContent } from './markdownNormalize'

const fencedMarkdownPattern = /```(?:md|markdown)\s*\n?([\s\S]*?)```/i
const fencedSvgPattern = /```(?:svg|xml)\s*\n?([\s\S]*?<\/svg>)\s*```/i
const rawSvgPattern = /(<svg\b[\s\S]*?<\/svg>)/i
const inlineMarkdownPattern = /`\s*((?:md|markdown)\s*#{1,6}[\s\S]*?)\s*`/i
const looseInlineMarkdownPattern = /([：:\s])((?:md|markdown)\s*#{1,6}[\s\S]*?)(?=(?:如果你要|如果要|告诉我|$))/i
const looseDocumentMarkdownPattern = /^(?:md|markdown)\s*#{1,6}[\s\S]*$/i
const artifactLinkPattern = /\[([^\]\n]+)\]\(artifact:([^) \n]+)\)/

function filenameFromMessage(content: string) {
  return content.match(/保存为\s+`([^`]+\.md)`/)?.[1]
    ?? content.match(/保存为\s+([^\s：:]+\.md)/)?.[1]
    ?? 'Markdown.md'
}

function titleFromMarkdown(markdown: string, fallback: string) {
  const heading = markdown.match(/^#{1,6}\s+(.+)$/m)?.[1]?.trim()
  const compactTitle = heading?.replace(/(?:一句话|简短描述|概述|简介|这里).*$/u, '').trim()
  return compactTitle || heading || fallback.replace(/\.md$/i, '')
}

function titleFromSvg(svg: string, fallback: string) {
  const title = svg.match(/<title[^>]*>([^<]+)<\/title>/i)?.[1]?.trim()
  return title || fallback.replace(/\.svg$/i, '')
}

function normalizeCandidate(raw: string) {
  return normalizeMarkdownContent(raw.replace(/^(?:md|markdown)\s*(?=#{1,6})/i, '').trim())
}

function hasSaveIntent(content: string) {
  return /保存为\s+`?[^`\s：:]+\.md`?/i.test(content)
}

function hasInlineMarkdownDocumentIntent(content: string) {
  return /Markdown\s*文件内容|Markdown\s*文档内容|一个\s*Markdown\s*文件|一个\s*Markdown\s*文档/i.test(content)
}

function hasDocumentMarkdownShape(content: string) {
  const normalized = normalizeCandidate(content)
  const headings = normalized.match(/^#{1,6}\s+\S/gm)?.length ?? 0
  const lists = normalized.match(/^\s*(?:[-*]|\d+[.、])\s+\S/gm)?.length ?? 0
  return headings >= 3 || (headings >= 2 && lists >= 2)
}

export function extractMessageArtifact(message: Pick<Message, 'id' | 'content'>, knownArtifacts: PreviewArtifact[] = []): PreviewArtifact | null {
  const artifactLink = message.content.match(artifactLinkPattern)
  if (artifactLink) {
    const title = artifactLink[1].trim()
    const key = artifactLink[2].trim()
    const known = knownArtifacts.find((artifact) => artifact.id === key)
    if (known) return { ...known, title: known.title || title }
    return {
      id: key,
      title,
      filename: title,
      mimeType: 'text/markdown',
      kind: 'markdown',
    }
  }

  const fencedSvg = message.content.match(fencedSvgPattern)
  const rawSvg = fencedSvg ? null : message.content.match(rawSvgPattern)
  const svg = fencedSvg?.[1] ?? rawSvg?.[1]
  if (svg) {
    return {
      id: `message:${message.id}:svg`,
      title: titleFromSvg(svg, 'SVG 图'),
      filename: 'visual.svg',
      mimeType: 'image/svg+xml',
      kind: 'svg',
      content: svg.trim(),
    }
  }

  const fenced = message.content.match(fencedMarkdownPattern)
  const inline = fenced ? null : message.content.match(inlineMarkdownPattern)
  const looseInline = fenced || inline || !hasInlineMarkdownDocumentIntent(message.content) ? null : message.content.match(looseInlineMarkdownPattern)
  const looseDocument = fenced || inline || looseInline || hasSaveIntent(message.content) || hasInlineMarkdownDocumentIntent(message.content) ? null : message.content.trim().match(looseDocumentMarkdownPattern)
  const raw = fenced?.[1] ?? inline?.[1] ?? looseInline?.[2] ?? looseDocument?.[0]
  if (!raw) return null
  if (!hasSaveIntent(message.content) && !hasInlineMarkdownDocumentIntent(message.content) && !hasDocumentMarkdownShape(raw)) return null

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
  const stripped = content
    .replace(artifactLinkPattern, '$1')
    .replace(fencedSvgPattern, '')
    .replace(rawSvgPattern, '')
    .replace(fencedMarkdownPattern, '')
    .replace(inlineMarkdownPattern, '')
    .replace(looseInlineMarkdownPattern, '$1')
    .replace(/[ \t]+\n/g, '\n')
    .replace(/\n{3,}/g, '\n\n')
    .trim()
  if (looseDocumentMarkdownPattern.test(stripped) && hasDocumentMarkdownShape(stripped)) return ''
  return stripped
}

export function getMessagePreviewArtifacts(messages: Pick<Message, 'id' | 'content'>[], knownArtifacts: PreviewArtifact[] = []) {
  return messages
    .map((message) => extractMessageArtifact(message, knownArtifacts))
    .filter((artifact): artifact is PreviewArtifact => artifact !== null)
}
