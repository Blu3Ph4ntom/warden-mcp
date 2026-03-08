# Warden MCP

Warden MCP is a plan-governance MCP server for coding agents. It keeps agents on a concrete execution plan, makes progress explicit, and blocks premature "done" claims with finish-gate checks.

## Fastest setup for Claude Code, Auggie, and most coding agents

If you just want the easiest local install for an MCP-capable coding agent, use npm:

```sh
npm install -g warden-mcp
warden-mcp health
```

Then add this MCP server entry:

```json
{
  "mcpServers": {
    "warden": {
      "command": "warden-mcp",
      "args": []
    }
  }
}
```

After your client connects, call these tools in order:

1. `get_agent_guide`
2. `health_check`
3. `get_status`

If your client supports Claude Code-style local MCP config, the same `warden-mcp` entry should work there too. That includes Claude Code directly and Claude-compatible setups such as Auggie-style local MCP flows.

## Install

### Claude Code plugin marketplace (git repo)

If you prefer Claude Code's plugin flow, this repository ships a repo-root plugin and marketplace entry.

1. In Claude Code, run `/plugin marketplace add https://github.com/Blu3Ph4ntom/warden-mcp`
2. Open `/plugin`, install `warden-mcp`, and enable it for your project or session

The plugin launches Warden with `npx -y warden-mcp`, so Claude Code downloads the published npm package on first run instead of requiring a separate global install.

This plugin now ships three layers for Claude Code:

- an MCP tool server (`.mcp.json`)
- lifecycle hooks (`hooks/hooks.json`) for enforced completion gating
- convenience command files (`commands/warden-start.md`, `commands/warden-next.md`, `commands/warden-finish.md`)

The hooks are the important part: on `Stop` and `TaskCompleted`, Claude calls Warden's finish gate and gets blocked from stopping when required work remains.

For the most reliable enforced mode, make sure `warden-mcp` is also available on your `PATH` (for example via `npm install -g warden-mcp` or a released binary). The hooks first try the local `warden-mcp` binary and only fall back to `npx -y warden-mcp` if needed.

After installing or updating the plugin, restart Claude Code so the hook configuration reloads.

### Recommended for most coding-agent users: npm

This is the easiest path for Claude Code, Auggie, Cursor, Windsurf, Codex, and other local MCP clients.

```sh
npm install -g warden-mcp
```

Verify the install:

```sh
warden-mcp health
```

### Native install via Go

Prerequisite: Go 1.24+

```sh
go install github.com/Blu3Ph4ntom/warden-mcp/cmd/warden-mcp@latest
```

Then make sure your Go bin directory is on `PATH` and verify the binary is available:

```sh
warden-mcp health
```

### Native install via npm

If you prefer installing through npm, the published package downloads the correct native `warden-mcp` binary for your platform during installation.

```sh
npm install -g warden-mcp
```

or:

```sh
npx warden-mcp health
```

This is now a **real native install path**. Go is not required for npm users, but the installer does need network access to the matching GitHub Release asset for the package version.

## Quickstart for coding agents

1. Install `warden-mcp` with npm or Go.
2. Add `warden-mcp` as a local MCP server in your client config.
3. If your client can set environment variables, set `WARDEN_WORKSPACE_ROOT` to your repo root.
4. After connecting, call `get_agent_guide`, then `health_check`, then `get_status`.

If your client does not have a custom startup flow yet, this one works well:

```json
{ "command": "warden-mcp", "args": [] }
```

## What Warden MCP provides

- strict plan initialization and validation
- task updates and next-step selection
- reset, prioritization, and reconciliation flows
- finish-gate enforcement before completion is allowed
- plan import/export/archive utilities

Warden speaks MCP over **stdio**, so most local MCP clients can launch it with a simple command entry:

```json
{ "command": "warden-mcp", "args": [] }
```

When launched with no CLI args, `warden-mcp` now defaults to MCP server (`serve`) mode so local MCP clients can invoke the command directly.

## Tool surface

The current public MCP tools are:

- `init_plan`, `health_check`, `get_agent_guide`, `validate_plan`, `edit_plan`
- `get_status`, `get_next_task`, `prioritize_tasks`
- `update_task`, `reset_task`, `request_finish`
- `list_plans`, `import_plan`, `export_plan`, `archive_plan`
- `reconcile_plan`

For most clients, the first helpful call after connecting is `get_agent_guide` or `health_check`, followed by `get_status`.

## Typical workflow

1. Install `warden-mcp` and connect it as an MCP server.
2. Run `get_agent_guide` or `health_check`, then `get_status`, to confirm the active plan context.
3. Create or import a plan with `init_plan` or `import_plan`.
4. Use `get_next_task` and `update_task` as work progresses.
5. Use `validate_plan` and `reconcile_plan` after manual edits or drift.
6. Call `request_finish` only when the plan actually satisfies its finish gate.

By default, repository plans typically live at `.agent/PLAN.md`.

If `warden-mcp` is launched from an unsafe OS directory such as Windows `System32`, it will refuse to treat that location as the workspace and will fall back to `~/.warden-mcp/workspaces/default/.agent/PLAN.md` instead. If your MCP client supports environment variables, set `WARDEN_WORKSPACE_ROOT` to your project root to keep plans stored inside the repo.

