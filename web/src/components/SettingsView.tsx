import { ArrowLeft, ChevronRight } from 'lucide-react'
import type { ReactNode } from 'react'
import type { BackendCapabilityState, ProviderCapability, RunStatus, RuntimeScriptId, StreamState, Thread } from '../domain'
import type { ProviderCheckResult } from '../state'
import type { ProviderDraftSettings } from '../useWorkspaceShellState'
import type { Locale } from '../i18n'
import { getDictionary } from '../i18n'
import { getSettingsCategoriesByGroup, getSettingsCategory, placeholderSettingRows, settingsCategoryGroups, type SettingsCategory, type SettingsCategoryGroup, type SettingsCategoryId, type SettingRowStatus } from './settingsCatalog'

type Props = {
  locale: Locale
  selectedCategoryId: SettingsCategoryId
  defaultWorkspaceMode: Thread['mode']
  selectedRuntimeScript: RuntimeScriptId
  dataSourceMode: 'mock' | 'real_api'
  backendCapability: BackendCapabilityState
  streamState: StreamState
  selectedThreadTitle?: string
  selectedRunStatus?: RunStatus
  providerCapabilities: ProviderCapability[]
  providerCheckResults: Record<string, ProviderCheckResult>
  providerDraftSettings: ProviderDraftSettings
  onSelectLocale: (locale: Locale) => void
  onSelectCategory: (categoryId: SettingsCategoryId) => void
  onSelectDefaultWorkspaceMode: (mode: Thread['mode']) => void
  onSelectRuntimeScript: (scriptId: RuntimeScriptId) => void
  onProviderDraftSettingsChange: (settings: ProviderDraftSettings) => void
  onCheckProvider: (providerId: string) => void
  onBack: () => void
}

const categoryGroups = Object.keys(settingsCategoryGroups) as SettingsCategoryGroup[]

function statusLabel(status: SettingRowStatus, t: ReturnType<typeof getDictionary>['settings']) {
  if (status === 'working') return t.working
  if (status === 'read_only') return t.readOnly
  if (status === 'disabled') return t.disabled
  return t.previewOnly
}

function capabilityLabel(status: BackendCapabilityState, t: ReturnType<typeof getDictionary>['settings']) {
  if (status === 'available') return t.available
  if (status === 'misconfigured') return t.misconfigured
  return t.unavailable
}

function categoryStatusLabel(category: SettingsCategory, t: ReturnType<typeof getDictionary>['settings']) {
  if (category.status === 'working') return t.working
  if (category.status === 'read_only') return t.readOnly
  if (category.status === 'mixed') return t.mixed
  return t.previewOnly
}

function SettingRow({ label, helperText, status, control, t }: { label: string; helperText: string; status: SettingRowStatus; control: ReactNode; t: ReturnType<typeof getDictionary>['settings'] }) {
  return (
    <div className="setting-row">
      <div className="setting-row-copy">
        <div className="setting-row-title">
          <span>{label}</span>
          <span className={`setting-status-badge ${status}`}>{statusLabel(status, t)}</span>
        </div>
        <p>{helperText}</p>
      </div>
      <div className="setting-row-control">{control}</div>
    </div>
  )
}

function SegmentedControl<T extends string>({ value, options, onChange }: { value: T; options: { value: T; label: string }[]; onChange: (value: T) => void }) {
  return (
    <div className="settings-segmented-control">
      {options.map((option) => (
        <button key={option.value} className={option.value === value ? 'selected' : undefined} onClick={() => onChange(option.value)}>
          {option.label}
        </button>
      ))}
    </div>
  )
}

function StatusValue({ children }: { children: ReactNode }) {
  return <span className="settings-status-value">{children}</span>
}

function ProviderTextInput({ value, placeholder, type = 'text', onChange }: { value: string; placeholder: string; type?: 'text' | 'password'; onChange: (value: string) => void }) {
  return <input className="settings-text-input" type={type} value={value} placeholder={placeholder} onChange={(event) => onChange(event.target.value)} />
}

