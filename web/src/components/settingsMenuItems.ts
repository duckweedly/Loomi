export type SettingsMenuItemId = 'settings' | 'theme' | 'update'

export type SettingsMenuItem = {
  id: SettingsMenuItemId
  label: string
  value?: string
}

export function createSettingsMenuItems(theme: 'dark' | 'light', copy = { settings: 'Settings', theme: 'Theme', update: 'Update', current: 'Current', dark: 'Dark', light: 'Light' }): SettingsMenuItem[] {
  return [
    { id: 'settings', label: copy.settings },
    { id: 'theme', label: copy.theme, value: theme === 'dark' ? copy.dark : copy.light },
    { id: 'update', label: copy.update, value: copy.current },
  ]
}
