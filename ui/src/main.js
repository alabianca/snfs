const path = require('path');
const { app, BrowserWindow, ipcMain } = require('electron')


function createWindow () {
  // Create the browser window.
  let win = new BrowserWindow({
    width: 800,
    height: 600,
    frame: false,
    webPreferences: {
      nodeIntegration: true
    }
  })

  ipcMain.on('window:close', () => win.close());
  ipcMain.on('window:minimize', () => win.minimize());
  ipcMain.on('window:maximize', () => win.isMaximized() ? win.unmaximize() : win.maximize());

  win.webContents.openDevTools()

  // and load the index.html of the app.
  win.loadURL('http://localhost:3000')
  //win.loadURL(path.normalize('file://' + path.join(__dirname, 'build/index.html')))
}



app.whenReady().then(createWindow)