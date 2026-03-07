# Warden MCP

Warden MCP is a plan-governance MCP server for coding agents. It keeps agents on a concrete execution plan, makes progress explicit, and blocks premature "done" claims with finish-gate checks.

## Install

### Native install via Go

Prerequisite: Go 1.24+

```sh
go install github.com/Blu3Ph4ntom/warden-mcp/cmd/warden-mcp@latest
```

Then make sure your Go bin directory is on `PATH` and verify the binary is available:

```sh
warden-mcp --help
```

### Native install via npm

If you prefer installing through npm, the published package downloads the correct native `warden-mcp` binary for your platform during installation.

```sh
npm install -g warden-mcp
```

or:

```sh
npx warden-mcp --help
```

This is now a **real native install path**. Go is not required for npm users, but the installer does need network access to the matching GitHub Release asset for the package version.

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

- `init_plan`, `validate_plan`, `edit_plan`
- `get_status`, `get_next_task`, `prioritize_tasks`
- `update_task`, `reset_task`, `request_finish`
- `list_plans`, `import_plan`, `export_plan`, `archive_plan`
- `reconcile_plan`, `health_check`

For most clients, the first helpful call after connecting is `health_check`, followed by `get_status`.

## Typical workflow

1. Install `warden-mcp` and connect it as an MCP server.
2. Run `health_check` and `get_status` to confirm the active plan context.
3. Create or import a plan with `init_plan` or `import_plan`.
4. Use `get_next_task` and `update_task` as work progresses.
5. Use `validate_plan` and `reconcile_plan` after manual edits or drift.
6. Call `request_finish` only when the plan actually satisfies its finish gate.

By default, repository plans typically live at `.agent/PLAN.md`.

If `warden-mcp` is launched from an unsafe OS directory such as Windows `System32`, it will refuse to treat that location as the workspace and will fall back to `~/.warden-mcp/workspaces/default/.agent/PLAN.md` instead. If your MCP client supports environment variables, set `WARDEN_WORKSPACE_ROOT` to your project root to keep plans stored inside the repo.

If an MCP client sends a bogus absolute default plan path expanded from an unsafe OS directory, such as `C:\\Windows\\System32\\.agent\\PLAN.md`, Warden will ignore that unsafe absolute default and resolve the active workspace plan path instead.

## MCP client setup examples

### Claude Code (`.mcp.json`)

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

- `warden-mcp: command not found`: ensure your Go bin directory is on `PATH`.
- npm install failed to fetch the native binary: run `npm rebuild warden-mcp` or reinstall after publishing the matching GitHub Release assets.
- client connects but the workflow seems unclear: run `health_check`, then `get_status`, and make sure the repository has a valid plan file.
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