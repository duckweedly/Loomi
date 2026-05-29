const { contextBridge, ipcRenderer } = require('electron')

window.addEventListener('DOMContentLoaded', () => {
  document.documentElement.dataset.runtime = 'electron'
})

contextBridge.exposeInMainWorld('loomiDesktop', {
  selectWorkspaceFolder: () => ipcRenderer.invoke('loomi:select-workspace-folder'),
  openArtifactFile: (artifact) => ipcRenderer.invoke('loomi:open-artifact-file', artifact),
})
