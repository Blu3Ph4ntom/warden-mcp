import test from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';

const repoRoot = process.cwd();

function readJson(relativePath) {
  const filePath = path.join(repoRoot, relativePath);
  return JSON.parse(fs.readFileSync(filePath, 'utf8'));
}

test('Claude plugin manifests stay aligned with package metadata', () => {
  const pkg = readJson('package.json');
  const plugin = readJson('.claude-plugin/plugin.json');
  const marketplace = readJson('.claude-plugin/marketplace.json');
  const mcpConfig = readJson('.mcp.json');
  const readme = fs.readFileSync(path.join(repoRoot, 'README.md'), 'utf8');

  assert.equal(plugin.name, pkg.name);
  assert.equal(plugin.version, pkg.version);
  assert.equal(plugin.description, pkg.description);
  assert.equal(plugin.mcpServers, './.mcp.json');

  assert.equal(marketplace.owner.name, 'Blu3Ph4ntom');
  assert.equal(marketplace.owner.url, 'https://github.com/Blu3Ph4ntom');
  assert.equal(marketplace.metadata.version, pkg.version);
  assert.equal(marketplace.plugins.length, 1);
  assert.equal(marketplace.plugins[0].name, pkg.name);
  assert.equal(marketplace.plugins[0].version, pkg.version);
  assert.equal(marketplace.plugins[0].description, pkg.description);
  assert.equal(marketplace.plugins[0].source, './');
  assert.equal(marketplace.plugins[0].author.name, 'Blu3Ph4ntom');
  assert.equal(
    marketplace.plugins[0].repository,
    'https://github.com/Blu3Ph4ntom/warden-mcp',
  );
  assert.equal(marketplace.plugins[0].license, 'MIT');

  assert.deepEqual(mcpConfig, {
    mcpServers: {
      'warden-mcp': {
        command: 'npx',
        args: ['-y', 'warden-mcp'],
      },
    },
  });

  assert.match(
    readme,
    /\/plugin marketplace add https:\/\/github\.com\/Blu3Ph4ntom\/warden-mcp/,
  );
  assert.match(readme, /install `warden-mcp`/);
  assert.match(readme, /npx -y warden-mcp/);
  assert.match(readme, /does not currently add custom slash commands/i);
  assert.match(readme, /show up as MCP tools\/functions rather than slash commands/i);
});