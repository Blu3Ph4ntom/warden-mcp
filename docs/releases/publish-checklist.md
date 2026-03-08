# Warden MCP publish checklist

Use this checklist when publishing a new Warden MCP patch release.

## 1. Confirm version surfaces

- `package.json` version matches the intended npm release version
- `internal/mcpserver/server.go` `ServerVersion` matches the same version
- release notes exist at `docs/releases/v<version>.md`

## 2. Run local validation

From the repository root:

- `go test ./...`
- `npm test`
- `npm pack --dry-run`
- `npm run build:release`

## 3. Verify release artifacts

Expected output directory:

- `dist/release/v<version>/`

Expected assets:

- `warden-mcp_<version>_windows_amd64.exe`
- `warden-mcp_<version>_windows_arm64.exe`
- `warden-mcp_<version>_darwin_amd64`
- `warden-mcp_<version>_darwin_arm64`
- `warden-mcp_<version>_linux_amd64`
- `warden-mcp_<version>_linux_arm64`
- `warden-mcp_<version>_checksums.txt`

Smoke-check at least one local binary with:

- `warden-mcp health`

## 4. Publish GitHub release assets first

- create or update the Git tag and GitHub Release for `v<version>`
- upload every file from `dist/release/v<version>/`
- confirm the checksums file matches the uploaded assets

## 5. Publish npm package second

- run `npm publish`
- verify the published package version matches the GitHub Release asset version

## 6. Post-publish smoke checks

- `npm install -g warden-mcp`
- `warden-mcp health`
- confirm MCP client setup works with `{ "command": "warden-mcp", "args": [] }`
- call `get_agent_guide`, then `health_check`, then `get_status`