import { ExternalLink, FileText } from 'lucide-react'
import { useState } from 'react'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import type { Message, Run } from '../domain'
import type { Locale } from '../i18n'
import { getRunPreviewArtifacts, type PreviewArtifact } from '../runtime/artifactPreview'
import { getMessagePreviewArtifacts } from '../runtime/messageArtifactPreview'
import { normalizeMarkdownContent } from '../runtime/markdownNormalize'
import { canOpenArtifactNatively, nativeArtifactOpenLabel, openArtifactNatively, type NativeArtifactOpenStatus } from '../runtime/nativeArtifactOpen'
import { getRightPanelItemCopy, rightPanelItems, type RightPanelItemId } from '../rightPanelItems'
import { ArtifactFrame } from './ArtifactFrame'

type Props = {
  open: boolean
  selectedPanelId: RightPanelItemId
  selectedArtifactId?: string
  run?: Run | null
  messages?: Message[]
  artifacts?: PreviewArtifact[]
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
  const [nativeOpenStatus, setNativeOpenStatus] = useState<NativeArtifactOpenStatus>('idle')
  const canOpenNative = canOpenArtifactNatively(artifact)
  const openNative = async () => {
    if (!canOpenNative) return
    setNativeOpenStatus('opening')
    setNativeOpenStatus(await openArtifactNatively(artifact) ? 'idle' : 'failed')
  }
  return (
    <div className="right-panel-artifact">
      <div className="right-panel-artifact-title">
        <span><FileText size={16} /></span>
        <div>
          <strong>{artifact.title}</strong>
          <small>{artifact.filename}</small>
        </div>
        <button type="button" className="artifact-native-open" disabled={!canOpenNative || nativeOpenStatus === 'opening'} onClick={openNative}>
          <ExternalLink size={14} />
          <span>{nativeArtifactOpenLabel(locale, nativeOpenStatus)}</span>
        </button>
      </div>
      {content ? (
        <div className={`right-panel-document kind-${artifact.kind}`}>
          {artifact.kind === 'markdown' ? (
            <ReactMarkdown remarkPlugins={[remarkGfm]}>{normalizeMarkdownContent(content)}</ReactMarkdown>
          ) : artifact.kind === 'svg' || artifact.kind === 'html' ? (
            <ArtifactFrame artifact={artifact} />
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

function mergeArtifacts(...groups: PreviewArtifact[][]) {
  const byId = new Map<string, PreviewArtifact>()
  for (const artifact of groups.flat()) {
    const existing = byId.get(artifact.id)
    byId.set(artifact.id, existing && !existing.content && artifact.content ? artifact : existing ?? artifact)
  }
  return [...byId.values()]
}

function PreviewPanel({ run, messages = [], selectedArtifactId, locale, artifacts: threadArtifacts = [] }: { run?: Run | null; messages?: Message[]; selectedArtifactId?: string; locale: Locale; artifacts?: PreviewArtifact[] }) {
  const runArtifacts = getRunPreviewArtifacts(run)
  const artifacts = mergeArtifacts(threadArtifacts, runArtifacts, getMessagePreviewArtifacts(messages, [...threadArtifacts, ...runArtifacts]))
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

export function RightToolDrawer({ open, selectedPanelId, selectedArtifactId, run, messages = [], artifacts: threadArtifacts = [], locale = 'en' }: Props) {
  const selectedPanel = rightPanelItems.find((item) => item.id === selectedPanelId) ?? rightPanelItems[0]
  const selectedPanelCopy = getRightPanelItemCopy(selectedPanel, locale)
  const runArtifacts = getRunPreviewArtifacts(run)
  const artifacts = mergeArtifacts(threadArtifacts, runArtifacts, getMessagePreviewArtifacts(messages, [...threadArtifacts, ...runArtifacts]))
  const selectedArtifact = artifacts.find((item) => item.id === selectedArtifactId) ?? artifacts.at(-1)

  return (
    <aside className={open ? 'right-tool-drawer open' : 'right-tool-drawer'}>
      <div className="right-panel-head">
        <div>
          <strong>{selectedArtifact?.title ?? selectedPanelCopy.title}</strong>
          <span>{selectedArtifact?.mimeType ?? selectedPanelCopy.description}</span>
        </div>
      </div>
      <PreviewPanel run={run} messages={messages} selectedArtifactId={selectedArtifactId} locale={locale} artifacts={threadArtifacts} />
    </aside>
  )
}
