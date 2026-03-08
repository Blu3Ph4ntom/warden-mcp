const { readFileSync } = require('node:fs');
const { spawnSync } = require('node:child_process');

function readHookInput() {
  try {
    const raw = readFileSync(0, 'utf8').trim();
    return raw ? JSON.parse(raw) : {};
  } catch {
    return {};
  }
}

function workspaceRoot(input) {
  return process.env.CLAUDE_PROJECT_DIR || input.cwd || process.cwd();
}

function runWarden(subcommand, cwd, extraArgs = []) {
  const planPath = process.env.WARDEN_PLAN_PATH || '.agent/PLAN.md';
  const directCommand = resolveCommand(process.platform === 'win32' ? 'warden-mcp.cmd' : 'warden-mcp');
  const npxCommand = resolveCommand(process.platform === 'win32' ? 'npx.cmd' : 'npx');
  const candidates = [];
  if (process.env.WARDEN_HOOK_WARDEN_BIN) {
    candidates.push({ command: process.env.WARDEN_HOOK_WARDEN_BIN, args: [subcommand, '--plan', planPath, ...extraArgs] });
  }
  candidates.push({ command: directCommand, args: [subcommand, '--plan', planPath, ...extraArgs] });
  candidates.push({ command: npxCommand, args: ['-y', 'warden-mcp', subcommand, '--plan', planPath, ...extraArgs] });
  let lastError = 'no runner attempted';
  for (const candidate of candidates) {
    const result = runCandidate(candidate, cwd);
    if (result.error && result.error.code === 'ENOENT') {
      lastError = `${candidate.command}: not found`;
      continue;
    }
    const parsed = parseEnvelope(result.stdout);
    if (parsed) return { ok: true, envelope: parsed, runner: candidate.command };
    lastError = `${candidate.command} exited ${result.status ?? 'unknown'}${result.stderr ? `: ${result.stderr.trim()}` : ''}`;
  }
  return { ok: false, error: lastError };
}

function resolveCommand(command) {
  if (process.platform !== 'win32') return command;
  const result = spawnSync('where.exe', [command], { encoding: 'utf8', env: process.env });
  if (result.status === 0 && result.stdout.trim()) {
    return result.stdout.split(/\r?\n/).find(Boolean)?.trim() || command;
  }
  return command;
}

function runCandidate(candidate, cwd) {
  if (process.platform !== 'win32') {
    return spawnSync(candidate.command, candidate.args, { cwd, encoding: 'utf8', env: process.env });
  }
  const script = ['&', quotePowerShellArg(candidate.command), ...candidate.args.map(quotePowerShellArg)].join(' ');
  return spawnSync('powershell.exe', ['-NoProfile', '-Command', script], {
    cwd,
    encoding: 'utf8',
    env: process.env,
  });
}

function quotePowerShellArg(value) {
  return `'${String(value).replace(/'/g, "''")}'`;
}

function parseEnvelope(raw) {
  try {
    return JSON.parse((raw || '').trim());
  } catch {
    return null;
  }
}

function nextTaskSummary(nextResult, finishEnvelope) {
  const nextTask = nextResult?.ok ? nextResult.envelope?.data?.next_task : null;
  if (nextTask?.task_id && nextTask?.title) return `${nextTask.task_id} — ${nextTask.title}`;
  const nextAction = finishEnvelope?.data?.next_required_actions?.[0];
  return nextAction || 'continue the next required task from the active plan';
}

function finishGateDecision(eventName, finishResult, nextResult) {
  if (!finishResult.ok || !finishResult.envelope?.ok) {
    const detail = finishResult.ok ? finishResult.envelope?.error?.message || 'finish gate failed' : finishResult.error;
    return {
      continue: true,
      decision: 'block',
      reason: 'Warden finish gate could not be evaluated.',
      systemMessage: `Warden enforced mode blocked ${eventName} because the finish gate could not be evaluated. Resolve the Warden runtime issue first. Detail: ${detail}`,
    };
  }
  if (finishResult.envelope.data.can_finish) {
    return {
      continue: true,
      decision: 'approve',
      suppressOutput: true,
      systemMessage: 'Warden finish gate approved completion.',
    };
  }
  const blockingReason = finishResult.envelope.data.blocking_reasons?.[0]?.message || 'required work remains';
  const nextStep = nextTaskSummary(nextResult, finishResult.envelope);
  return {
    continue: true,
    decision: 'block',
    reason: `Warden denied completion: ${blockingReason}`,
    systemMessage: `Warden enforced mode denied ${eventName}. Continue with ${nextStep}. Do not stop until request_finish returns can_finish=true.`,
  };
}

function sessionStartMessage(statusResult, nextResult) {
  const planTitle = statusResult?.ok ? statusResult.envelope?.data?.plan?.title : null;
  const currentPhase = statusResult?.ok ? statusResult.envelope?.data?.plan?.current_phase_id : null;
  const nextStep = nextTaskSummary(nextResult, null);
  const detail = planTitle ? `${planTitle}${currentPhase ? ` (${currentPhase})` : ''}` : 'the active Warden plan';
  return {
    continue: true,
    suppressOutput: true,
    systemMessage: `Warden enforced mode is active for ${detail}. Use Warden status/next before major changes, and expect Stop or TaskCompleted to be blocked until finish is approved. Current next required task: ${nextStep}.`,
  };
}

module.exports = { finishGateDecision, readHookInput, runWarden, sessionStartMessage, workspaceRoot };