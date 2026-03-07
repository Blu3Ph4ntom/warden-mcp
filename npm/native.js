const fs = require('node:fs');
const path = require('node:path');
const pkg = require('../package.json');

const TARGETS = {
  win32: { x64: 'windows_amd64', arm64: 'windows_arm64' },
  darwin: { x64: 'darwin_amd64', arm64: 'darwin_arm64' },
  linux: { x64: 'linux_amd64', arm64: 'linux_arm64' },
};

function packageVersion() {
  return process.env.WARDEN_MCP_VERSION || process.env.npm_package_version || pkg.version;
}

function binaryName(platform = process.platform) {
  return platform === 'win32' ? 'warden-mcp.exe' : 'warden-mcp';
}

function resolveTarget(platform = process.platform, arch = process.arch) {
  const suffix = TARGETS[platform]?.[arch];
  if (!suffix) {
    throw new Error(`Unsupported platform/arch for Warden MCP native install: ${platform}/${arch}`);
  }
  return { platform, arch, suffix, binaryName: binaryName(platform) };
}

function assetName(version = packageVersion(), platform = process.platform, arch = process.arch) {
  const target = resolveTarget(platform, arch);
  return `warden-mcp_${version}_${target.suffix}${platform === 'win32' ? '.exe' : ''}`;
}

function checksumsName(version = packageVersion()) {
  return `warden-mcp_${version}_checksums.txt`;
}

function baseUrl(version = packageVersion()) {
  return process.env.WARDEN_MCP_DIST_BASE_URL || `https://github.com/Blu3Ph4ntom/warden-mcp/releases/download/v${version}`;
}

function packageRoot() {
  return path.resolve(__dirname, '..');
}

function defaultInstallDir() {
  return process.env.WARDEN_MCP_INSTALL_DIR || path.join(packageRoot(), 'npm', 'native');
}

function installedBinaryPath({ platform = process.platform, installDir = defaultInstallDir() } = {}) {
  return process.env.WARDEN_MCP_NATIVE_PATH || path.join(installDir, binaryName(platform));
}

function manifestPath(installDir = defaultInstallDir()) {
  return path.join(installDir, 'manifest.json');
}

function ensureDir(dirPath) {
  fs.mkdirSync(dirPath, { recursive: true });
}

module.exports = {
  assetName,
  baseUrl,
  binaryName,
  checksumsName,
  defaultInstallDir,
  ensureDir,
  installedBinaryPath,
  manifestPath,
  packageRoot,
  packageVersion,
  resolveTarget,
};

