#!/usr/bin/env node

const crypto = require('node:crypto');
const fs = require('node:fs');
const http = require('node:http');
const https = require('node:https');
const path = require('node:path');
const {
  assetName,
  baseUrl,
  checksumsName,
  ensureDir,
  installedBinaryPath,
  manifestPath,
  packageVersion,
  resolveTarget,
} = require('./native');

const GO_INSTALL_COMMAND = 'go install github.com/Blu3Ph4ntom/warden-mcp/cmd/warden-mcp@latest';

function requestBuffer(url, redirects = 0) {
  return new Promise((resolve, reject) => {
    const client = url.startsWith('https:') ? https : http;
    client
      .get(url, { headers: { 'user-agent': 'warden-mcp-installer' } }, (response) => {
        if ([301, 302, 303, 307, 308].includes(response.statusCode) && response.headers.location) {
          response.resume();
          if (redirects >= 5) return reject(new Error(`Too many redirects while fetching ${url}`));
          return resolve(requestBuffer(new URL(response.headers.location, url).toString(), redirects + 1));
        }
        if (response.statusCode !== 200) {
          const chunks = [];
          response.on('data', (chunk) => chunks.push(chunk));
          response.on('end', () => reject(new Error(`GET ${url} failed with ${response.statusCode}: ${Buffer.concat(chunks).toString('utf8').trim()}`)));
          return;
        }
        const chunks = [];
        response.on('data', (chunk) => chunks.push(chunk));
        response.on('end', () => resolve(Buffer.concat(chunks)));
      })
      .on('error', reject);
  });
}

function sha256(buffer) {
  return crypto.createHash('sha256').update(buffer).digest('hex');
}

function parseChecksums(content) {
  const entries = new Map();
  for (const line of content.split(/\r?\n/)) {
    const trimmed = line.trim();
    if (!trimmed) continue;
    const [hash, fileName] = trimmed.split(/\s+/, 2);
    if (hash && fileName) entries.set(fileName.trim(), hash.trim().toLowerCase());
  }
  return entries;
}

async function main() {
  if (process.env.WARDEN_MCP_SKIP_DOWNLOAD === '1') {
    console.error('[warden-mcp] skipping native download because WARDEN_MCP_SKIP_DOWNLOAD=1');
    return;
  }

  const version = packageVersion();
  const target = resolveTarget();
  const asset = assetName(version, target.platform, target.arch);
  const destination = installedBinaryPath({ platform: target.platform });
  const installDir = path.dirname(destination);
  const assetUrl = process.env.WARDEN_MCP_BINARY_URL || `${baseUrl(version)}/${asset}`;
  const sumsUrl = process.env.WARDEN_MCP_CHECKSUMS_URL || `${baseUrl(version)}/${checksumsName(version)}`;

  const checksumEntries = parseChecksums((await requestBuffer(sumsUrl)).toString('utf8'));
  const expectedHash = checksumEntries.get(asset);
  if (!expectedHash) throw new Error(`No checksum entry found for ${asset} in ${sumsUrl}`);

  const binary = await requestBuffer(assetUrl);
  const actualHash = sha256(binary);
  if (actualHash !== expectedHash) {
    throw new Error(`Checksum mismatch for ${asset}: expected ${expectedHash}, got ${actualHash}`);
  }

  ensureDir(installDir);
  fs.writeFileSync(destination, binary, target.platform === 'win32' ? undefined : { mode: 0o755 });
  if (target.platform !== 'win32') fs.chmodSync(destination, 0o755);
  fs.writeFileSync(manifestPath(installDir), `${JSON.stringify({ version, asset, sha256: actualHash, url: assetUrl }, null, 2)}\n`);
  console.error(`[warden-mcp] installed ${asset} -> ${destination}`);
}

main().catch((error) => {
  console.error(`[warden-mcp] native install failed: ${error.message}`);
  console.error(`[warden-mcp] Retry with npm rebuild warden-mcp or install via Go: ${GO_INSTALL_COMMAND}`);
  process.exit(1);
});

