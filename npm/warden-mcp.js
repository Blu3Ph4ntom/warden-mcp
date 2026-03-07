#!/usr/bin/env node

const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');
const { spawn } = require('node:child_process');

const BINARY_NAME = process.platform === 'win32' ? 'warden-mcp.exe' : 'warden-mcp';
const INSTALL_COMMAND = 'go install github.com/Blu3Ph4ntom/warden-mcp/cmd/warden-mcp@latest';

function unique(items) {
  return [...new Set(items.filter(Boolean))];
}

function candidateDirs() {
  const dirs = [];
  if (process.env.WARDEN_MCP_NATIVE_PATH) {
    dirs.push(path.dirname(process.env.WARDEN_MCP_NATIVE_PATH));
  }
  if (process.env.GOBIN) {
    dirs.push(process.env.GOBIN);
  }
  if (process.env.GOPATH) {
    dirs.push(...process.env.GOPATH.split(path.delimiter).map((entry) => path.join(entry, 'bin')));
  }
  dirs.push(path.join(os.homedir(), 'go', 'bin'));
  return unique(dirs);
}

function candidateBinaries() {
  const binaries = [];
  if (process.env.WARDEN_MCP_NATIVE_PATH) {
    binaries.push(process.env.WARDEN_MCP_NATIVE_PATH);
  }
  for (const dir of candidateDirs()) {
    binaries.push(path.join(dir, BINARY_NAME));
  }
  return unique(binaries);
}

function findNativeBinary() {
  return candidateBinaries().find((candidate) => {
    try {
      return fs.existsSync(candidate) && fs.statSync(candidate).isFile();
    } catch {
      return false;
    }
  });
}

function printSetupHelp() {
  const lookedIn = candidateBinaries().map((candidate) => `  - ${candidate}`).join('\n');
  console.error(
    [
      'warden-mcp npm launcher could not find the native Warden MCP binary.',
      '',
      'Install Go 1.24+ and then run:',
      `  ${INSTALL_COMMAND}`,
      '',
      'Make sure your Go bin directory is on PATH, then retry `warden-mcp`.',
      '',
      'Searched:',
      lookedIn,
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

const binaryPath = findNativeBinary();
if (!binaryPath) {
  printSetupHelp();
}

run(binaryPath);