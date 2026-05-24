import type { Locale } from '../i18n'

export type SettingsCategoryGroup = 'primary' | 'agent_core' | 'management'
export type SettingsCategoryStatus = 'working' | 'mock' | 'preview' | 'read_only' | 'mixed'
export type SettingRowControlType = 'toggle' | 'select' | 'button' | 'status' | 'placeholder' | 'segmented'
export type SettingRowStatus = 'working' | 'mock' | 'disabled' | 'read_only'

export type SettingsCategoryId =
  | 'general'
  | 'appearance'
  | 'providers'
  | 'connectors'
  | 'plugins'
  | 'skill'
  | 'mcp'
  | 'notebook'
  | 'memory'
  | 'activity-recorder'
  | 'context'
  | 'safety'
  | 'tools'
  | 'routes'
  | 'about'
  | 'advanced'

export type SettingsCategory = {
  id: SettingsCategoryId
  label: string
  group: SettingsCategoryGroup
  status: SettingsCategoryStatus
  description: string
}

export type SettingRow = {
  id: string
  label: string
  helperText: string
  controlType: SettingRowControlType
  status: SettingRowStatus
  value?: string | boolean | null
}

export type SettingSection = {
  id: string
  title: string
  description?: string
  categoryId: SettingsCategoryId
  rows: SettingRow[]
}

export const settingsCategoryGroups: Record<SettingsCategoryGroup, Record<Locale, string>> = {
  primary: { zh: '工作区', en: 'Workspace' },
  agent_core: { zh: 'Agent Core', en: 'Agent Core' },
  management: { zh: '管理', en: 'Management' },
}

const categoryCopy: Record<SettingsCategoryId, Record<Locale, Pick<SettingsCategory, 'label' | 'description'>>> = {
  general: { zh: { label: '通用', description: '当前会话的工作区默认值和运行状态可见性。' }, en: { label: 'General', description: 'Current-session workspace defaults and runtime visibility.' } },
  appearance: { zh: { label: '外观', description: '后续里程碑的主题和显示预览。' }, en: { label: 'Appearance', description: 'Theme and display previews for a later milestone.' } },
  providers: { zh: { label: 'Providers', description: '显示已脱敏模型 provider 能力，provider 管理暂缓。' }, en: { label: 'Providers', description: 'Redacted model-provider capability with provider management deferred.' } },
  connectors: { zh: { label: '连接器', description: '未来外部连接设置的预览区域。' }, en: { label: 'Connectors', description: 'Preview area for future external connection settings.' } },
  plugins: { zh: { label: '插件', description: '未来插件控制的预览区域。' }, en: { label: 'Plugins', description: 'Preview area for future plugin controls.' } },
  skill: { zh: { label: '技能', description: '未来技能控制的预览区域。' }, en: { label: 'Skill', description: 'Preview area for future skill controls.' } },
  mcp: { zh: { label: 'MCP', description: '未来 MCP server 控制的预览区域。' }, en: { label: 'MCP', description: 'Preview area for future MCP server controls.' } },
  notebook: { zh: { label: 'Notebook', description: '未来 notebook 设置的预览区域。' }, en: { label: 'Notebook', description: 'Preview area for future notebook settings.' } },
  memory: { zh: { label: 'Memory', description: '未来记忆和检索设置的预览区域。' }, en: { label: 'Memory', description: 'Preview area for future memory and retrieval settings.' } },
  'activity-recorder': { zh: { label: '活动记录', description: '未来桌面活动记录 opt-in 的预览区域。' }, en: { label: 'Activity Recorder', description: 'Preview area for future opt-in desktop activity capture.' } },
  context: { zh: { label: '上下文', description: '未来上下文组装控制的预览区域。' }, en: { label: 'Context', description: 'Preview area for future context assembly controls.' } },
  safety: { zh: { label: '安全', description: '未来权限和审计控制的预览区域。' }, en: { label: 'Safety', description: 'Preview area for future permission and audit controls.' } },
  tools: { zh: { label: '工具', description: '未来工具权限设置的预览区域。' }, en: { label: 'Tools', description: 'Preview area for future tool permission settings.' } },
  routes: { zh: { label: '路由', description: '未来模型和 worker 路由控制的预览区域。' }, en: { label: 'Routes', description: 'Preview area for future model and worker routing controls.' } },
  about: { zh: { label: '关于', description: '已知本地应用状态和占位构建信息。' }, en: { label: 'About', description: 'Known local app state plus placeholder build metadata.' } },
  advanced: { zh: { label: '高级', description: '未来诊断和底层选项的预览区域。' }, en: { label: 'Advanced', description: 'Preview area for future diagnostics and low-level options.' } },
}

