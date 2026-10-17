# Changelog

## Unreleased

### Changed

- CLI stages (`cli-upgrade`, `cli-apply-config`, `cli-upgrade-k8s`) no longer set `triggers`. They update in place when their command, environment or rendered files change, instead of being replaced. A replacement re-runs the create command, which could reboot nodes.

### Added

- Generated machine configs are validated with the Talos machinery library (the same checks as `talosctl validate`) before they reach a node. An invalid config now fails the `cli-gen-machine-config` step with the Talos error and the machine ID, instead of booting a node that never starts the Talos API (`connection refused` on port 50000).

### Fixed

- The initial `apply-config` uses `--mode auto`; talosctl 1.14 removed the `reboot` mode.
- The default `talosImage` is the Image Factory installer without customizations (`factory.talos.dev/installer/376567988ad370138ad8b2698212367b8edcb69b5fd68c80be1f2ec7d603b4ba:<version>`). The previous default, `ghcr.io/siderolabs/installer:<version>`, is no longer published, so nodes could not pull the upgrade image.
- With the v1.14 contract, the install image is set in the `UnattendedInstallConfig` document instead of `machine.install.image`, which the node rejected next to that document.

### Removed

- The bash skip checks added in 0.5.0 (`talosctl_node`, the `apply-config --dry-run` match on `No changes`, and the installed-image, tag and schematic checks before `talosctl upgrade`), and the `TALOS_TARGET_IMAGE` environment variable. They parsed human-readable `talosctl` output and were only needed because the triggers forced replacements.

### Upgrade notes

- Stacks deployed with 0.4.1, 0.5.0 or 0.5.1 store `triggers`, so the first `pulumi up` replaces these commands once and re-runs them (including `talosctl upgrade`). Remove `triggers` from those resources in the stack state first. See "Talos Image Upgrades" in the README.
- With the default `talosVersionContract` (v1.14.2), kubelet node settings are generated as a `KubeNodeConfig` document. Config patches that still set legacy fields such as `machine.kubelet.nodeIP` or `machine.nodeLabels` conflict with it and are now rejected; move them to a `KubeNodeConfig` document, or pin `talosVersionContract` to the oldest Talos version your nodes run. Talos 1.14 also requires Kubernetes 1.32 or newer.

## 0.5.1 - 2026-10-16

### Fixed

- 0.5.0 npm package requested plugin version 0.5.0+dirty; use 0.5.1.
- `talosctl upgrade` now uses the machine's `talosImage`. Before, it read the install image from a stale generated configuration, so image changes never reached the nodes. The secure `apply-config` also sets `machine.install.image` to `talosImage`, and it no longer merges the config patches a second time, which duplicated list entries.
- Extension documents (`apiVersion`/`kind`) in config patches are passed to `talosctl gen config` as a separate patch, so they are kept.

### Added

- `Cluster.machineTopology` output and optional `Apply.machineTopology` input. They hold plain machine IDs grouped by type, so per-node resources can be planned before `applyMachines` is known. When `machineTopology` is omitted, it is derived from `applyMachines`. Programs written for 0.4.1 work unchanged.
- `talosctl upgrade` is skipped when the node already runs the target image (installed image, Talos tag and Image Factory schematic match). `apply-config` is skipped when the dry run reports no changes. Both the create and the update commands are guarded, so a replacement caused by a trigger change does not reboot a node that is already up to date.

### Changed

- CLI stages have an `update` command and re-run when their inputs or files change. They keep the 0.4.1 `triggers`, so existing stacks see updates rather than replacements.

### Upgrade notes

- On the first `pulumi up` after upgrading from 0.4.1, every node shows updates (no replaces) for its CLI stages. Nodes that already run their `talosImage` are not upgraded or rebooted. Nodes whose `talosImage` was changed under 0.4.1 are upgraded and rebooted once. `upgrade-k8s` runs once without rebooting nodes. See "Talos Image Upgrades" in the README.
