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
  rightPanelWidth: number
  previewArtifactId?: string
  selectedRightPanelId: RightPanelItemId
  settingsCategoryId: SettingsCategoryId
  settingsOpen: boolean
  sidebarCollapsed: boolean
  sidebarWidth: number
  theme: 'dark' | 'light'
  themePreference: ThemePreference
}

type Theme = WorkspaceShellSnapshot['theme']
export type ThemePreference = 'dark' | 'light' | 'system'

const themeStorageKey = 'loomi.theme'
const sidebarWidthStorageKey = 'loomi.sidebarWidth'
const rightPanelWidthStorageKey = 'loomi.rightPanelWidth'
export const defaultSidebarWidth = 260
export const sidebarMinWidth = 220
export const sidebarMaxWidth = 340
export const defaultRightPanelWidth = 430
export const rightPanelMinWidth = 360
export const rightPanelMaxWidth = 720

export function clampSidebarWidth(sidebarWidth: number) {
  return Math.min(sidebarMaxWidth, Math.max(sidebarMinWidth, Math.round(sidebarWidth)))
}

export function clampRightPanelWidth(rightPanelWidth: number) {
  return Math.min(rightPanelMaxWidth, Math.max(rightPanelMinWidth, Math.round(rightPanelWidth)))
}

function isLegacyDefaultSidebarWidth(sidebarWidth: number) {
  return [128, 136, 140, 148, 156, 168, 172, 184, 188, 196, 204, 216, 224, 236, 248, 264].includes(Math.round(sidebarWidth))
}

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

function clearStoredTheme() {
  if (typeof window === 'undefined') return
  try {
    window.localStorage.removeItem(themeStorageKey)
  } catch {
    // localStorage may be unavailable in privacy-restricted desktop contexts.
  }
}

function readStoredSidebarWidth() {
  if (typeof window === 'undefined') return null
  try {
    const storedValue = window.localStorage.getItem(sidebarWidthStorageKey)
    if (storedValue === null) return null
    const storedWidth = Number(storedValue)
    if (!Number.isFinite(storedWidth)) return null
    if (isLegacyDefaultSidebarWidth(storedWidth)) return null
    return clampSidebarWidth(storedWidth)
  } catch {
    return null
  }
}

function readStoredRightPanelWidth() {
  if (typeof window === 'undefined') return null
  try {
    const storedValue = window.localStorage.getItem(rightPanelWidthStorageKey)
    if (storedValue === null) return null
    const storedWidth = Number(storedValue)
    return Number.isFinite(storedWidth) ? clampRightPanelWidth(storedWidth) : null
  } catch {
    return null
  }
}

function writeStoredSidebarWidth(sidebarWidth: number) {
  if (typeof window === 'undefined') return
  try {
    window.localStorage.setItem(sidebarWidthStorageKey, String(clampSidebarWidth(sidebarWidth)))
  } catch {
    // localStorage may be unavailable in privacy-restricted desktop contexts.
  }
}

function writeStoredRightPanelWidth(rightPanelWidth: number) {
  if (typeof window === 'undefined') return
  try {
    window.localStorage.setItem(rightPanelWidthStorageKey, String(clampRightPanelWidth(rightPanelWidth)))
  } catch {
    // localStorage may be unavailable in privacy-restricted desktop contexts.
  }
}

export function resolveSystemTheme(): Theme {
  if (typeof window === 'undefined' || !window.matchMedia) return 'light'
  return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
}

function createInitialShellSnapshot(initialState: Partial<WorkspaceShellSnapshot> = {}): WorkspaceShellSnapshot {
  const storedTheme = readStoredTheme()
  return {
    ...baseShellState,
    sidebarWidth: readStoredSidebarWidth() ?? baseShellState.sidebarWidth,
    rightPanelWidth: readStoredRightPanelWidth() ?? baseShellState.rightPanelWidth,
    theme: storedTheme ?? resolveSystemTheme(),
    themePreference: storedTheme ?? 'system',
    ...initialState,
  }
}