const categoryStructure: Omit<SettingsCategory, 'label' | 'description'>[] = [
  { id: 'general', group: 'primary', status: 'working' },
  { id: 'appearance', group: 'primary', status: 'mock' },
  { id: 'providers', group: 'agent_core', status: 'mixed' },
  { id: 'connectors', group: 'agent_core', status: 'mock' },
  { id: 'plugins', group: 'agent_core', status: 'mock' },
  { id: 'skill', group: 'agent_core', status: 'mock' },
  { id: 'mcp', group: 'agent_core', status: 'mock' },
  { id: 'notebook', group: 'agent_core', status: 'mock' },
  { id: 'memory', group: 'agent_core', status: 'mock' },
  { id: 'activity-recorder', group: 'agent_core', status: 'mock' },
  { id: 'context', group: 'agent_core', status: 'mock' },
  { id: 'safety', group: 'management', status: 'mock' },
  { id: 'tools', group: 'management', status: 'mock' },
  { id: 'routes', group: 'management', status: 'mock' },
  { id: 'about', group: 'management', status: 'mixed' },
  { id: 'advanced', group: 'management', status: 'mock' },
]

export const settingsCategories: SettingsCategory[] = categoryStructure.map((category) => ({ ...category, ...categoryCopy[category.id].en }))

export function getLocalizedSettingsCategories(locale: Locale): SettingsCategory[] {
  return categoryStructure.map((category) => ({ ...category, ...categoryCopy[category.id][locale] }))
}

export const generalSettingSections: SettingSection[] = [
  {
    id: 'workspace-defaults',
    title: 'Workspace defaults',
    description: 'Session-local preferences for future local workspace actions.',
    categoryId: 'general',
    rows: [
      { id: 'default-workspace-mode', label: 'Default workspace mode', helperText: 'Applies to future local conversations created from the sidebar.', controlType: 'segmented', status: 'working' },
      { id: 'mock-runtime-scenario', label: 'Mock runtime scenario', helperText: 'Applies only to future mock sends and does not mutate active runs.', controlType: 'segmented', status: 'working' },
    ],
  },
  {
    id: 'runtime-status',
    title: 'Runtime status',
    description: 'Read-only visibility for the currently selected workspace runtime.',
    categoryId: 'general',
    rows: [
      { id: 'data-source-mode', label: 'Data source mode', helperText: 'Shows whether the frontend is using mock data or the local API.', controlType: 'status', status: 'read_only' },
      { id: 'backend-capability', label: 'Backend capability', helperText: 'Displays runtime availability without exposing secrets.', controlType: 'status', status: 'read_only' },
      { id: 'provider-capability', label: 'Provider capability', helperText: 'Shows redacted provider id, family, model, and status when available.', controlType: 'status', status: 'read_only' },
    ],
  },
]

export const placeholderSettingRows: SettingRow[] = [
  { id: 'preview-toggle', label: 'Preview control', helperText: 'Mock only. This control is not connected to providers, tools, files, or backend writes.', controlType: 'placeholder', status: 'disabled', value: null },
  { id: 'not-connected', label: 'Connection state', helperText: 'Not connected in M5.5. Future settings will require a separate implementation plan.', controlType: 'status', status: 'mock', value: 'Preview only' },
]

export const placeholderCategoryIds = settingsCategories
  .filter((category) => category.status === 'mock')
  .map((category) => category.id)

export function getSettingsCategory(categoryId: SettingsCategoryId, locale: Locale = 'en') {
  return getLocalizedSettingsCategories(locale).find((category) => category.id === categoryId) ?? getLocalizedSettingsCategories(locale)[0]
}

export function getSettingsCategoriesByGroup(group: SettingsCategoryGroup, locale: Locale = 'en') {
  return getLocalizedSettingsCategories(locale).filter((category) => category.group === group)
}
