import { describe, expect, test } from 'bun:test'
import type { RunEvent } from '../domain'
import { deriveBackendCapabilityStatus, deriveCapabilitySignalFromEvent, getBackendCapabilityCopy } from './backendCapabilityStatus'

describe('backend capability status', () => {
  test('applies capability precedence from recovering through mock', () => {
    expect(deriveBackendCapabilityStatus({ dataSourceMode: 'mock' })).toBe('mock')
    expect(deriveBackendCapabilityStatus({ dataSourceMode: 'real_api', runtimeSource: 'local_simulated' })).toBe('local-simulated')
    expect(deriveBackendCapabilityStatus({ dataSourceMode: 'real_api', runtimeSource: 'model_gateway' })).toBe('model-gateway')
    expect(deriveBackendCapabilityStatus({ dataSourceMode: 'real_api', backendUnavailable: true, runtimeSource: 'model_gateway' })).toBe('backend-unavailable')
    expect(deriveBackendCapabilityStatus({ dataSourceMode: 'real_api', modelSetupMissing: true, backendUnavailable: true })).toBe('model-setup-missing')
    expect(deriveBackendCapabilityStatus({ dataSourceMode: 'real_api', providerUnavailable: true, modelSetupMissing: true })).toBe('provider-unavailable')
    expect(deriveBackendCapabilityStatus({ dataSourceMode: 'real_api', streamDisconnected: true, activeRun: true, providerUnavailable: true })).toBe('stream-disconnected')
    expect(deriveBackendCapabilityStatus({ dataSourceMode: 'real_api', streamDisconnected: true, activeRun: false, providerUnavailable: true })).toBe('provider-unavailable')
    expect(deriveBackendCapabilityStatus({ dataSourceMode: 'real_api', runRecovering: true, streamDisconnected: true })).toBe('run-recovering')
  })

  test('keeps mock and local simulated copy distinct from model gateway execution', () => {
    expect(getBackendCapabilityCopy('mock').title).toBe('Mock')
    expect(getBackendCapabilityCopy('mock').detail).toContain('not real model output')
    expect(getBackendCapabilityCopy('local-simulated').title).toBe('Local simulated')
    expect(getBackendCapabilityCopy('local-simulated').detail).toContain('simulated')
    expect(getBackendCapabilityCopy('model-gateway').detail).toContain('Real provider')
  })

  test('distinguishes backend unavailable, setup missing, provider unavailable, and stream disconnected', () => {
    expect(getBackendCapabilityCopy('backend-unavailable').detail).toContain('backend')
    expect(getBackendCapabilityCopy('model-setup-missing').detail).toContain('setup')
    expect(getBackendCapabilityCopy('provider-unavailable').detail).toContain('provider')
    expect(getBackendCapabilityCopy('stream-disconnected').detail).toContain('stream')
    expect(getBackendCapabilityCopy('run-recovering').detail).toContain('recover')
  })

  test('derives capability signals from runtime events', () => {
    const base: RunEvent = { id: 'evt', runId: 'run-a', threadId: 'thread-a', type: 'provider.error', label: 'Provider', detail: 'Provider failed', time: 'Now', status: 'running' }

    expect(deriveCapabilitySignalFromEvent(base)).toEqual({ providerUnavailable: true })
    expect(deriveCapabilitySignalFromEvent({ ...base, type: 'model.setup_missing', detail: 'setup missing' })).toEqual({ modelSetupMissing: true })
    expect(deriveCapabilitySignalFromEvent({ ...base, type: 'backend.unavailable', detail: 'backend unavailable' })).toEqual({ backendUnavailable: true })
    expect(deriveCapabilitySignalFromEvent({ ...base, type: 'run.recovering', status: 'recovering' })).toEqual({ runRecovering: true })
  })
})
