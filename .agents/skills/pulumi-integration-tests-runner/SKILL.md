---
name: pulumi-integration-tests-runner
description: Run pulumi-talos-cluster integration tests via make targets or direct commands using the pulumi-talos-cluster MCP server.
---

## Common Targets
- All integration tests: `make -C integration-tests integration_tests`
- Go integration tests: `make -C integration-tests integration_tests_go`
- Node.js integration tests: `make -C integration-tests integration_tests_nodejs`
- Python integration tests: `make -C integration-tests integration_tests_python`
- Scope by name: `TEST=TestHcloudClusterGo make -C integration-tests integration_tests_go`
- Unit/lint sanity: `make lint`, `make unit_tests`

## Available Tests (Grouped by Language)

**Go**
- `TestHcloudClusterGo` (Basic Hetzner Cloud cluster)
- `TestHcloudHAClusterGo` (High Availability Hetzner Cloud cluster)

**Python**
- `TestHcloudHAClusterPython` (High Availability Hetzner Cloud cluster)

**Node.js**
- `TestHcloudClusterJS` (Basic Hetzner Cloud cluster)

## Common Flow
You are given a task. After editing code, you should run the integration test. You have n (default=2) attempts. After each failed try, you should fix it and try again.

## Mandatory MCP Execution
- Always run schema generation, SDK regeneration, provider builds, and integration tests with the `shell_exec` tool of the `pulumi-talos-cluster-mcp` server (tool name in a session: `mcp__pulumi-talos-cluster-mcp__shell_exec`). The server is configured in `.mcp.json` at the repository root.
- Do not install or download `pulumi`, `pulumictl`, `talosctl`, or replacement toolchain binaries in the local workbench.
- If `shell_exec` is unavailable, stop before running a destructive generation target and report that the MCP server must be enabled.
- The tool takes a single `command` string (a shell command line; `&&`, pipes, and `VAR=value` prefixes are allowed) and runs it from the repository root on the remote runner. There are no `timeout` or `directory` parameters: the server enforces its own timeout, so use paths relative to the repository root (or absolute runner paths).
- The result is structured JSON with `stdout`, `stderr`, `exit_code`, `status`, and `execution_time`; check `exit_code`/`status`, not just stdout.
- Run the generation pipeline as separate MCP calls: `make generate_schema`, `make generate`, then `make build`.

## Usage Examples (MCP)
- Regenerate schema: `{"command": "make generate_schema"}`
- Regenerate SDKs: `{"command": "make generate"}`
- Build provider and SDKs: `{"command": "make build"}`
- Run Go integration tests: `{"command": "make -C integration-tests integration_tests_go"}`
- Run a single test: `{"command": "make -C integration-tests integration_tests_go TEST=TestHcloudClusterGo"}`
- Cancel stuck runs: stop/timeout the command and rerun.

## Notes
- Only scoped tests are allowed for integration_tests (except for talosctl).
- Prefer scoped tests (TEST=...) to stay within time limits.
- Capture stdout/stderr from MCP responses; summarize results after each run.
