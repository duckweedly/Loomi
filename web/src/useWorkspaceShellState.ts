import { useEffect, useMemo, useState } from 'react'
import type { Thread } from './domain'
import type { Locale } from './i18n'
import type { SettingsCategoryId } from './components/settingsCatalog'
import type { RightPanelItemId } from './rightPanelItems'

export type ProviderDraftSettings = {
  baseUrl: string
  model: string
  apiKey: string
  apiKeySet: boolean
}

type WorkspaceShellSnapshot = {
  defaultWorkspaceMode: Thread['mode']
  locale: Locale
  providerDraftSettings: ProviderDraftSettings
  rightPanelMenuOpen: boolean
  rightPanelOpen: boolean
  previewArtifactId?: string
  selectedRightPanelId: RightPanelItemId
  settingsCategoryId: SettingsCategoryId
  settingsOpen: boolean
  sidebarCollapsed: boolean
  sidebarWidth: number
  theme: 'dark' | 'light'
}

type Theme = WorkspaceShellSnapshot['theme']

const themeStorageKey = 'loomi.theme'

function isTheme(value: string | null): value is Theme {
  return value === 'dark' || value === 'light'
}

function readStoredTheme(): Theme | null {
  if (typeof window === 'undefined') return null
  try {
    const storedTheme = window.localStorage.getItem(themeStorageKey)
    return isTheme(storedTheme) ? storedTheme : null
  } catch {
    return null
  }
}

function writeStoredTheme(theme: Theme) {
  if (typeof window === 'undefined') return
  try {
    window.localStorage.setItem(themeStorageKey, theme)
  } catch {
    // localStorage may be unavailable in privacy-restricted desktop contexts.
  }
}

export function resolveSystemTheme(): Theme {
  if (typeof window === 'undefined' || !window.matchMedia) return 'light'
  return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
}

function createInitialShellSnapshot(initialState: Partial<WorkspaceShellSnapshot> = {}): WorkspaceShellSnapshot {
  return {
    ...baseShellState,
    theme: readStoredTheme() ?? resolveSystemTheme(),
    ...initialState,
  }
}

const baseShellState: WorkspaceShellSnapshot = {
  defaultWorkspaceMode: 'chat',
  locale: 'zh',
  providerDraftSettings: { baseUrl: '', model: '', apiKey: '', apiKeySet: false },
  rightPanelMenuOpen: false,
  rightPanelOpen: false,
  selectedRightPanelId: 'preview',
  settingsCategoryId: 'general',
  settingsOpen: false,
  sidebarCollapsed: false,
  sidebarWidth: 340,
  theme: 'light',
}

type WorkspaceShellAction =
  | { type: 'setDefaultWorkspaceMode'; defaultWorkspaceMode: Thread['mode'] }
  | { type: 'setLocale'; locale: Locale }
  | { type: 'setProviderDraftSettings'; providerDraftSettings: ProviderDraftSettings }
  | { type: 'setSettingsCategory'; settingsCategoryId: SettingsCategoryId }
  | { type: 'openSettings'; settingsCategoryId?: SettingsCategoryId }
  | { type: 'closeSettings' }
  | { type: 'setSidebarWidth'; sidebarWidth: number }
  | { type: 'setSidebarCollapsed'; sidebarCollapsed: boolean }
  | { type: 'toggleTheme' }
  | { type: 'togglePreviewPanel' }
  | { type: 'openRightPanel'; selectedRightPanelId: RightPanelItemId }
  | { type: 'openArtifact'; artifactId?: string }

