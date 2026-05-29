import type { Locale } from '../i18n'

export type SettingsCategoryGroup = 'primary' | 'agent_core' | 'management'
export type SettingsCategoryStatus = 'working' | 'mock' | 'preview' | 'read_only' | 'mixed'
export type SettingRowControlType = 'toggle' | 'select' | 'button' | 'status' | 'placeholder' | 'segmented'
export type SettingRowStatus = 'working' | 'mock' | 'disabled' | 'read_only'

export type SettingsCategoryId =
  | 'general'
  | 'appearance'
  | 'providers'
  | 'web-search'
  | 'skill'
  | 'mcp'
  | 'memory'
  | 'tools'
  | 'about'

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
  agent_core: { zh: '模型与上下文', en: 'Models & Context' },
  management: { zh: '能力与安全', en: 'Capabilities & Safety' },
}

const categoryCopy: Record<SettingsCategoryId, Record<Locale, Pick<SettingsCategory, 'label' | 'description'>>> = {
  general: { zh: { label: '通用', description: '当前会话的基础偏好。' }, en: { label: 'General', description: 'Basic preferences for the current session.' } },
  appearance: { zh: { label: '外观', description: '语言、主题和当前界面的显示偏好。' }, en: { label: 'Appearance', description: 'Language, theme, and display preferences for this workspace.' } },
  providers: { zh: { label: '模型连接', description: '管理模型供应商、可用模型与本地执行通道。' }, en: { label: 'Model Connections', description: 'Manage model providers, available models, and local execution channels.' } },
  'web-search': { zh: { label: '网页搜索', description: '管理网页搜索密钥和可用状态。' }, en: { label: 'Web Search', description: 'Manage web search keys and availability.' } },
  skill: { zh: { label: '技能', description: '查看本机已安装的 Codex、Claude Code 和项目技能。' }, en: { label: 'Skill', description: 'View installed Codex, Claude Code, and project skills.' } },
  mcp: { zh: { label: 'MCP', description: '管理本地 stdio MCP server 配置、保存状态和连接测试。' }, en: { label: 'MCP', description: 'Manage local stdio MCP server config, saved state, and connection tests.' } },
  memory: { zh: { label: '记忆与上下文', description: '审批记忆提案、查看运行上下文并管理安全记忆。' }, en: { label: 'Memory & Context', description: 'Review memory proposals, inspect runtime context, and manage safe memory.' } },
  tools: { zh: { label: '工具权限', description: '只读查看 builtin 和 MCP 工具目录、安全策略和执行状态。' }, en: { label: 'Tool Permissions', description: 'Read-only catalog for builtin and MCP tools, safety policy, and execution state.' } },
  about: { zh: { label: '关于', description: '当前本地应用和连接状态。' }, en: { label: 'About', description: 'Current local app and connection state.' } },
}

const categoryStructure: Omit<SettingsCategory, 'label' | 'description'>[] = [
  { id: 'general', group: 'primary', status: 'working' },
  { id: 'appearance', group: 'primary', status: 'working' },
  { id: 'providers', group: 'agent_core', status: 'mixed' },
  { id: 'web-search', group: 'agent_core', status: 'mixed' },
  { id: 'skill', group: 'agent_core', status: 'read_only' },
  { id: 'mcp', group: 'agent_core', status: 'working' },
  { id: 'memory', group: 'agent_core', status: 'working' },
  { id: 'tools', group: 'management', status: 'read_only' },
  { id: 'about', group: 'management', status: 'mixed' },
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
    ],
  },
]

export function getSettingsCategory(categoryId: SettingsCategoryId, locale: Locale = 'en') {
  return getLocalizedSettingsCategories(locale).find((category) => category.id === categoryId) ?? getLocalizedSettingsCategories(locale)[0]
}

export function getSettingsCategoriesByGroup(group: SettingsCategoryGroup, locale: Locale = 'en') {
  return getLocalizedSettingsCategories(locale).filter((category) => category.group === group)
}
