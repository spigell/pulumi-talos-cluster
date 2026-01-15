---
name: pulumi-integration-tests-runner
description: Run pulumi-talos-cluster integration tests via make targets or direct commands using the pulumi-talos-cluster MCP server.
---

## MCP Server
- Server: `pulumi-talos-cluster-mcp`
- Tool: `shell_execute`
- Allowed commands: `wc`, `touch`, `pwd`, `find`, `go`, `grep`, `pulumi`, `ls`, `cat`, `talosctl`, `make` (keep timeouts reasonable)
- Always set `directory` (absolute path) and `timeout` seconds; pass command as an array, e.g. `["make","integration_tests_go"]`.

## Default Workdir
- `/project/workspace-pulumi/pulumi-talos-cluster`

## Common Targets
- All integration tests: `make integration_tests`
- Go integration tests: `make integration_tests_go`
- Node.js integration tests: `make integration_tests_nodejs`
- Python integration tests: `make integration_tests_python`
- Scope by name: `TEST=TestHcloud make integration_tests_go`
- Unit/lint sanity: `make lint`, `make unit_tests`

## Usage Examples (MCP)
- List repo root: `["ls"]`
- Run Go integration tests (30m timeout): `["make","integration_tests_go"]` with `timeout`: 1800
- Run a single test: `["env","TEST=TestHcloud","make","integration_tests_go"]` with `timeout`: 1200
- Cancel stuck runs: stop/timeout the command and rerun.

## Notes
- Integration tests may provision cloud resources; ensure credentials and talosctl are available.
- Prefer scoped tests (TEST=...) to stay within time limits.
- Capture stdout/stderr from MCP responses; summarize results after each run.
