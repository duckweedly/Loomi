import type { LocalProviderDetection, ProviderCapability, ToolCatalogItem, WorkspaceRootConfig } from '../domain'

export type DesktopReadinessAction =
  | 'retry'
  | 'open_settings'
  | 'detect_local_provider'
  | 'enable_local_codex'
  | 'choose_workspace'

export type DesktopReadinessIssue = {
  code:
    | 'ready'
    | 'api_unconnected'
    | 'db_schema_unready'
    | 'provider_unconfigured'
    | 'local_codex_detected_disabled'
    | 'tool_catalog_unavailable'
    | 'workspace_unselected'
  title: string
  detail: string
  action: DesktopReadinessAction
  providerId?: string
}

export type DesktopReadiness = {
  primary: DesktopReadinessIssue
  issues: DesktopReadinessIssue[]
}

type Input = {
  apiConnected: boolean
  dbReady: boolean
  providerCapabilities: ProviderCapability[]
  localProviderDetections: LocalProviderDetection[]
  toolCatalog: ToolCatalogItem[]
  toolCatalogLoaded: boolean
  workspaceRootConfig?: WorkspaceRootConfig | null
}

function canSend(provider: ProviderCapability) {
  return provider.executionState !== 'unsupported' && ['available', 'configured', 'reachable', 'completion-ok'].includes(provider.status)
}

function hasExecutableTool(tools: ToolCatalogItem[]) {
  return tools.some((tool) => tool.enabled && tool.executionState === 'executable')
}

export function deriveDesktopReadiness(input: Input): DesktopReadiness {
  const issues: DesktopReadinessIssue[] = []
  if (!input.apiConnected) {
    issues.push({
      code: 'api_unconnected',
      title: 'Loomi API 未连接',
      detail: '启动本地后端，并确认 VITE_LOOMI_API_BASE_URL 指向同一个地址。',
      action: 'retry',
    })
  }
  if (input.apiConnected && !input.dbReady) {
    issues.push({
      code: 'db_schema_unready',
      title: 'DB/schema 未 ready',
      detail: 'Postgres 或 migration 还没准备好；先看 /readyz 和后端日志。',
      action: 'retry',
    })
  }
  const providerReady = input.providerCapabilities.some(canSend)
  const localCodex = input.localProviderDetections.find((provider) => provider.providerId === 'local_codex' && provider.status === 'available')
  if (input.apiConnected && input.dbReady && !providerReady && localCodex) {
    issues.push({
      code: 'local_codex_detected_disabled',
      title: 'Local Codex detected 但未启用',
      detail: '当前 API session 需要显式启用 Local Codex 后才能作为 provider 发送。',
      action: 'enable_local_codex',
      providerId: 'local_codex',
    })
  } else if (input.apiConnected && input.dbReady && !providerReady) {
    issues.push({
      code: 'provider_unconfigured',
      title: 'provider 未配置',
      detail: '配置一个 OpenAI-compatible provider，或检测并启用 Local Codex。',
      action: 'open_settings',
    })
  }
  if (input.apiConnected && input.dbReady && input.toolCatalogLoaded && !hasExecutableTool(input.toolCatalog)) {
    issues.push({
      code: 'tool_catalog_unavailable',
      title: 'tool catalog 不可用',
      detail: '后端没有返回可执行工具；检查 /v1/tools/catalog 和工具注册。',
      action: 'retry',
    })
  }
  if (input.apiConnected && input.dbReady && (!input.workspaceRootConfig || !input.workspaceRootConfig.configured)) {
    issues.push({
      code: 'workspace_unselected',
      title: 'workspace 未选择',
      detail: '选择一个工作区目录后，workspace tools 才能读取项目文件。',
      action: 'choose_workspace',
    })
  }
  return {
    primary: issues[0] ?? {
      code: 'ready',
      title: 'Desktop ready',
      detail: 'API、DB、provider、tool catalog 和 workspace 已就绪。',
      action: 'retry',
    },
    issues,
  }
}