const baseShellState: WorkspaceShellSnapshot = {
  defaultWorkspaceMode: 'chat',
  locale: 'zh',
  providerDraftSettings: { baseUrl: '', model: '', apiKey: '', apiKeySet: false },
  rightPanelMenuOpen: false,
  rightPanelOpen: false,
  rightPanelWidth: defaultRightPanelWidth,
  selectedRightPanelId: 'preview',
  settingsCategoryId: 'general',
  settingsOpen: false,
  sidebarCollapsed: false,
  sidebarWidth: defaultSidebarWidth,
  theme: 'light',
  themePreference: 'system',
}

type WorkspaceShellAction =
  | { type: 'setDefaultWorkspaceMode'; defaultWorkspaceMode: Thread['mode'] }
  | { type: 'setLocale'; locale: Locale }
  | { type: 'setProviderDraftSettings'; providerDraftSettings: ProviderDraftSettings }
  | { type: 'setSettingsCategory'; settingsCategoryId: SettingsCategoryId }
  | { type: 'openSettings'; settingsCategoryId?: SettingsCategoryId }
  | { type: 'closeSettings' }
  | { type: 'setSidebarWidth'; sidebarWidth: number }
  | { type: 'setRightPanelWidth'; rightPanelWidth: number }
  | { type: 'setSidebarCollapsed'; sidebarCollapsed: boolean }
  | { type: 'setThemePreference'; themePreference: ThemePreference }
  | { type: 'toggleTheme' }
  | { type: 'togglePreviewPanel' }
  | { type: 'closeRightPanel' }
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
      return { ...state, sidebarWidth: clampSidebarWidth(action.sidebarWidth) }
    case 'setRightPanelWidth':
      return { ...state, rightPanelWidth: clampRightPanelWidth(action.rightPanelWidth) }
    case 'setSidebarCollapsed':
      return { ...state, sidebarCollapsed: action.sidebarCollapsed }
    case 'setThemePreference':
      return {
        ...state,
        themePreference: action.themePreference,
        theme: action.themePreference === 'system' ? resolveSystemTheme() : action.themePreference,
      }
    case 'toggleTheme':
      return {
        ...state,
        theme: state.theme === 'dark' ? 'light' : 'dark',
        themePreference: state.theme === 'dark' ? 'light' : 'dark',
      }
    case 'togglePreviewPanel':
      return {
        ...state,
        selectedRightPanelId: 'preview',
        rightPanelOpen: !state.rightPanelOpen,
        rightPanelMenuOpen: false,
      }
    case 'closeRightPanel':
      return {
        ...state,
        rightPanelOpen: false,
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
    setRightPanelWidth: (rightPanelWidth: number) => dispatch({ type: 'setRightPanelWidth', rightPanelWidth }),
    setSidebarCollapsed: (sidebarCollapsed: boolean) => dispatch({ type: 'setSidebarCollapsed', sidebarCollapsed }),
    setThemePreference: (themePreference: ThemePreference) => dispatch({ type: 'setThemePreference', themePreference }),
    toggleTheme: () => dispatch({ type: 'toggleTheme' }),
    togglePreviewPanel: () => dispatch({ type: 'togglePreviewPanel' }),
    closeRightPanel: () => dispatch({ type: 'closeRightPanel' }),
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
      if (action.type === 'setSidebarWidth') writeStoredSidebarWidth(next.sidebarWidth)
      if (action.type === 'setRightPanelWidth') writeStoredRightPanelWidth(next.rightPanelWidth)
      if (action.type === 'setThemePreference') {
        if (action.themePreference === 'system') clearStoredTheme()
        else writeStoredTheme(action.themePreference)
      }
      return next
    }),
  ), [state])

  useEffect(() => {
    const mediaQuery = window.matchMedia?.('(prefers-color-scheme: dark)')
    if (!mediaQuery) return undefined

    const handleChange = (event: MediaQueryListEvent) => {
      setState((current) => (
        current.themePreference === 'system'
          ? { ...current, theme: event.matches ? 'dark' : 'light' }
          : current
      ))
    }

    mediaQuery.addEventListener('change', handleChange)
    return () => mediaQuery.removeEventListener('change', handleChange)
  }, [])

  return {
    ...state,
    ...actions,
  }
}
