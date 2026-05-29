const { app, BrowserWindow, dialog, ipcMain, nativeImage, shell } = require('electron')
const fs = require('node:fs/promises')
const os = require('node:os')
const path = require('node:path')

const startUrl = process.env.ELECTRON_START_URL
const appName = 'Loomi'
const macIconPath = path.join(__dirname, '..', 'public', 'icons', 'loomi-dock.png')
const iconPath = process.platform === 'darwin'
  ? macIconPath
  : path.join(__dirname, '..', 'public', 'icons', 'loomi.ico')

function safeArtifactFilename(title, mimeType) {
  const base = String(title || 'Loomi artifact')
    .replace(/[<>:"/\\|?*\u0000-\u001f]/g, ' ')
    .replace(/\s+/g, ' ')
    .trim()
    .slice(0, 80) || 'Loomi artifact'
  const lower = base.toLowerCase()
  if (lower.endsWith('.svg') || lower.endsWith('.html') || lower.endsWith('.htm') || lower.endsWith('.md') || lower.endsWith('.txt')) return base
  if (mimeType === 'image/svg+xml') return `${base}.svg`
  if (mimeType === 'text/html') return `${base}.html`
  if (mimeType === 'text/plain') return `${base}.txt`
  return `${base}.md`
}

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

  ipcMain.handle('loomi:open-artifact-file', async (_event, input) => {
    const content = typeof input?.content === 'string' ? input.content : ''
    if (!content) return { ok: false, error: 'Artifact content is empty.' }
    const dir = path.join(os.tmpdir(), 'loomi-artifacts')
    await fs.mkdir(dir, { recursive: true })
    const filename = safeArtifactFilename(input?.filename || input?.title, input?.mimeType)
    const filePath = path.join(dir, filename)
    await fs.writeFile(filePath, content, 'utf8')
    const error = await shell.openPath(filePath)
    if (error) return { ok: false, error }
    return { ok: true, path: filePath }
  })

  app.on('activate', () => {
    if (BrowserWindow.getAllWindows().length === 0) createWindow()
  })
})

app.on('window-all-closed', () => {
  if (process.platform !== 'darwin') app.quit()
})
