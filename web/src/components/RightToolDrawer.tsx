import { FileText } from 'lucide-react'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import type { Message, Run } from '../domain'
import type { Locale } from '../i18n'
import { getRunPreviewArtifacts, type PreviewArtifact } from '../runtime/artifactPreview'
import { getMessagePreviewArtifacts } from '../runtime/messageArtifactPreview'
import { normalizeMarkdownContent } from '../runtime/markdownNormalize'
import { getRightPanelItemCopy, rightPanelItems, type RightPanelItemId } from '../rightPanelItems'

type Props = {
  open: boolean
  selectedPanelId: RightPanelItemId
  selectedArtifactId?: string
  run?: Run | null
  messages?: Message[]
  locale?: Locale
}

function previewCopy(locale: Locale) {
  return locale === 'zh'
    ? {
      emptyTitle: '暂无预览',
      emptyDetail: '生成文档、网页或文件产物后会显示在这里。',
      artifact: '产物',
      source: '来源工具',
    }
    : {
      emptyTitle: 'No preview yet',
      emptyDetail: 'Documents, pages, and file artifacts will appear here.',
      artifact: 'Artifact',
      source: 'Source tool',
    }
}

function ArtifactPreview({ artifact, locale }: { artifact: PreviewArtifact; locale: Locale }) {
  const content = artifact.content || artifact.excerpt || ''
  return (
    <div className="right-panel-artifact">
      <div className="right-panel-artifact-title">
        <span><FileText size={16} /></span>
        <div>
          <strong>{artifact.title}</strong>
          <small>{artifact.filename}</small>
        </div>
      </div>
      {content ? (
        <div className={`right-panel-document kind-${artifact.kind}`}>
          {artifact.kind === 'markdown' ? (
            <ReactMarkdown remarkPlugins={[remarkGfm]}>{normalizeMarkdownContent(content)}</ReactMarkdown>
          ) : (
            <pre>{content}</pre>
          )}
        </div>
      ) : (
        <div className="right-panel-empty compact">
          <strong>{previewCopy(locale).artifact}</strong>
          <p>{artifact.mimeType}</p>
        </div>
      )}
    </div>
  )
}

function PreviewPanel({ run, messages = [], selectedArtifactId, locale }: { run?: Run | null; messages?: Message[]; selectedArtifactId?: string; locale: Locale }) {
  const artifacts = [...getRunPreviewArtifacts(run), ...getMessagePreviewArtifacts(messages)]
  const artifact = artifacts.find((item) => item.id === selectedArtifactId) ?? artifacts.at(-1)
  const copy = previewCopy(locale)
  if (artifact) return <ArtifactPreview artifact={artifact} locale={locale} />

  return (
    <div className="right-panel-empty">
      <span className="right-panel-empty-icon"><FileText size={18} /></span>
      <strong>{copy.emptyTitle}</strong>
      <p>{copy.emptyDetail}</p>
    </div>
  )
}

export function RightToolDrawer({ open, selectedPanelId, selectedArtifactId, run, messages = [], locale = 'en' }: Props) {
  const selectedPanel = rightPanelItems.find((item) => item.id === selectedPanelId) ?? rightPanelItems[0]
  const selectedPanelCopy = getRightPanelItemCopy(selectedPanel, locale)
  const artifacts = [...getRunPreviewArtifacts(run), ...getMessagePreviewArtifacts(messages)]
  const selectedArtifact = artifacts.find((item) => item.id === selectedArtifactId) ?? artifacts.at(-1)

  return (
    <aside className={open ? 'right-tool-drawer open' : 'right-tool-drawer'}>
      <div className="right-panel-head">
        <div>
          <strong>{selectedArtifact?.title ?? selectedPanelCopy.title}</strong>
          <span>{selectedArtifact?.mimeType ?? selectedPanelCopy.description}</span>
        </div>
      </div>
      <PreviewPanel run={run} messages={messages} selectedArtifactId={selectedArtifactId} locale={locale} />
    </aside>
  )
}
