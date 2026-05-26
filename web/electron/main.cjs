const { app, BrowserWindow, dialog, ipcMain, nativeImage } = require('electron')
const path = require('node:path')

const startUrl = process.env.ELECTRON_START_URL
const appName = 'Loomi'
const macIconPath = path.join(__dirname, '..', 'public', 'icons', 'loomi-dock.png')
const iconPath = process.platform === 'darwin'
  ? macIconPath
  : path.join(__dirname, '..', 'public', 'icons', 'loomi.ico')

function createWindow() {
  const mainWindow = new BrowserWindow({
    width: 1280,
    height: 860,
    minWidth: 1120,
    minHeight: 720,
    backgroundColor: '#070810',
    title: appName,
    icon: iconPath,
    titleBarStyle: 'hiddenInset',
    trafficLightPosition: { x: 14, y: 12 },
    webPreferences: {
      contextIsolation: true,
      nodeIntegration: false,
      preload: path.join(__dirname, 'preload.cjs'),
    },
  })

  if (startUrl) {
    void mainWindow.loadURL(startUrl)
  } else {
    void mainWindow.loadFile(path.join(__dirname, '..', 'dist', 'index.html'))
  }
}

app.whenReady().then(() => {
  app.setName(appName)
  if (process.platform === 'darwin') {
    app.dock.setIcon(nativeImage.createFromPath(macIconPath))
  }

  createWindow()

  ipcMain.handle('loomi:select-workspace-folder', async () => {
    const window = BrowserWindow.getFocusedWindow() || BrowserWindow.getAllWindows()[0]
    const result = await dialog.showOpenDialog(window, {
      properties: ['openDirectory'],
      title: '选择 Loomi 可访问的目录',
    })
    if (result.canceled || result.filePaths.length === 0) return { canceled: true }
    return { canceled: false, path: result.filePaths[0] }
  })

  app.on('activate', () => {
    if (BrowserWindow.getAllWindows().length === 0) createWindow()
  })
})

app.on('window-all-closed', () => {
  if (process.platform !== 'darwin') app.quit()
})
