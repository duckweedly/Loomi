const { contextBridge, ipcRenderer } = require('electron')

window.addEventListener('DOMContentLoaded', () => {
  document.documentElement.dataset.runtime = 'electron'
})

contextBridge.exposeInMainWorld('loomiDesktop', {
  selectWorkspaceFolder: () => ipcRenderer.invoke('loomi:select-workspace-folder'),
})
