import { CheckSquare, Code2, Files, GitCompare, Play, SquareTerminal, type LucideIcon } from 'lucide-react'
import type { Locale } from './i18n'

export type RightPanelItemId = 'preview' | 'diff' | 'terminal' | 'files' | 'background-tasks' | 'plan'

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
    shortcut: '⇧ ⌘ P',
    title: 'Preview',
    description: 'Browser and artifact previews will appear here.',
    Icon: Play,
  },
  {
    id: 'diff',
    label: 'Diff',
    shortcut: '⇧ ⌘ D',
    title: 'Diff',
    description: 'Code changes and review surfaces will appear here.',
    Icon: GitCompare,
  },
  {
    id: 'terminal',
    label: 'Terminal',
    shortcut: '^ `',
    title: 'Terminal',
    description: 'Shell sessions will appear here once the runtime is wired.',
    Icon: SquareTerminal,
  },
  {
    id: 'files',
    label: 'Files',
    shortcut: '⇧ ⌘ F',
    title: 'Files',
    description: 'Workspace file navigation will appear here.',
    Icon: Files,
  },
  {
    id: 'background-tasks',
    label: 'Background tasks',
    title: 'Background tasks',
    description: 'Long-running jobs and agent tasks will appear here.',
    Icon: Code2,
  },
  {
    id: 'plan',
    label: 'Plan',
    title: 'Plan',
    description: 'Planning state and step progress will appear here.',
    Icon: CheckSquare,
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
    diff: {
      label: '变更',
      title: '变更',
      description: '代码变更和审阅界面会显示在这里。',
    },
    terminal: {
      label: '终端',
      title: '终端',
      description: '运行时接好后，Shell 会话会显示在这里。',
    },
    files: {
      label: '文件',
      title: '文件',
      description: '工作区文件导航会显示在这里。',
    },
    'background-tasks': {
      label: '后台任务',
      title: '后台任务',
      description: '长任务和 Agent 任务会显示在这里。',
    },
    plan: {
      label: '计划',
      title: '计划',
      description: '计划状态和步骤进度会显示在这里。',
    },
  }
  return { ...item, ...copy[item.id] }
}
