export type SidebarMode = 'chat' | 'work'

export type SidebarModeMenuItem = {
  id: 'new-chat' | 'projects' | 'scheduled'
  label: string
  action?: 'create-thread'
}

export function createSidebarModeMenuItems(mode: SidebarMode): SidebarModeMenuItem[] {
  if (mode === 'chat') {
    return [{ id: 'new-chat', label: 'New Chat', action: 'create-thread' }]
  }

  return [
    { id: 'projects', label: 'Projects' },
    { id: 'scheduled', label: 'Scheduled' },
  ]
}
