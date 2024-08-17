const os = require('os');
const fs = require('fs');
const path = require('path');
const { execSync } = require('child_process');

const platform = os.platform();
const arch = os.arch();
let binaryName = 'gitwizard';
let outputDir = 'bin';

if (!fs.existsSync(outputDir)) {
  fs.mkdirSync(outputDir);
}

let outputPath = path.join(__dirname, outputDir, binaryName);

if (platform === 'win32') {
  outputPath += '.exe';
}

// Set the appropriate build command based on platform and architecture
let buildCmd = '';

if (platform === 'darwin') {
  if (arch === 'arm64') {
    buildCmd = `GOOS=darwin GOARCH=arm64 go build -o ${outputPath} main.go`;
  } else if (arch === 'x64' || arch === 'amd64') {
    buildCmd = `GOOS=darwin GOARCH=amd64 go build -o ${outputPath} main.go`;
  }
} else if (platform === 'linux') {
  if (arch === 'x64' || arch === 'amd64') {
    buildCmd = `GOOS=linux GOARCH=amd64 go build -o ${outputPath} main.go`;
  }
} else if (platform === 'win32') {
  buildCmd = `GOOS=windows GOARCH=amd64 go build -o ${outputPath} main.go`;
}

// Execute the build command
try {
  execSync(buildCmd, { stdio: 'inherit' });
  fs.chmodSync(outputPath, '755');
  console.log(`Successfully built ${binaryName} for ${platform}-${arch}`);
} catch (error) {
  console.error(`Failed to build ${binaryName}: ${error.message}`);
  process.exit(1);
}
