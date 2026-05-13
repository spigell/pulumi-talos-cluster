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

## Usage Examples (MCP)
- Run Go integration tests (30m timeout): `["make","-C","integration-tests","integration_tests_go"]` with `timeout`: 1800
- Run a single test: `["make","-C","integration-tests","integration_tests_go","TEST=TestHcloudClusterGo"]` with `timeout`: 1200
- Cancel stuck runs: stop/timeout the command and rerun.

## Notes
- Only scoped tests are allowed for integration_tests (except for talosctl).
- Prefer scoped tests (TEST=...) to stay within time limits.
- Capture stdout/stderr from MCP responses; summarize results after each run.