import type { Locale } from '../i18n'
import type { PreviewArtifact } from './artifactPreview'

export type NativeArtifactOpenStatus = 'idle' | 'opening' | 'failed'

export function nativeArtifactOpenLabel(locale: Locale, status: NativeArtifactOpenStatus) {
  if (locale === 'zh') {
    if (status === 'opening') return '正在打开'
    if (status === 'failed') return '无法打开'
    return '本机打开'
  }
  if (status === 'opening') return 'Opening'
  if (status === 'failed') return 'Could not open'
  return 'Open in app'
}

export function canOpenArtifactNatively(artifact: PreviewArtifact) {
  const content = artifact.content || artifact.excerpt || ''
  return Boolean(content && typeof window !== 'undefined' && window.loomiDesktop?.openArtifactFile)
}

export async function openArtifactNatively(artifact: PreviewArtifact) {
  const content = artifact.content || artifact.excerpt || ''
  if (!content || typeof window === 'undefined' || !window.loomiDesktop?.openArtifactFile) return false
  const result = await window.loomiDesktop.openArtifactFile({
    title: artifact.title,
    filename: artifact.filename,
    mimeType: artifact.mimeType,
    content,
  })
  return result.ok
}
