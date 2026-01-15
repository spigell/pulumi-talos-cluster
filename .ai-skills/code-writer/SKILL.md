---
name: code-writer
description: How to work on pulumi-talos-cluster provider code, using the existing talosctl/applier patterns and available MCP tools.
---

## Overview
- Code lives under `provider/` (Go), generated SDKs under `sdk/`, integration tests under `integration-tests/`.
- Talosctl automation is encapsulated in `provider/pkg/provider/applier/`:
  - `applier/` orchestrates init/controlplane/worker apply, upgrade, and kube upgrade.
  - `applier/talosctl` wraps talosctl via `local.Command` with retries, temp TALOS_HOME, and helpers (`RunCommand`, `RunGetCommand`, `CatFile`).
  - Generation helpers in `applier/talosctl_gen.go` run `talosctl gen secrets` / `talosctl gen config`.
- `Cluster` should generate secrets/configs via talosctl helpers, stash secrets, surface machine configs, talosconfig, and client certs.
- `Apply` should use the applier to apply configs/bootstraps using the real talosconfig; honor `skipInitApply`; fetch kubeconfig after bootstrap.

## MCP Tools
- Use `pulumi-talos-cluster-mcp` server’s `shell_execute` for commands (allowed: go, make, ls, cat, find, grep, pulumi, talosctl).
- Always set `directory` (absolute) and `timeout`; command is an array, e.g., `["make","build"]`.

## Common Commands
- Build: `["make","build"]`
- Regenerate schema/SDKs: `["make","generate_schema"]`, `["make","generate"]`
- Run unit tests: `["make","unit_tests"]`
- Run Go integration test (scoped): `["make","-C","integration-tests","integration_tests_go","TEST=TestHcloudClusterGo"]` (timeout ~1800s)

## Editing Guidelines
- Keep talosctl invocations inside applier helpers; avoid ad-hoc shell commands elsewhere.
- Use gofmt after Go edits: `gofmt -w <files>`.
- Do not touch testdata programs when fixing provider code.

## Talosctl Patterns
- Workdir helper (unexported): `generateWorkDirNameForTalosctl(stack, step, machineID)` used for CLI tasks.
- `talosctl.RunCommand` prepares `talosctl.yaml`, optional extra files, retries with bash loop.
- `Apply` uses CLI flow: apply-config, bootstrap (init only), upgrades; `skipInitApply` skips non-init applies for cloud-init flows.

## Outputs/Contracts
- `Cluster` outputs: `clientConfiguration` (CA, client cert/key, talosconfig), `machines`, `generatedConfigurations`, `secretsStash`.
- `Apply` outputs: `credentials` with `kubeconfig`, `talosconfig`.

## Safety
- Respect existing user changes; avoid reverting unrelated files.
- No destructive git commands.

## When Stuck
- Inspect logs in MCP output; adjust talosctl endpoints/ports/talosconfig as needed.
- If TLS errors occur, ensure talosconfig includes CA/client cert/key from generated secrets.***
