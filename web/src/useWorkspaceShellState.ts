import { useState } from 'react'
import type { RightPanelItemId } from './rightPanelItems'

type WorkspaceShellSnapshot = {
  runDetailsOpen: boolean
  rightPanelMenuOpen: boolean
  rightPanelOpen: boolean
  selectedRightPanelId: RightPanelItemId
  sidebarCollapsed: boolean
  sidebarWidth: number
  theme: 'dark' | 'light'
}

const initialShellState: WorkspaceShellSnapshot = {
  runDetailsOpen: false,
  rightPanelMenuOpen: false,
  rightPanelOpen: false,
  selectedRightPanelId: 'preview',
  sidebarCollapsed: false,
  sidebarWidth: 292,
  theme: 'dark',
}

type WorkspaceShellAction =
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