function ProviderCapabilityList({ providerCapabilities, t }: { providerCapabilities: ProviderCapability[]; t: ReturnType<typeof getDictionary>['settings'] }) {
  if (!providerCapabilities.length) return <StatusValue>{t.notConnected}</StatusValue>
  return (
    <div className="provider-capability-list">
      {providerCapabilities.map((provider) => (
        <span key={provider.id}>
          {provider.id} · {provider.family} · {provider.model} · {capabilityLabel(provider.status, t)}
        </span>
      ))}
    </div>
  )
}

function ProviderCheckConsole({ providerCapabilities, providerCheckResults, onCheckProvider, t }: Pick<Props, 'providerCapabilities' | 'providerCheckResults' | 'onCheckProvider'> & { t: ReturnType<typeof getDictionary>['settings'] }) {
  if (!providerCapabilities.length) {
    return (
      <div className="provider-console-empty">
        <strong>{t.providerConsoleEmpty}</strong>
        <span>{t.providerConsoleEnvGuide}</span>
      </div>
    )
  }

  return (
    <div className="provider-console-list">
      {providerCapabilities.map((provider) => {
        const result = providerCheckResults[provider.id]
        const checking = result?.status === 'checking'
        const resultText = result?.status ? t.providerCheckResult(result.status, result.message) : provider.message
        return (
          <article className="provider-console-card" key={provider.id}>
            <div className="provider-console-main">
              <div className="provider-console-title">
                <strong>{provider.id}</strong>
                <span className={`setting-status-badge ${provider.status}`}>{capabilityLabel(provider.status, t)}</span>
              </div>
              <div className="provider-console-meta">
                <span>{provider.family}</span>
                <span>{provider.model}</span>
              </div>
              {resultText && <p className={`provider-check-result ${result?.status ?? 'idle'}`}>{resultText}</p>}
            </div>
            <button className="provider-test-button" disabled={checking} onClick={() => onCheckProvider(provider.id)}>
              {checking ? t.providerChecking : t.providerTestConnection}
            </button>
          </article>
        )
      })}
    </div>
  )
}

function RuntimeStatusRows({ dataSourceMode, backendCapability, streamState, selectedThreadTitle, selectedRunStatus, providerCapabilities, t }: Pick<Props, 'dataSourceMode' | 'backendCapability' | 'streamState' | 'selectedThreadTitle' | 'selectedRunStatus' | 'providerCapabilities'> & { t: ReturnType<typeof getDictionary>['settings'] }) {
  return (
    <>
      <SettingRow label={t.dataSourceMode} helperText={t.dataSourceModeHelper} status="read_only" t={t} control={<StatusValue>{dataSourceMode === 'real_api' ? 'Real API' : 'Mock'}</StatusValue>} />
      <SettingRow label={t.backendCapability} helperText={t.backendCapabilityHelper} status="read_only" t={t} control={<StatusValue>{capabilityLabel(backendCapability, t)}</StatusValue>} />
      <SettingRow label={t.streamState} helperText={t.streamStateHelper} status="read_only" t={t} control={<StatusValue>{streamState}</StatusValue>} />
      <SettingRow label={t.selectedThread} helperText={t.selectedThreadHelper} status="read_only" t={t} control={<StatusValue>{selectedThreadTitle ?? t.noThreadSelected}</StatusValue>} />
      <SettingRow label={t.selectedRunStatus} helperText={t.selectedRunStatusHelper} status="read_only" t={t} control={<StatusValue>{selectedRunStatus ?? t.noActiveRun}</StatusValue>} />
      <SettingRow label={t.providerCapability} helperText={t.providerCapabilityHelper} status="read_only" t={t} control={<ProviderCapabilityList providerCapabilities={providerCapabilities} t={t} />} />
    </>
  )
}

