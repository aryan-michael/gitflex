const os = require('os');
const path = require('path');
const fs = require('fs');
const execSync = require('child_process').execSync;

const platform = os.platform();
const arch = os.arch();
const binDir = path.resolve(__dirname, '../bin');
let binaryName = 'gitflux';

if (platform === 'win32') {
  binaryName += arch === 'arm64' ? '-windows-arm64.exe' : '-windows-amd64.exe';
} else if (platform === 'darwin') {
  binaryName += arch === 'arm64' ? '-macos-arm64' : '-macos';
} else if (platform === 'linux') {
  binaryName += arch === 'arm64' ? '-linux-arm64' : '-linux';
}

const binaryPath = path.join(binDir, binaryName);
const targetPath = path.join(binDir, 'gitflux');

// Copy the correct binary to the target path
fs.copyFileSync(binaryPath, targetPath);

if (platform === 'win32') {
  try {
    // Create a batch script to run the executable
    const batchScriptPath = path.join(binDir, 'gitflux.cmd');
    const batchScriptContent = `@echo off\n"${binaryPath}" %*`;
    fs.writeFileSync(batchScriptPath, batchScriptContent);

    // Add the bin directory to PATH
    const command = `setx PATH "%PATH%;${binDir}"`;
    execSync(command, { stdio: 'inherit' });
    console.log('gitflux has been added to your PATH. Please restart your terminal.');
  } catch (error) {
    console.error('Error during post-installation setup:', error.message);
  }
} else {
  const shell = process.env.SHELL || '/bin/bash';
  const rcFile = shell.includes('zsh') ? '.zshrc' : '.bashrc';
  const profilePath = path.join(os.homedir(), rcFile);

  fs.appendFileSync(profilePath, `\nexport PATH="$PATH:${binDir}"\n`);
  console.log(`gitflux has been added to your PATH. Please run 'source ~/${rcFile}' or restart your terminal.`);
}
