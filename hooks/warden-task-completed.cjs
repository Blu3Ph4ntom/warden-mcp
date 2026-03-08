const { finishGateDecision, readHookInput, runWarden, workspaceRoot } = require('./warden-hook-lib.cjs');

const input = readHookInput();
const cwd = workspaceRoot(input);
const finish = runWarden('finish', cwd);
const next = finish.ok && finish.envelope?.ok && !finish.envelope.data.can_finish ? runWarden('next', cwd) : null;

process.stdout.write(`${JSON.stringify(finishGateDecision(input.hook_event_name || 'TaskCompleted', finish, next))}\n`);