function PlaceholderPanel({ selectedCategory, t }: { selectedCategory: SettingsCategory; t: ReturnType<typeof getDictionary>['settings'] }) {
  return (
    <section className="settings-card placeholder-panel">
      <div className="settings-card-head">
        <h2>{t.categoryPreview(selectedCategory.label)}</h2>
        <p>{selectedCategory.description}</p>
      </div>
      {placeholderSettingRows.map((row) => (
        <SettingRow
          key={`${selectedCategory.id}-${row.id}`}
          label={row.id === 'preview-toggle' ? t.previewControl : t.connectionState}
          helperText={row.id === 'preview-toggle' ? t.previewControlHelper : t.connectionStateHelper}
          status={row.status}
          t={t}
          control={<button className="settings-disabled-control" disabled>{t.previewOnly}</button>}
        />
      ))}
    </section>
  )
}

export function SettingsView({
  locale,
  selectedCategoryId,
  defaultWorkspaceMode,
  selectedRuntimeScript,
  dataSourceMode,
  backendCapability,
  streamState,
  selectedThreadTitle,
  selectedRunStatus,
  providerCapabilities,
  providerCheckResults,
  providerDraftSettings,
  onSelectLocale,
  onSelectCategory,
  onSelectDefaultWorkspaceMode,
  onSelectRuntimeScript,
  onProviderDraftSettingsChange,
  onCheckProvider,
  onBack,
}: Props) {
  const dictionary = getDictionary(locale)
  const t = dictionary.settings
  const selectedCategory = getSettingsCategory(selectedCategoryId, locale)
  const isGeneral = selectedCategory.id === 'general'
  const isProviders = selectedCategory.id === 'providers'
  const isAbout = selectedCategory.id === 'about'

  return (
    <div className="settings-shell" aria-label={t.title}>
      <nav className="settings-sidebar" aria-label="Settings categories">
        <button className="settings-back-button" onClick={onBack}>
          <ArrowLeft size={15} /> {t.back}
        </button>
        {categoryGroups.map((group) => (
          <div className="settings-nav-group" key={group}>
            <span>{settingsCategoryGroups[group][locale]}</span>
            {getSettingsCategoriesByGroup(group, locale).map((category) => (
              <button key={category.id} className={category.id === selectedCategoryId ? 'selected' : undefined} onClick={() => onSelectCategory(category.id)}>
                <span>{category.label}</span>
                <small>{categoryStatusLabel(category, t)}</small>
                <ChevronRight size={13} />
              </button>
            ))}
          </div>
        ))}
      </nav>

      <section className="settings-content">
        <header className="settings-content-header">
          <div>
            <span className="settings-eyebrow">{t.title}</span>
            <h1>{selectedCategory.label}</h1>
            <p>{selectedCategory.description}</p>
          </div>
          <span className={`settings-category-pill ${selectedCategory.status}`}>{categoryStatusLabel(selectedCategory, t)}</span>
        </header>

        {isGeneral && (
          <div className="settings-card-stack">
            <section className="settings-card">
              <div className="settings-card-head">
                <h2>{t.workspaceDefaults}</h2>
                <p>{t.workspaceDefaultsDescription}</p>
              </div>
              <SettingRow
                label={t.language}
                helperText={t.languageHelper}
                status="working"
                t={t}
                control={(
                  <SegmentedControl
                    value={locale}
                    options={[{ value: 'zh', label: t.chinese }, { value: 'en', label: t.english }]}
                    onChange={onSelectLocale}
                  />
                )}
              />
              <SettingRow
                label={t.defaultWorkspaceMode}
                helperText={t.defaultWorkspaceModeHelper}
                status="working"
                t={t}
                control={(
                  <SegmentedControl
                    value={defaultWorkspaceMode}
                    options={[{ value: 'chat', label: dictionary.app.chat }, { value: 'work', label: dictionary.app.work }]}
                    onChange={onSelectDefaultWorkspaceMode}
                  />
                )}
              />
              <SettingRow
                label={t.mockRuntimeScenario}
                helperText={t.mockRuntimeScenarioHelper}
                status="working"
                t={t}
                control={(
                  <SegmentedControl
                    value={selectedRuntimeScript}
                    options={[{ value: 'success', label: t.success }, { value: 'failure', label: t.failure }]}
                    onChange={onSelectRuntimeScript}
                  />
                )}
              />
            </section>

            <section className="settings-card">
              <div className="settings-card-head">
                <h2>{t.runtimeStatus}</h2>
                <p>{t.runtimeStatusDescription}</p>
              </div>
              <RuntimeStatusRows dataSourceMode={dataSourceMode} backendCapability={backendCapability} streamState={streamState} selectedThreadTitle={selectedThreadTitle} selectedRunStatus={selectedRunStatus} providerCapabilities={providerCapabilities} t={t} />
            </section>
          </div>
        )}

        {isProviders && (
          <div className="settings-card-stack">
            <section className="settings-card">
              <div className="settings-card-head">
                <h2>{t.providerConsoleTitle}</h2>
                <p>{t.providerConsoleDescription}</p>
              </div>
              <div className="setting-row provider-console-row">
                <div className="setting-row-copy">
                  <div className="setting-row-title">
                    <span>{t.providerConfiguredProviders}</span>
                    <span className="setting-status-badge read_only">{t.readOnly}</span>
                  </div>
                  <p>{t.providerConfiguredProvidersHelper}</p>
                </div>
                <ProviderCheckConsole providerCapabilities={providerCapabilities} providerCheckResults={providerCheckResults} onCheckProvider={onCheckProvider} t={t} />
              </div>
            </section>

            <section className="settings-card">
              <div className="settings-card-head">
                <h2>{t.providerLocalDraftTitle}</h2>
                <p>{t.providerLocalDraftDescription}</p>
              </div>
              <SettingRow
                label={t.providerBaseUrl}
                helperText={t.providerBaseUrlHelper}
                status="working"
                t={t}
                control={(
                  <ProviderTextInput
                    value={providerDraftSettings.baseUrl}
                    placeholder="https://gateway.example.test/v1"
                    onChange={(baseUrl) => onProviderDraftSettingsChange({ ...providerDraftSettings, baseUrl })}
                  />
                )}
              />
              <SettingRow
                label={t.providerModel}
                helperText={t.providerModelHelper}
                status="working"
                t={t}
                control={(
                  <ProviderTextInput
                    value={providerDraftSettings.model}
                    placeholder="gpt-5.5"
                    onChange={(model) => onProviderDraftSettingsChange({ ...providerDraftSettings, model })}
                  />
                )}
              />
              <SettingRow
                label={t.providerApiKey}
                helperText={t.providerApiKeyHelper}
                status="working"
                t={t}
                control={(
                  <div className="settings-secret-control">
                    <ProviderTextInput
                      value=""
                      type="password"
                      placeholder={providerDraftSettings.apiKeySet ? t.providerConfigured : t.providerNotConfigured}
                      onChange={(apiKey) => onProviderDraftSettingsChange({ ...providerDraftSettings, apiKeySet: apiKey.length > 0 })}
                    />
                    <StatusValue>{providerDraftSettings.apiKeySet ? t.providerConfigured : t.providerNotConfigured}</StatusValue>
                  </div>
                )}
              />
            </section>
          </div>
        )}

        {isAbout && (
          <div className="settings-card-stack">
            <section className="settings-card">
              <div className="settings-card-head">
                <h2>{t.aboutLocalApp}</h2>
                <p>{t.aboutLocalAppDescription}</p>
              </div>
              <SettingRow label={t.appName} helperText={t.appNameHelper} status="read_only" t={t} control={<StatusValue>Loomi</StatusValue>} />
              <SettingRow label={t.appVersion} helperText={t.appVersionHelper} status="mock" t={t} control={<StatusValue>{t.previewOnly}</StatusValue>} />
              <SettingRow label={t.appStatus} helperText={t.appStatusHelper} status="read_only" t={t} control={<StatusValue>{dataSourceMode === 'real_api' ? 'Real API' : 'Mock'} · {capabilityLabel(backendCapability, t)}</StatusValue>} />
            </section>
          </div>
        )}

        {!isGeneral && !isProviders && !isAbout && <PlaceholderPanel selectedCategory={selectedCategory} t={t} />}
      </section>
    </div>
  )
}
