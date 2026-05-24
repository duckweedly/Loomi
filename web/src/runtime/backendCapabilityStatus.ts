export type BackendCapabilityStatus =
  | 'mock'
  | 'local-simulated'
  | 'real-model'
  | 'backend-unavailable'
  | 'model-setup-missing'
  | 'provider-unavailable'
  | 'stream-disconnected'
  | 'run-recovering'

export type BackendCapabilityInput = {
  dataSourceMode: 'mock' | 'real_api'
  runtimeSource?: 'local_simulated' | 'real_model'
  backendUnavailable?: boolean
  modelSetupMissing?: boolean
  providerUnavailable?: boolean
  streamDisconnected?: boolean
  activeRun?: boolean
  runRecovering?: boolean
}

const capabilityCopy: Record<BackendCapabilityStatus, { title: string; detail: string }> = {
  mock: { title: 'Mock', detail: 'Deterministic local behavior; not real model output.' },
  'local-simulated': { title: 'Local simulated', detail: 'Real API path is connected, but generation is simulated.' },
  'real-model': { title: 'Real model', detail: 'real provider execution is available.' },
  'backend-unavailable': { title: 'Backend unavailable', detail: 'The backend cannot provide runtime execution.' },
  'model-setup-missing': { title: 'Model setup missing', detail: 'Model setup or credentials are required before generation.' },
  'provider-unavailable': { title: 'Provider unavailable', detail: 'The provider rejected or failed generation.' },
  'stream-disconnected': { title: 'Stream disconnected', detail: 'The event stream disconnected before terminal reconciliation.' },
  'run-recovering': { title: 'Run recovering', detail: 'The UI is recovering the latest known run state.' },
}

export function deriveBackendCapabilityStatus(input: BackendCapabilityInput): BackendCapabilityStatus {
  if (input.runRecovering) return 'run-recovering'
  if (input.streamDisconnected && input.activeRun !== false) return 'stream-disconnected'
  if (input.providerUnavailable) return 'provider-unavailable'
  if (input.modelSetupMissing) return 'model-setup-missing'
  if (input.backendUnavailable) return 'backend-unavailable'
  if (input.dataSourceMode === 'mock') return 'mock'
  return input.runtimeSource === 'real_model' ? 'real-model' : 'local-simulated'
}

export function deriveCapabilitySignalFromEvent(event: { type: string; detail?: string; status?: string; severity?: string }) {
  const text = `${event.type} ${event.detail ?? ''}`.toLowerCase()
  if (event.status === 'recovering' || event.type === 'run.recovering') return { runRecovering: true }
  if (text.includes('provider')) return { providerUnavailable: true }
  if (text.includes('setup') || text.includes('missing')) return { modelSetupMissing: true }
  if (text.includes('backend') || text.includes('unavailable')) return { backendUnavailable: true }
  return {}
}

export function getBackendCapabilityCopy(status: BackendCapabilityStatus) {
  return capabilityCopy[status]
}
