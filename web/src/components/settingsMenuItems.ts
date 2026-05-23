export type SettingsMenuItemId = 'settings' | 'theme' | 'update'

export type SettingsMenuItem = {
  id: SettingsMenuItemId
  label: string
  value?: string
}

export function createSettingsMenuItems(theme: 'dark' | 'light'): SettingsMenuItem[] {
  return [
    { id: 'settings', label: 'Settings' },
    { id: 'theme', label: 'Theme', value: theme === 'dark' ? 'Dark' : 'Light' },
    { id: 'update', label: 'Update', value: 'Current' },
  ]
}
