const { readHookInput, runWarden, sessionStartMessage, workspaceRoot } = require('./warden-hook-lib.cjs');

const input = readHookInput();
const cwd = workspaceRoot(input);
const status = runWarden('status', cwd);
const next = runWarden('next', cwd);

process.stdout.write(`${JSON.stringify(sessionStartMessage(status, next))}\n`);