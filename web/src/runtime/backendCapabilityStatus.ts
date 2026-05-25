import type { ProviderCapability } from '../domain'
import type { Locale } from '../i18n'
import { getDictionary } from '../i18n'

export type BackendCapabilityStatus =
  | 'mock'
  | 'local-simulated'
  | 'model-gateway'
  | 'backend-unavailable'
  | 'model-setup-missing'
  | 'provider-unavailable'
  | 'stream-disconnected'
  | 'run-recovering'

export type BackendCapabilityInput = {
  dataSourceMode: 'mock' | 'real_api' | 'model_gateway'
  runtimeSource?: 'local_simulated' | 'model_gateway'
  backendUnavailable?: boolean
  modelSetupMissing?: boolean
  providerUnavailable?: boolean
  streamDisconnected?: boolean
  activeRun?: boolean
  runRecovering?: boolean
}

export function shouldShowProviderUnavailableWarning(dataSourceMode: 'mock' | 'real_api' | 'model_gateway', providerCapabilities: ProviderCapability[]) {
  return dataSourceMode !== 'mock' && !providerCapabilities.some((provider) => provider.status === 'available' && provider.executionState !== 'unsupported')
}

export function getProviderUnavailableWarning(providerCapabilities: ProviderCapability[], locale: Locale = 'en') {
  const copy = getDictionary(locale).chatCanvas
  const localCodex = providerCapabilities.find((provider) => provider.id === 'local_codex' && provider.localProvider)
  if (localCodex?.executionState === 'unsupported') return copy.localCodexUnsupportedWarning
  if (localCodex?.status === 'unavailable') return copy.localCodexUnavailableWarning
  return copy.providerUnavailableWarning
}

export function deriveBackendCapabilityStatus(input: BackendCapabilityInput): BackendCapabilityStatus {
  if (input.runRecovering) return 'run-recovering'
  if (input.streamDisconnected && input.activeRun !== false) return 'stream-disconnected'
  if (input.providerUnavailable) return 'provider-unavailable'
  if (input.modelSetupMissing) return 'model-setup-missing'
  if (input.backendUnavailable) return 'backend-unavailable'
  if (input.dataSourceMode === 'mock') return 'mock'
  return input.runtimeSource === 'model_gateway' ? 'model-gateway' : 'local-simulated'
}

export function deriveCapabilitySignalFromEvent(event: { type: string; detail?: string; status?: string; severity?: string }) {
  const text = `${event.type} ${event.detail ?? ''}`.toLowerCase()
  if (event.status === 'recovering' || event.type === 'run.recovering') return { runRecovering: true }
  if (text.includes('provider')) return { providerUnavailable: true }
  if (text.includes('setup') || text.includes('missing')) return { modelSetupMissing: true }
  if (text.includes('backend') || text.includes('unavailable')) return { backendUnavailable: true }
  return {}
}

export function getBackendCapabilityCopy(status: BackendCapabilityStatus, locale: Locale = 'en') {
  return getDictionary(locale).backendCapability[status]
}
