import assert from 'node:assert/strict';
import fs from 'node:fs';
import http from 'node:http';
import { spawn, spawnSync } from 'node:child_process';
import { mkdtempSync } from 'node:fs';
import { tmpdir } from 'node:os';
import path from 'node:path';
import test from 'node:test';
import { createRequire } from 'node:module';
import { fileURLToPath } from 'node:url';

const rootDir = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..');
const require = createRequire(import.meta.url);
const native = require(path.join(rootDir, 'npm', 'native.js'));
const installPath = path.join(rootDir, 'npm', 'install.js');
const wrapperPath = path.join(rootDir, 'npm', 'warden-mcp.js');

function runNode(scriptPath, env) {
  return new Promise((resolve, reject) => {
    const child = spawn(process.execPath, [scriptPath], { env, stdio: ['ignore', 'pipe', 'pipe'] });
    let stdout = '';
    let stderr = '';
    child.stdout.on('data', (chunk) => {
      stdout += chunk.toString();
    });
    child.stderr.on('data', (chunk) => {
      stderr += chunk.toString();
    });
    child.on('error', reject);
    child.on('close', (status) => resolve({ status, stdout, stderr }));
  });
}

test('downloads and verifies the native binary during install', async () => {
  const version = '0.1.1-test';
  const installDir = mkdtempSync(path.join(tmpdir(), 'warden-mcp-install-'));
  const asset = native.assetName(version);
  const manifest = path.join(installDir, 'manifest.json');
  const binary = Buffer.from('not-a-real-binary-but-good-enough-for-download-test');
  const hash = require('node:crypto').createHash('sha256').update(binary).digest('hex');
  const server = http.createServer((request, response) => {
    if (request.url?.endsWith(`/${native.checksumsName(version)}`)) return response.end(`${hash}  ${asset}\n`);
    if (request.url?.endsWith(`/${asset}`)) return response.end(binary);
    response.statusCode = 404;
    response.end('not found');
  });

  await new Promise((resolve) => server.listen(0, '127.0.0.1', resolve));
  const { port } = server.address();
  const result = await runNode(installPath, {
    ...process.env,
    WARDEN_MCP_VERSION: version,
    WARDEN_MCP_INSTALL_DIR: installDir,
    WARDEN_MCP_DIST_BASE_URL: `http://127.0.0.1:${port}/releases/download/v${version}`,
  });
  await new Promise((resolve, reject) => server.close((error) => (error ? reject(error) : resolve())));

  assert.equal(result.status, 0, result.stderr);
  assert.deepEqual(fs.readFileSync(path.join(installDir, native.binaryName())), binary);
  assert.match(fs.readFileSync(manifest, 'utf8'), /0\.1\.1-test/);
});

test('fails install when checksum validation fails', async () => {
  const version = '0.1.1-bad';
  const installDir = mkdtempSync(path.join(tmpdir(), 'warden-mcp-bad-'));
  const asset = native.assetName(version);
  const server = http.createServer((request, response) => {
    if (request.url?.endsWith(`/${native.checksumsName(version)}`)) return response.end(`deadbeef  ${asset}\n`);
    if (request.url?.endsWith(`/${asset}`)) return response.end('wrong payload');
    response.statusCode = 404;
    response.end('not found');
  });

  await new Promise((resolve) => server.listen(0, '127.0.0.1', resolve));
  const { port } = server.address();
  const result = await runNode(installPath, {
    ...process.env,
    WARDEN_MCP_VERSION: version,
    WARDEN_MCP_INSTALL_DIR: installDir,
    WARDEN_MCP_DIST_BASE_URL: `http://127.0.0.1:${port}/releases/download/v${version}`,
  });
  await new Promise((resolve, reject) => server.close((error) => (error ? reject(error) : resolve())));

  assert.equal(result.status, 1);
  assert.match(`${result.stdout}${result.stderr}`, /Checksum mismatch|native install failed/);
});

test('prints guidance when the npm-installed binary is missing', () => {
  const installDir = mkdtempSync(path.join(tmpdir(), 'warden-mcp-missing-'));
  const result = spawnSync(process.execPath, [wrapperPath, '--help'], {
    env: { ...process.env, WARDEN_MCP_INSTALL_DIR: installDir, WARDEN_MCP_VERSION: '0.1.1' },
    encoding: 'utf8',
  });

  assert.equal(result.status, 1);
  assert.match(`${result.stdout}${result.stderr}`, /npm rebuild warden-mcp/);
});

test('runs the env-overridden native executable', () => {
  const result = spawnSync(process.execPath, [wrapperPath, '--version'], {
    env: { ...process.env, WARDEN_MCP_NATIVE_PATH: process.execPath },
    encoding: 'utf8',
  });

  assert.equal(result.status, 0, result.stderr);
  assert.match(result.stdout.trim(), /^v/);
});