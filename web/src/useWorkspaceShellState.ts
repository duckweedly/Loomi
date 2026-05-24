import { useState } from 'react'
import type { Thread } from './domain'
import type { Locale } from './i18n'
import type { SettingsCategoryId } from './components/settingsCatalog'
import type { RightPanelItemId } from './rightPanelItems'

export type ProviderDraftSettings = {
  baseUrl: string
  model: string
  apiKeySet: boolean
}

type WorkspaceShellSnapshot = {
  defaultWorkspaceMode: Thread['mode']
  locale: Locale
  providerDraftSettings: ProviderDraftSettings
  runDetailsOpen: boolean
  rightPanelMenuOpen: boolean
  rightPanelOpen: boolean
  selectedRightPanelId: RightPanelItemId
  settingsCategoryId: SettingsCategoryId
  settingsOpen: boolean
  sidebarCollapsed: boolean
  sidebarWidth: number
  theme: 'dark' | 'light'
}

const initialShellState: WorkspaceShellSnapshot = {
  defaultWorkspaceMode: 'chat',
  locale: 'zh',
  providerDraftSettings: { baseUrl: '', model: '', apiKeySet: false },
  runDetailsOpen: false,
  rightPanelMenuOpen: false,
  rightPanelOpen: false,
  selectedRightPanelId: 'preview',
  settingsCategoryId: 'general',
  settingsOpen: false,
  sidebarCollapsed: false,
  sidebarWidth: 292,
  theme: 'dark',
}

type WorkspaceShellAction =
  | { type: 'setDefaultWorkspaceMode'; defaultWorkspaceMode: Thread['mode'] }
  | { type: 'setLocale'; locale: Locale }
  | { type: 'setProviderDraftSettings'; providerDraftSettings: ProviderDraftSettings }
  | { type: 'setSettingsCategory'; settingsCategoryId: SettingsCategoryId }
  | { type: 'openSettings' }
  | { type: 'closeSettings' }
  | { type: 'setSidebarWidth'; sidebarWidth: number }
  | { type: 'setSidebarCollapsed'; sidebarCollapsed: boolean }
  | { type: 'toggleTheme' }
  | { type: 'toggleRunDetails' }
  | { type: 'toggleRightPanelMenu' }
  | { type: 'openRightPanel'; selectedRightPanelId: RightPanelItemId }
  | { type: 'closeRunDetails' }
  | { type: 'openArtifact' }

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
        settingsCategoryId: 'general',
      }
    case 'closeSettings':
      return { ...state, settingsOpen: false }
    case 'setSidebarWidth':
      return { ...state, sidebarWidth: action.sidebarWidth }
    case 'setSidebarCollapsed':
      return { ...state, sidebarCollapsed: action.sidebarCollapsed }
    case 'toggleTheme':
      return { ...state, theme: state.theme === 'dark' ? 'light' : 'dark' }
    case 'toggleRunDetails':
      return {
        ...state,
        runDetailsOpen: !state.runDetailsOpen,
        rightPanelMenuOpen: false,
        rightPanelOpen: false,
      }
    case 'toggleRightPanelMenu':
      if (state.rightPanelOpen) {
        return { ...state, rightPanelOpen: false, rightPanelMenuOpen: false }
      }
      return {
        ...state,
        rightPanelMenuOpen: !state.rightPanelMenuOpen,
        runDetailsOpen: false,
      }
    case 'openRightPanel':
      return {
        ...state,
        selectedRightPanelId: action.selectedRightPanelId,
        rightPanelOpen: true,
        rightPanelMenuOpen: false,
        runDetailsOpen: false,
      }
    case 'closeRunDetails':
      return { ...state, runDetailsOpen: false }
    case 'openArtifact':
      return {
        ...state,
        selectedRightPanelId: 'preview',
        rightPanelOpen: true,
        rightPanelMenuOpen: false,
        runDetailsOpen: false,
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
    openSettings: () => dispatch({ type: 'openSettings' }),
    closeSettings: () => dispatch({ type: 'closeSettings' }),
    setSidebarWidth: (sidebarWidth: number) => dispatch({ type: 'setSidebarWidth', sidebarWidth }),
    setSidebarCollapsed: (sidebarCollapsed: boolean) => dispatch({ type: 'setSidebarCollapsed', sidebarCollapsed }),
    toggleTheme: () => dispatch({ type: 'toggleTheme' }),
    toggleRunDetails: () => dispatch({ type: 'toggleRunDetails' }),
    toggleRightPanelMenu: () => dispatch({ type: 'toggleRightPanelMenu' }),
    openRightPanel: (selectedRightPanelId: RightPanelItemId) => dispatch({ type: 'openRightPanel', selectedRightPanelId }),
    closeRunDetails: () => dispatch({ type: 'closeRunDetails' }),
    openArtifact: () => dispatch({ type: 'openArtifact' }),
  }
}

export function createWorkspaceShellState(initialState: Partial<WorkspaceShellSnapshot> = {}) {
  let state = { ...initialShellState, ...initialState }

  return bindWorkspaceShellActions(
    () => state,
    (action) => {
      state = reduceWorkspaceShellState(state, action)
    },
  )
}

export function useWorkspaceShellState() {
  const [state, setState] = useState(initialShellState)
  const actions = bindWorkspaceShellActions(
    () => state,
    (action) => setState((current) => reduceWorkspaceShellState(current, action)),
  )

  return {
    ...state,
    ...actions,
  }
}