function reduceWorkspaceShellState(state: WorkspaceShellSnapshot, action: WorkspaceShellAction): WorkspaceShellSnapshot {
  switch (action.type) {
    case 'setDefaultWorkspaceMode':
      return { ...state, defaultWorkspaceMode: action.defaultWorkspaceMode }
    case 'setLocale':
      return { ...state, locale: action.locale }
    case 'setProviderDraftSettings':
      return { ...state, providerDraftSettings: action.providerDraftSettings }
    case 'setSettingsCategory':
      return { ...state, settingsCategoryId: action.settingsCategoryId }
    case 'openSettings':
      return {
        ...state,
        settingsOpen: true,
        settingsCategoryId: action.settingsCategoryId ?? 'general',
      }
    case 'closeSettings':
      return { ...state, settingsOpen: false }
    case 'setSidebarWidth':
      return { ...state, sidebarWidth: action.sidebarWidth }
    case 'setSidebarCollapsed':
      return { ...state, sidebarCollapsed: action.sidebarCollapsed }
    case 'toggleTheme':
      return { ...state, theme: state.theme === 'dark' ? 'light' : 'dark' }
    case 'togglePreviewPanel':
      return {
        ...state,
        selectedRightPanelId: 'preview',
        rightPanelOpen: !state.rightPanelOpen,
        rightPanelMenuOpen: false,
      }
    case 'openRightPanel':
      return {
        ...state,
        selectedRightPanelId: action.selectedRightPanelId,
        rightPanelOpen: true,
        rightPanelMenuOpen: false,
      }
    case 'openArtifact':
      return {
        ...state,
        previewArtifactId: action.artifactId,
        selectedRightPanelId: 'preview',
        rightPanelOpen: true,
        rightPanelMenuOpen: false,
      }
  }
}

function bindWorkspaceShellActions(getState: () => WorkspaceShellSnapshot, dispatch: (action: WorkspaceShellAction) => void) {
  return {
    snapshot: getState,
    setDefaultWorkspaceMode: (defaultWorkspaceMode: Thread['mode']) => dispatch({ type: 'setDefaultWorkspaceMode', defaultWorkspaceMode }),
    setLocale: (locale: Locale) => dispatch({ type: 'setLocale', locale }),
    setProviderDraftSettings: (providerDraftSettings: ProviderDraftSettings) => dispatch({ type: 'setProviderDraftSettings', providerDraftSettings }),
    setSettingsCategory: (settingsCategoryId: SettingsCategoryId) => dispatch({ type: 'setSettingsCategory', settingsCategoryId }),
    openSettings: (settingsCategoryId?: SettingsCategoryId) => dispatch({ type: 'openSettings', settingsCategoryId }),
    closeSettings: () => dispatch({ type: 'closeSettings' }),
    setSidebarWidth: (sidebarWidth: number) => dispatch({ type: 'setSidebarWidth', sidebarWidth }),
    setSidebarCollapsed: (sidebarCollapsed: boolean) => dispatch({ type: 'setSidebarCollapsed', sidebarCollapsed }),
    toggleTheme: () => dispatch({ type: 'toggleTheme' }),
    togglePreviewPanel: () => dispatch({ type: 'togglePreviewPanel' }),
    openRightPanel: (selectedRightPanelId: RightPanelItemId) => dispatch({ type: 'openRightPanel', selectedRightPanelId }),
    openArtifact: (artifactId?: string) => dispatch({ type: 'openArtifact', artifactId }),
  }
}

export function createWorkspaceShellState(initialState: Partial<WorkspaceShellSnapshot> = {}) {
  let state = createInitialShellSnapshot(initialState)

  return bindWorkspaceShellActions(
    () => state,
    (action) => {
      state = reduceWorkspaceShellState(state, action)
    },
  )
}

export function useWorkspaceShellState() {
  const [state, setState] = useState(() => createInitialShellSnapshot())
  const actions = useMemo(() => bindWorkspaceShellActions(
    () => state,
    (action) => setState((current) => {
      const next = reduceWorkspaceShellState(current, action)
      if (action.type === 'toggleTheme') writeStoredTheme(next.theme)
      return next
    }),
  ), [state])

  useEffect(() => {
    if (readStoredTheme()) return undefined
    const mediaQuery = window.matchMedia?.('(prefers-color-scheme: dark)')
    if (!mediaQuery) return undefined

    const handleChange = (event: MediaQueryListEvent) => {
      setState((current) => ({ ...current, theme: event.matches ? 'dark' : 'light' }))
    }

    mediaQuery.addEventListener('change', handleChange)
    return () => mediaQuery.removeEventListener('change', handleChange)
  }, [])

  return {
    ...state,
    ...actions,
  }
}
