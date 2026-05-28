import { describe, expect, test } from 'bun:test'
import type { LocalProviderDetection, ProviderCapability, ToolCatalogItem, WorkspaceRootConfig } from '../domain'
import { deriveDesktopReadiness } from './desktopReadiness'

const availableProvider: ProviderCapability = {
  id: 'custom',
  family: 'openai_compatible',
  model: 'gpt-5.5',
  status: 'available',
  executionState: 'supported',
}

const workspace: WorkspaceRootConfig = { configured: true, displayName: 'Loomi' }

const workspaceTool: ToolCatalogItem = {
  name: 'workspace.read',
  displayName: 'Workspace read',
  description: 'Read files',
  source: 'builtin',
  group: 'workspace',
  riskLevel: 'low',
  approvalPolicy: 'always_required',
  enabled: true,
  executionState: 'executable',
}

const detectedLocalCodex: LocalProviderDetection = {
  providerId: 'local_codex',
  displayName: 'Local Codex',
  providerKind: 'codex',
  authMode: 'oauth',
  status: 'available',
  modelCandidates: ['gpt-5'],
  source: 'local_config',
  redactionApplied: true,
}

describe('desktop readiness', () => {
  test('reports the first actionable real desktop blocker', () => {
    expect(deriveDesktopReadiness({
      apiConnected: false,
      dbReady: false,
      providerCapabilities: [],
      localProviderDetections: [],
      toolCatalog: [],
      toolCatalogLoaded: false,
      workspaceRootConfig: null,
    }).primary.code).toBe('api_unconnected')

    expect(deriveDesktopReadiness({
      apiConnected: true,
      dbReady: false,
      providerCapabilities: [availableProvider],
      localProviderDetections: [],
      toolCatalog: [workspaceTool],
      toolCatalogLoaded: true,
      workspaceRootConfig: workspace,
    }).primary.code).toBe('db_schema_unready')

    expect(deriveDesktopReadiness({
      apiConnected: true,
      dbReady: true,
      providerCapabilities: [],
      localProviderDetections: [detectedLocalCodex],
      toolCatalog: [workspaceTool],
      toolCatalogLoaded: true,
      workspaceRootConfig: workspace,
    }).primary.code).toBe('local_codex_detected_disabled')

    expect(deriveDesktopReadiness({
      apiConnected: true,
      dbReady: true,
      providerCapabilities: [],
      localProviderDetections: [],
      toolCatalog: [workspaceTool],
      toolCatalogLoaded: true,
      workspaceRootConfig: workspace,
    }).primary.code).toBe('provider_unconfigured')

    expect(deriveDesktopReadiness({
      apiConnected: true,
      dbReady: true,
      providerCapabilities: [availableProvider],
      localProviderDetections: [],
      toolCatalog: [],
      toolCatalogLoaded: true,
      workspaceRootConfig: workspace,
    }).primary.code).toBe('tool_catalog_unavailable')

    expect(deriveDesktopReadiness({
      apiConnected: true,
      dbReady: true,
      providerCapabilities: [availableProvider],
      localProviderDetections: [],
      toolCatalog: [workspaceTool],
      toolCatalogLoaded: true,
      workspaceRootConfig: { configured: false, displayName: 'No folder selected' },
    }).primary.code).toBe('workspace_unselected')
  })

  test('does not report ready while workspace root config is still missing', () => {
    const readiness = deriveDesktopReadiness({
      apiConnected: true,
      dbReady: true,
      providerCapabilities: [availableProvider],
      localProviderDetections: [],
      toolCatalog: [workspaceTool],
      toolCatalogLoaded: true,
      workspaceRootConfig: null,
    })

    expect(readiness.primary.code).toBe('workspace_unselected')
  })
})