If an MCP client sends a bogus absolute default plan path expanded from an unsafe OS directory, such as `C:\\Windows\\System32\\.agent\\PLAN.md`, Warden will ignore that unsafe absolute default and resolve the active workspace plan path instead.

## MCP client setup examples

### Claude Code (`.mcp.json`)

This is the main copy/paste setup for Claude Code local MCP use:

If you installed the repo plugin from the Claude marketplace flow above, you can skip this manual config and use the plugin instead.

If you want hard completion guardrails without the marketplace plugin, add the MCP server config below and also copy the repo's `hooks/` directory into your Claude project/plugin setup. MCP alone is advisory; the hooks are what enforce `request_finish` before stopping.

```json
{
  "mcpServers": {
    "warden": {
      "command": "warden-mcp",
      "args": []
    }
  }
}
```

If Claude Code lets you attach environment variables for the server, this is the safest repo-local variant:

```json
{
  "mcpServers": {
    "warden": {
      "command": "warden-mcp",
      "args": [],
      "env": {
        "WARDEN_WORKSPACE_ROOT": "${workspaceFolder}"
      }
    }
  }
}
```

### Auggie / Claude-compatible local MCP setup

If Auggie is using a Claude Code-compatible local MCP flow, use the same `warden-mcp` command entry shown above. The important part is still just:

```json
{ "command": "warden-mcp", "args": [] }
```

If Auggie exposes environment variables for MCP servers, set `WARDEN_WORKSPACE_ROOT` to the project root so Warden stores the active plan inside the repo.

In Auggie, expect Warden to show up as MCP tools/functions rather than slash commands. Use tool calls like `get_agent_guide`, `health_check`, and `get_status`.

Hard stop-interception requires lifecycle hooks similar to Claude Code's `Stop` hook support. If Auggie does not expose that kind of hook/interceptor surface, it can use Warden tools but cannot be fully forced to obey the finish gate.

### Claude Code enforced mode behavior

When the plugin hooks are active, Claude Code should:

1. inject a startup reminder that Warden enforcement is active,
2. call Warden on `TaskCompleted` and `Stop`,
3. block completion if `request_finish` says `can_finish: false`, and
4. push Claude back to the next required task.

If you want a visible bootstrap path in addition to automatic enforcement, use the plugin's Warden command files after install.

If enforcement does not seem active:

- restart Claude Code after the plugin update,
- run Claude Code with debug logging and inspect loaded hooks,
- confirm `warden-mcp health --plan .agent/PLAN.md` works in your shell,
- if not, install `warden-mcp` globally or place the released binary on your `PATH`.

### Codex CLI (`~/.codex/config.toml`)

```toml
[mcp_servers.warden]
command = "warden-mcp"
args = []
```

### Cursor (`mcp.json` server entry)

```json
{
  "warden": {
    "command": "warden-mcp",
    "args": []
  }
}
```

### Windsurf (`mcp_config.json`)

```json
{
  "mcpServers": {
    "warden": {
      "command": "warden-mcp",
      "args": []
    }
  }
}
```

### OpenCode (`opencode.json`)

```json
{
  "mcp": {
    "warden": {
      "type": "local",
      "command": ["warden-mcp"]
    }
  }
}
```

### Generic/custom MCP JSON

```json
{
  "mcpServers": {
    "warden": {
      "command": "warden-mcp",
      "args": []
    }
  }
}
```

If your client supports environment variables, add them alongside the command in that client's native format.

For basically any coding agent with local MCP support, the server contract is the same:

```json
{ "command": "warden-mcp", "args": [] }
```

Recommended first calls after install are always:

- `get_agent_guide`
- `health_check`
- `get_status`

## Release packaging

For each npm release, publish matching GitHub Release assets first:

- `warden-mcp_<version>_windows_amd64.exe`
- `warden-mcp_<version>_windows_arm64.exe`
- `warden-mcp_<version>_darwin_amd64`
- `warden-mcp_<version>_darwin_arm64`
- `warden-mcp_<version>_linux_amd64`
- `warden-mcp_<version>_linux_arm64`
- `warden-mcp_<version>_checksums.txt`

Generate them from the repo root with:

```sh
npm run build:release
```

## Troubleshooting

- `warden-mcp: command not found`: ensure your Go bin directory is on `PATH`, then verify the install with `warden-mcp health`.
- npm install failed to fetch the native binary: run `npm rebuild warden-mcp` or reinstall after publishing the matching GitHub Release assets.
- client opens the wrong workspace or stores plans outside your repo: set `WARDEN_WORKSPACE_ROOT` to the project root in your MCP server config.
- Auggie or another Claude-compatible client asks for a local MCP command: use `warden-mcp` with no args.
- client connects but the workflow seems unclear: run `get_agent_guide`, then `health_check`, then `get_status`, and make sure the repository has a valid plan file.
- manual plan edits caused drift: use `validate_plan` and `reconcile_plan` before continuing.

## Development

From the repository root:

```sh
go test ./...
npm test
npm run build:release
```

## License

MIT