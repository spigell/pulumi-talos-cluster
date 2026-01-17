---
name: talosctl-operator
description: How to work on pulumi-talos-cluster provider code using talosctl/applier patterns and available MCP tools.
---

## Overview
- Provider code: `provider/`; generated SDKs: `sdk/`; integration tests: `integration-tests/`.
- Talosctl automation lives in `provider/pkg/provider/applier/`:
  - `applier/` orchestrates init/controlplane/worker apply, upgrade, kube upgrade.
  - `applier/talosctl` wraps talosctl via `local.Command` with retries, temp TALOS_HOME (`generateWorkDirNameForTalosctl`), helpers (`RunCommand`, `RunGetCommand`, `CatFile`).
  - `applier/talosctl_gen.go` runs `talosctl gen secrets` / `gen config`.
- Cluster should generate secrets/configs via talosctl helpers, stash secrets, surface machine configs/talosconfig/client certs.
- Apply should use the applier to apply configs/bootstraps with the real talosconfig; honor `skipInitApply`; fetch kubeconfig after bootstrap.

## MCP Tools
- Server: `pulumi-talos-cluster-mcp`, tool `shell_execute`.
- Allowed commands: `go`, `make`, `ls`, `cat`, `find`, `grep`, `pulumi`, `talosctl`.
- Always set `directory` (absolute) and `timeout`; command as array, e.g. `["make","build"]`.

## Common Commands
- Build: `["make","build"]`
- Regenerate schema/SDKs: `["make","generate_schema"]`, `["make","generate"]`
- Unit tests: `["make","unit_tests"]`
- Go integration (scoped): `["make","-C","integration-tests","integration_tests_go","TEST=TestHcloudClusterGo"]` (timeout ~1800s)
- Full SDK regen/install pipeline: run `make generate_schema`, `make generate`, `make build`, then install the SDK as needed (e.g., `make install_nodejs_sdk`).

## Editing Guidelines
- Keep talosctl invocations inside applier helpers; avoid ad-hoc shell.
- Run `gofmt -w` on Go edits.
- Do not edit testdata programs when fixing provider code.

## Talosctl Patterns
- Workdir layout: `<tmp>/talos-home-for-<stack>/<step>-<machine-id>/talosctl.yaml`.
- `RunCommand` prepares talosconfig + extra files, adds retries.
- Apply flow: apply-config, bootstrap (init), upgrades; `skipInitApply` skips non-init applies for cloud-init flows.

## Outputs/Contracts
- Cluster outputs: `clientConfiguration` (CA, client cert/key, talosconfig), `machines`, `generatedConfigurations`, `secretsStash`.
- Apply outputs: `credentials` with `kubeconfig`, `talosconfig`.

## Safety
- Avoid reverting unrelated changes; no destructive git commands.
- Use MCP for builds/tests; respect timeouts.
