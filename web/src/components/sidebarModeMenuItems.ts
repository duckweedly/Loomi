export type SidebarMode = 'chat' | 'work'

export type SidebarModeMenuItem = {
  id: 'new-chat' | 'projects' | 'scheduled'
  label: string
  action?: 'create-thread'
}

export function createSidebarModeMenuItems(mode: SidebarMode, copy = { newChat: 'New Chat', projects: 'Projects', scheduled: 'Scheduled' }): SidebarModeMenuItem[] {
  if (mode === 'chat') {
    return [{ id: 'new-chat', label: copy.newChat, action: 'create-thread' }]
  }

  return [
    { id: 'projects', label: copy.projects },
    { id: 'scheduled', label: copy.scheduled },
  ]
}
