import type { Locale } from '../i18n'

export type SettingsCategoryGroup = 'primary' | 'agent_core' | 'management'
export type SettingsCategoryStatus = 'working' | 'mock' | 'preview' | 'read_only' | 'mixed'
export type SettingRowControlType = 'toggle' | 'select' | 'button' | 'status' | 'placeholder' | 'segmented'
export type SettingRowStatus = 'working' | 'mock' | 'disabled' | 'read_only'

export type SettingsCategoryId =
  | 'general'
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
  agent_core: { zh: 'Agent Core', en: 'Agent Core' },
  management: { zh: '管理', en: 'Management' },
}

const categoryCopy: Record<SettingsCategoryId, Record<Locale, Pick<SettingsCategory, 'label' | 'description'>>> = {
  general: { zh: { label: '通用', description: '当前会话的工作区默认值和运行状态可见性。' }, en: { label: 'General', description: 'Current-session workspace defaults and runtime visibility.' } },
  providers: { zh: { label: '供应商', description: '管理模型供应商与可用模型。' }, en: { label: 'Providers', description: 'Manage model providers and available models.' } },
  'web-search': { zh: { label: '网页搜索', description: '管理 Tavily 和 Brave Search 的网页搜索能力。' }, en: { label: 'Web Search', description: 'Manage Tavily and Brave Search web search capability.' } },
  skill: { zh: { label: '技能', description: '查看本机已安装的 Codex、Claude Code 和项目技能。' }, en: { label: 'Skill', description: 'View installed Codex, Claude Code, and project skills.' } },
  mcp: { zh: { label: 'MCP', description: '管理本地 stdio MCP server 配置、保存状态和连接测试。' }, en: { label: 'MCP', description: 'Manage local stdio MCP server config, saved state, and connection tests.' } },
  memory: { zh: { label: 'Memory', description: '查看、检索并删除已批准的本地记忆。' }, en: { label: 'Memory', description: 'View, search, and delete approved local memories.' } },
  tools: { zh: { label: '工具', description: '只读查看 builtin 和 MCP 工具目录、安全策略和执行状态。' }, en: { label: 'Tools', description: 'Read-only catalog for builtin and MCP tools, safety policy, and execution state.' } },
  about: { zh: { label: '关于', description: '已知本地应用状态和占位构建信息。' }, en: { label: 'About', description: 'Known local app state plus placeholder build metadata.' } },
}

const categoryStructure: Omit<SettingsCategory, 'label' | 'description'>[] = [
  { id: 'general', group: 'primary', status: 'working' },
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
      { id: 'theme', label: 'Theme', helperText: 'Switches the current workspace between light and dark appearance.', controlType: 'segmented', status: 'working' },
    ],
  },
  {
    id: 'runtime-status',
    title: 'Runtime status',
    description: 'Read-only visibility for the currently selected workspace runtime.',
    categoryId: 'general',
    rows: [
      { id: 'backend-capability', label: 'Backend capability', helperText: 'Displays runtime availability without exposing secrets.', controlType: 'status', status: 'read_only' },
      { id: 'provider-capability', label: 'Provider capability', helperText: 'Shows redacted provider id, family, model, and status when available.', controlType: 'status', status: 'read_only' },
    ],
  },
]

export function getSettingsCategory(categoryId: SettingsCategoryId, locale: Locale = 'en') {
  return getLocalizedSettingsCategories(locale).find((category) => category.id === categoryId) ?? getLocalizedSettingsCategories(locale)[0]
}

export function getSettingsCategoriesByGroup(group: SettingsCategoryGroup, locale: Locale = 'en') {
  return getLocalizedSettingsCategories(locale).filter((category) => category.group === group)
}
