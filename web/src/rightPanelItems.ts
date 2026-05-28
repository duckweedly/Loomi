import { Play, type LucideIcon } from 'lucide-react'
import type { Locale } from './i18n'

export type RightPanelItemId = 'preview'

export type RightPanelItem = {
  id: RightPanelItemId
  label: string
  shortcut?: string
  title: string
  description: string
  Icon: LucideIcon
}

export const rightPanelItems: RightPanelItem[] = [
  {
    id: 'preview',
    label: 'Preview',
    title: 'Preview',
    description: 'Browser and artifact previews will appear here.',
    Icon: Play,
  },
]

export function getRightPanelItemCopy(item: RightPanelItem, locale: Locale) {
  if (locale === 'en') return item
  const copy: Record<RightPanelItemId, Pick<RightPanelItem, 'label' | 'title' | 'description'>> = {
    preview: {
      label: '预览',
      title: '预览',
      description: '浏览器和产物预览会显示在这里。',
    },
  }
  return { ...item, ...copy[item.id] }
}
