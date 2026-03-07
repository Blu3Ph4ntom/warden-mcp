#!/usr/bin/env node

const fs = require('node:fs');
const { spawn } = require('node:child_process');
const { assetName, installedBinaryPath, packageVersion } = require('./native');

function printInstallHelp(binaryPath) {
  const expectedAsset = assetName(packageVersion());
  console.error(
    [
      'warden-mcp could not find its npm-installed native binary.',
      '',
      `Expected installed binary: ${binaryPath}`,
      `Expected release asset: ${expectedAsset}`,
      '',
      'Try one of these recovery steps:',
      '  npm rebuild warden-mcp',
      '  npm install -g warden-mcp --force',
      '',
      'Fallback native install:',
      '  go install github.com/Blu3Ph4ntom/warden-mcp/cmd/warden-mcp@latest',
    ].join('\n')
  );
  process.exit(1);
}

function run(binaryPath) {
  const child = spawn(binaryPath, process.argv.slice(2), { stdio: 'inherit' });
  child.on('error', (error) => {
    console.error(`Failed to start native Warden MCP binary at ${binaryPath}: ${error.message}`);
    process.exit(1);
  });
  child.on('exit', (code, signal) => {
    if (signal) {
      process.kill(process.pid, signal);
      return;
    }
    process.exit(code ?? 0);
  });
}

const binaryPath = installedBinaryPath();
if (!fs.existsSync(binaryPath)) {
  printInstallHelp(binaryPath);
}

run(binaryPath);