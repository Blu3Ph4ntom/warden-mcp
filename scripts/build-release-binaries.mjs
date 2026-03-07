import crypto from 'node:crypto';
import fs from 'node:fs';
import path from 'node:path';
import { spawnSync } from 'node:child_process';
import { fileURLToPath } from 'node:url';

const pkg = JSON.parse(fs.readFileSync(new URL('../package.json', import.meta.url), 'utf8'));
const version = process.env.WARDEN_MCP_VERSION || pkg.version;
const rootDir = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..');
const outDir = process.env.WARDEN_MCP_RELEASE_OUTDIR || path.join(rootDir, 'dist', 'release', `v${version}`);
const targetFilter = new Set((process.env.WARDEN_RELEASE_TARGETS || '').split(',').map((item) => item.trim()).filter(Boolean));
const targets = [
  ['windows', 'amd64'],
  ['windows', 'arm64'],
  ['darwin', 'amd64'],
  ['darwin', 'arm64'],
  ['linux', 'amd64'],
  ['linux', 'arm64'],
];

function assetName(goos, goarch) {
  return `warden-mcp_${version}_${goos}_${goarch}${goos === 'windows' ? '.exe' : ''}`;
}

function sha256(filePath) {
  return crypto.createHash('sha256').update(fs.readFileSync(filePath)).digest('hex');
}

fs.mkdirSync(outDir, { recursive: true });
const checksumLines = [];

for (const [goos, goarch] of targets) {
  if (targetFilter.size && !targetFilter.has(`${goos}/${goarch}`)) continue;
  const output = path.join(outDir, assetName(goos, goarch));
  const result = spawnSync('go', ['build', '-trimpath', '-ldflags=-s -w', '-o', output, './cmd/warden-mcp'], {
    cwd: rootDir,
    env: { ...process.env, CGO_ENABLED: '0', GOOS: goos, GOARCH: goarch },
    encoding: 'utf8',
  });
  if (result.status !== 0) {
    process.stderr.write(result.stdout || '');
    process.stderr.write(result.stderr || '');
    process.exit(result.status ?? 1);
  }
  checksumLines.push(`${sha256(output)}  ${path.basename(output)}`);
  process.stdout.write(`built ${path.basename(output)}\n`);
}

const checksumsPath = path.join(outDir, `warden-mcp_${version}_checksums.txt`);
fs.writeFileSync(checksumsPath, `${checksumLines.join('\n')}\n`);
process.stdout.write(`wrote ${path.relative(rootDir, checksumsPath)}\n`);

