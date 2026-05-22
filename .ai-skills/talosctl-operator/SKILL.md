---
name: talosctl-operator
description: How to work with talosctl inside the pulumi-talos-cluster provider.
---

## Overview
- Talosctl automation lives in `provider/pkg/provider/applier/`:
  - `applier/` orchestrates init/controlplane/worker apply, upgrade, kube upgrade.
  - `applier/talosctl` wraps talosctl via `local.Command` with retries, temp TALOS_HOME (`generateWorkDirNameForTalosctl`), helpers (`RunCommand`, `RunGetCommand`).
- Cluster should generate secrets/configs via talosctl helpers, stash secrets, surface machine configs/talosconfig/client certs.

## Limitations
- Do not use local go test invocation (even with make) because your environment does not have talosctl available.

## Editing Guidelines
- Keep talosctl invocations inside applier helpers; avoid ad-hoc shell.
- Run `gofmt -w` locally on Go edits.
- Run `["make", "build_provider"]` via MCP after each edit.
- Run talosctl integration-tests via MCP after each edit.
- Do not edit testdata programs when fixing provider code.

## Talosctl Patterns
- `RunCommand` prepares talosconfig + extra files, adds retries.
- Do not use `RunGetCommand` unless explicitly requested.

## Outputs/Contracts
- Cluster outputs: `clientConfiguration` (CA, client cert/key, talosconfig), `machines`, `generatedConfigurations`, `secretsStash`.
- Apply outputs: `credentials` with `kubeconfig`, `talosconfig`.