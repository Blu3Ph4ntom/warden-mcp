import test from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import { createRequire } from 'node:module';

const repoRoot = process.cwd();
const require = createRequire(import.meta.url);
const { finishGateDecision, sessionStartMessage } = require(path.join(repoRoot, 'hooks', 'warden-hook-lib.cjs'));

test('finishGateDecision blocks completion and points Claude at the next task', () => {
  const finish = { ok: true, envelope: { ok: true, data: { can_finish: false, blocking_reasons: [{ message: 'phase is not complete' }], next_required_actions: ['Continue PH15-T03'] } } };
  const next = { ok: true, envelope: { ok: true, data: { next_task: { task_id: 'PH15-T03', title: 'Implement Claude hook runner scripts' } } } };
  const decision = finishGateDecision('Stop', finish, next);
  assert.equal(decision.decision, 'block');
  assert.match(decision.reason, /Warden denied completion/i);
  assert.match(decision.systemMessage, /PH15-T03/);
});

test('finishGateDecision approves completion when Warden can finish', () => {
  const finish = { ok: true, envelope: { ok: true, data: { can_finish: true, blocking_reasons: [] } } };
  const decision = finishGateDecision('Stop', finish, null);
  assert.equal(decision.decision, 'approve');
});

test('sessionStartMessage advertises enforced mode and next task', () => {
  const status = { ok: true, envelope: { data: { plan: { title: 'Warden MCP Full Platform Plan', current_phase_id: 'PH15' } } } };
  const next = { ok: true, envelope: { data: { next_task: { task_id: 'PH15-T03', title: 'Implement Claude hook runner scripts' } } } };
  const message = sessionStartMessage(status, next);
  assert.equal(message.continue, true);
  assert.equal(message.suppressOutput, true);
  assert.match(message.systemMessage, /Warden enforced mode is active/i);
  assert.match(message.systemMessage, /PH15-T03/);
});

test('Claude plugin layout includes hooks and command files for enforced mode', () => {
  const hooks = JSON.parse(fs.readFileSync(path.join(repoRoot, 'hooks', 'hooks.json'), 'utf8'));
  assert.ok(hooks.hooks.Stop?.length);
  assert.ok(hooks.hooks.TaskCompleted?.length);
  assert.ok(hooks.hooks.SessionStart?.length);
  for (const file of ['warden-start.md', 'warden-next.md', 'warden-finish.md']) {
    assert.ok(fs.existsSync(path.join(repoRoot, 'commands', file)), file);
  }
});