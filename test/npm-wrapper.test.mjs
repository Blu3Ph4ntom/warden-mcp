import assert from 'node:assert/strict';
import { spawnSync } from 'node:child_process';
import { mkdtempSync } from 'node:fs';
import { tmpdir } from 'node:os';
import path from 'node:path';
import test from 'node:test';
import { fileURLToPath } from 'node:url';

const rootDir = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..');
const wrapperPath = path.join(rootDir, 'npm', 'warden-mcp.js');

test('prints setup guidance when the native binary is missing', () => {
  const homeDir = mkdtempSync(path.join(tmpdir(), 'warden-mcp-missing-'));
  const result = spawnSync(process.execPath, [wrapperPath, '--help'], {
    env: {
      ...process.env,
      HOME: homeDir,
      USERPROFILE: homeDir,
      GOBIN: path.join(homeDir, 'gobin'),
      GOPATH: path.join(homeDir, 'gopath'),
      WARDEN_MCP_NATIVE_PATH: path.join(homeDir, 'missing', 'warden-mcp'),
    },
    encoding: 'utf8',
  });

  const output = `${result.stdout}${result.stderr}`;
  assert.equal(result.status, 1);
  assert.match(output, /native Warden MCP binary/i);
  assert.match(output, /go install github\.com\/Blu3Ph4ntom\/warden-mcp\/cmd\/warden-mcp@latest/);
});